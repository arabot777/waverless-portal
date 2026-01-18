package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/logger"
	"github.com/wavespeedai/waverless-portal/pkg/rocketmq"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"github.com/wavespeedai/waverless-portal/pkg/wavespeed"
	"gorm.io/gorm"
)

type BillingJob struct {
	db              *gorm.DB
	workerRepo      *mysql.WorkerRepo
	billingRepo     *mysql.BillingRepo
	endpointRepo    *mysql.EndpointRepo
	endpointService *service.EndpointService
	interval        time.Duration
}

func NewBillingJob(db *gorm.DB, workerRepo *mysql.WorkerRepo, billingRepo *mysql.BillingRepo, endpointRepo *mysql.EndpointRepo, endpointService *service.EndpointService) *BillingJob {
	return &BillingJob{
		db:              db,
		workerRepo:      workerRepo,
		billingRepo:     billingRepo,
		endpointRepo:    endpointRepo,
		endpointService: endpointService,
		interval:        60 * time.Second,
	}
}

func (j *BillingJob) Start(ctx context.Context) {
	logger.InfoCtx(ctx, "[BillingJob] started")
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.InfoCtx(ctx, "[BillingJob] stopped")
			return
		case <-ticker.C:
			j.run(ctx)
		}
	}
}

func (j *BillingJob) run(ctx context.Context) {
	workers, err := j.workerRepo.GetBillableWorkers(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "[BillingJob] GetBillableWorkers error: %v", err)
		return
	}

	logger.InfoCtx(ctx, "[BillingJob] found %d billable workers", len(workers))

	for _, worker := range workers {
		j.processWorker(ctx, &worker)
	}
}

func (j *BillingJob) processWorker(ctx context.Context, worker *model.Worker) {
	// 边界检查: pod_started_at 必须有值才能计费
	if worker.PodStartedAt == nil {
		logger.ErrorCtx(ctx, "[BillingJob] worker %s has no pod_started_at, skip", worker.WorkerID)
		return
	}

	// 计算扣费时间段
	deductStart := *worker.PodStartedAt
	if worker.LastBilledAt != nil {
		deductStart = *worker.LastBilledAt
	}

	now := time.Now()
	deductEnd := now
	terminated := false

	// 检查是否已终止
	if worker.PodTerminatedAt != nil {
		deductEnd = *worker.PodTerminatedAt
		terminated = true
	} else if worker.Status == "OFFLINE" {
		if worker.LastHeartbeat != nil {
			deductEnd = *worker.LastHeartbeat
		}
		terminated = true
	}

	// 边界检查: deductEnd 不能早于 deductStart
	if deductEnd.Before(deductStart) {
		logger.ErrorCtx(ctx, "[BillingJob] worker %s deductEnd %v before deductStart %v, skip", worker.WorkerID, deductEnd, deductStart)
		return
	}

	duration := int64(deductEnd.Sub(deductStart).Seconds())

	logger.InfoCtx(ctx, "[BillingJob] worker %s: status=%s, deductStart=%v, deductEnd=%v, duration=%d, terminated=%v",
		worker.WorkerID, worker.Status, deductStart, deductEnd, duration, terminated)

	// 运行中的 worker 不足 60 秒跳过，已终止的必须计费
	if !terminated && duration < 60 {
		logger.InfoCtx(ctx, "[BillingJob] worker %s duration %d < 60s, skip", worker.WorkerID, duration)
		return
	}

	if duration <= 0 {
		if terminated {
			// 终止但无需计费，直接标记完成
			j.workerRepo.Update(ctx, worker.WorkerID, map[string]interface{}{
				"billing_status": "final_billed",
			})
		}
		return
	}

	// 获取 endpoint 价格
	endpoint, err := j.endpointRepo.GetByID(ctx, worker.EndpointID)
	if err != nil {
		logger.ErrorCtx(ctx, "[BillingJob] GetEndpoint error: %v", err)
		return
	}

	// 计算费用: price_per_hour / 3600 * seconds
	cost := endpoint.PricePerHour * duration / 3600
	if cost <= 0 {
		return
	}

	// 生成幂等 key
	idempotentKey := fmt.Sprintf("portal-%s-%d", worker.WorkerID, deductStart.Unix())

	// 事务: 本地记录计费流水 + 更新 worker
	deductEndTime := deductStart.Add(time.Duration(duration) * time.Second)
	err = j.billingRepo.Transaction(ctx, func(billingTx *mysql.BillingRepo, userTx *mysql.UserRepo) error {
		// 创建流水
		tx := &model.BillingTransaction{
			UserID:             worker.UserID,
			OrgID:              endpoint.OrgID,
			EndpointID:         worker.EndpointID,
			ClusterID:          worker.ClusterID,
			WorkerID:           worker.WorkerID,
			GPUType:            endpoint.GPUType,
			GPUCount:           endpoint.GPUCount,
			BillingPeriodStart: deductStart,
			BillingPeriodEnd:   deductEndTime,
			DurationSeconds:    duration,
			PricePerHour:       endpoint.PricePerHour,
			Amount:             cost,
			Status:             "success",
		}
		if err := billingTx.CreateTransaction(ctx, tx); err != nil {
			return err
		}

		// 更新 worker
		updates := map[string]interface{}{
			"last_billed_at":       deductEndTime,
			"total_billed_seconds": gorm.Expr("total_billed_seconds + ?", duration),
			"total_billed_amount":  gorm.Expr("total_billed_amount + ?", cost),
		}
		if terminated {
			updates["billing_status"] = "final_billed"
		}
		return j.workerRepo.Update(ctx, worker.WorkerID, updates)
	})

	if err != nil {
		logger.ErrorCtx(ctx, "[BillingJob] record billing for worker %s error: %v", worker.WorkerID, err)
		return
	}

	// 发送 MQ 到主站扣款 (异步，失败不影响本地记录)
	billingMsg := &rocketmq.BillingMessage{
		UserID:      worker.UserID,
		OrgID:       endpoint.OrgID,
		RequestID:   idempotentKey,
		EndpointID:  worker.EndpointID,
		WorkerID:    worker.WorkerID,
		Amount:      cost,
		DurationSec: duration,
		Service:     "waverless-portal",
	}
	if err := rocketmq.SendBillingMessage(ctx, billingMsg); err != nil {
		logger.ErrorCtx(ctx, "[BillingJob] send MQ failed for worker %s: %v", worker.WorkerID, err)
	}

	logger.InfoCtx(ctx, "[BillingJob] worker %s billed %d for %d seconds", worker.WorkerID, cost, duration)

	// 检查余额，不足则停机
	if !terminated {
		j.checkBalanceAndStop(ctx, endpoint)
	}
}

func (j *BillingJob) checkBalanceAndStop(ctx context.Context, endpoint *model.UserEndpoint) {
	balance, err := wavespeed.GetOrgBalanceInternal(ctx, endpoint.OrgID)
	if err != nil {
		logger.ErrorCtx(ctx, "[BillingJob] get balance for org %s error: %v", endpoint.OrgID, err)
		return
	}

	// 余额不足则停机
	if balance < 0 {
		logger.InfoCtx(ctx, "[BillingJob] org %s balance %d, stopping endpoint %d",
			endpoint.OrgID, balance, endpoint.ID)
		if err := j.endpointService.ScaleEndpoint(ctx, endpoint, 0); err != nil {
			logger.ErrorCtx(ctx, "[BillingJob] stop endpoint %d error: %v", endpoint.ID, err)
		}
	}
}
