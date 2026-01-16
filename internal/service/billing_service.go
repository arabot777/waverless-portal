package service

import (
	"context"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/logger"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
)

type BillingService struct {
	repo         *mysql.BillingRepo
	userRepo     *mysql.UserRepo
	endpointRepo *mysql.EndpointRepo
}

func NewBillingService(repo *mysql.BillingRepo, userRepo *mysql.UserRepo, endpointRepo *mysql.EndpointRepo) *BillingService {
	return &BillingService{repo: repo, userRepo: userRepo, endpointRepo: endpointRepo}
}

func (s *BillingService) OnWorkerCreated(ctx context.Context, workerID, clusterID, physicalName string, podStartedAt time.Time, gpuType string, gpuCount int) error {
	endpoint, err := s.endpointRepo.GetByPhysicalName(ctx, physicalName, clusterID)
	if err != nil {
		return err
	}
	return s.repo.CreateWorkerState(ctx, &model.WorkerBillingState{
		WorkerID: workerID, UserID: endpoint.UserID, OrgID: endpoint.OrgID,
		EndpointID: endpoint.ID, ClusterID: clusterID, GPUType: gpuType, GPUCount: gpuCount,
		PricePerGPUHour: endpoint.PricePerHour, PodStartedAt: podStartedAt,
		LastBilledAt: podStartedAt, BillingStatus: "active",
	})
}

func (s *BillingService) OnWorkerTerminated(ctx context.Context, workerID string, podTerminatedAt time.Time) error {
	if err := s.billWorker(ctx, workerID, &podTerminatedAt); err != nil {
		logger.ErrorCtx(ctx, "Failed to bill worker %s: %v", workerID, err)
	}
	return s.repo.UpdateWorkerState(ctx, workerID, map[string]interface{}{
		"pod_terminated_at": podTerminatedAt, "billing_status": "final_billed",
	})
}

func (s *BillingService) RunBillingJob(ctx context.Context) error {
	workers, err := s.repo.ListActiveWorkers(ctx)
	if err != nil {
		return err
	}
	for _, worker := range workers {
		if err := s.billWorker(ctx, worker.WorkerID, nil); err != nil {
			logger.ErrorCtx(ctx, "Failed to bill worker %s: %v", worker.WorkerID, err)
		}
	}
	return nil
}

func (s *BillingService) billWorker(ctx context.Context, workerID string, terminatedAt *time.Time) error {
	return s.repo.Transaction(ctx, func(billingTx *mysql.BillingRepo, userTx *mysql.UserRepo) error {
		worker, err := billingTx.GetWorkerState(ctx, workerID)
		if err != nil {
			return err
		}

		billingEnd := time.Now()
		if terminatedAt != nil {
			billingEnd = *terminatedAt
		}

		durationSeconds := int(billingEnd.Sub(worker.LastBilledAt).Seconds())
		if durationSeconds <= 0 {
			return nil
		}

		gpuHours := (float64(durationSeconds) / 3600.0) * float64(worker.GPUCount)
		amount := gpuHours * worker.PricePerGPUHour

		balance, err := userTx.GetBalance(ctx, worker.UserID)
		if err != nil {
			return err
		}

		status := "success"
		var errorMsg string
		if balance.Balance < amount {
			status = "insufficient_balance"
			errorMsg = "insufficient balance"
		} else {
			userTx.DeductBalance(ctx, worker.UserID, amount)
		}

		billingTx.CreateTransaction(ctx, &model.BillingTransaction{
			UserID: worker.UserID, OrgID: worker.OrgID, EndpointID: worker.EndpointID,
			ClusterID: worker.ClusterID, WorkerID: workerID, GPUType: worker.GPUType,
			GPUCount: worker.GPUCount, BillingPeriodStart: worker.LastBilledAt,
			BillingPeriodEnd: billingEnd, DurationSeconds: durationSeconds,
			PricePerGPUHour: worker.PricePerGPUHour, GPUHours: gpuHours, Amount: amount,
			BalanceBefore: balance.Balance + amount, BalanceAfter: balance.Balance,
			Status: status, ErrorMessage: errorMsg,
		})

		return billingTx.UpdateWorkerState(ctx, workerID, map[string]interface{}{
			"last_billed_at": billingEnd,
		})
	})
}

func (s *BillingService) GetUsageStats(ctx context.Context, userID string, from, to time.Time) (map[string]interface{}, error) {
	totalAmount, totalGPUHours, totalSeconds, err := s.repo.GetUsageStats(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_amount": totalAmount, "total_gpu_hours": totalGPUHours,
		"total_seconds": totalSeconds, "from": from, "to": to,
	}, nil
}

func (s *BillingService) GetWorkerBillingRecords(ctx context.Context, userID string, limit, offset int) ([]model.BillingTransaction, int64, error) {
	return s.repo.ListTransactions(ctx, userID, limit, offset)
}
