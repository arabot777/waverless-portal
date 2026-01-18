package jobs

import (
	"context"
	"time"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/logger"
	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"github.com/wavespeedai/waverless-portal/pkg/waverless"
	"gorm.io/gorm"
)

// WorkerSyncJob 同步 Worker 信息
type WorkerSyncJob struct {
	db              *gorm.DB
	clusterService  *service.ClusterService
	endpointService *service.EndpointService
}

func NewWorkerSyncJob(db *gorm.DB, clusterService *service.ClusterService, endpointService *service.EndpointService) *WorkerSyncJob {
	return &WorkerSyncJob{db: db, clusterService: clusterService, endpointService: endpointService}
}

// Start 启动 Worker 同步任务 (每 10 秒执行一次)
func (j *WorkerSyncJob) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	j.sync(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.sync(ctx)
		}
	}
}

// SyncEndpoint 同步单个 endpoint 的 workers (供 API 调用)
func (j *WorkerSyncJob) SyncEndpoint(ctx context.Context, ep *model.UserEndpoint) {
	cluster, err := j.clusterService.GetCluster(ctx, ep.ClusterID)
	if err != nil {
		return
	}
	client := waverless.NewClient(cluster.APIEndpoint, cluster.APIKey)
	j.syncEndpointWorkers(ctx, client, ep)
}

func (j *WorkerSyncJob) sync(ctx context.Context) {
	// 获取所有活跃的 endpoints
	var endpoints []model.UserEndpoint
	if err := j.db.Where("status IN ?", []string{"running", "deploying"}).Find(&endpoints).Error; err != nil {
		logger.Infof("[WorkerSync] failed to list endpoints: %v", err)
		return
	}

	// 按集群分组
	clusterEndpoints := make(map[string][]model.UserEndpoint)
	for _, ep := range endpoints {
		clusterEndpoints[ep.ClusterID] = append(clusterEndpoints[ep.ClusterID], ep)
	}

	// 遍历每个集群同步 workers
	for clusterID, eps := range clusterEndpoints {
		cluster, err := j.clusterService.GetCluster(ctx, clusterID)
		if err != nil {
			logger.Infof("[WorkerSync] failed to get cluster %s: %v", clusterID, err)
			continue
		}

		client := waverless.NewClient(cluster.APIEndpoint, cluster.APIKey)
		for _, ep := range eps {
			j.syncEndpointWorkers(ctx, client, &ep)
		}
	}
}

func (j *WorkerSyncJob) syncEndpointWorkers(ctx context.Context, client *waverless.Client, ep *model.UserEndpoint) {
	workerList, err := client.GetEndpointWorkers(ctx, ep.PhysicalName)
	if err != nil {
		logger.Infof("[WorkerSync] failed to get workers for %s: %v", ep.PhysicalName, err)
		return
	}

	logger.Infof("[WorkerSync] endpoint %s got %d workers from waverless", ep.PhysicalName, len(workerList))

	now := time.Now()
	seenWorkerIDs := make(map[string]bool)

	for _, wm := range workerList {
		workerID := getString(wm, "id")
		if workerID == "" {
			logger.Infof("[WorkerSync] worker has no id, raw: %+v", wm)
			continue
		}
		logger.Infof("[WorkerSync] processing worker %s", workerID)
		seenWorkerIDs[workerID] = true

		// 查找或创建 worker 记录
		var worker model.Worker
		result := j.db.Where("worker_id = ?", workerID).First(&worker)

		if result.Error == gorm.ErrRecordNotFound {
			// 新 worker
			worker = model.Worker{
				WorkerID:      workerID,
				EndpointID:    ep.ID,
				ClusterID:     ep.ClusterID,
				UserID:        ep.UserID,
				PodName:       getString(wm, "pod_name"),
				Status:        getString(wm, "status"),
				BillingStatus: "pending",
				LastSyncedAt:  &now,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			// 解析时间字段 (waverless 返回驼峰命名)
			if t := parseTime(wm, "podCreatedAt"); t != nil {
				worker.PodCreatedAt = t
			}
			if t := parseTime(wm, "podStartedAt"); t != nil {
				worker.PodStartedAt = t
				worker.LastBilledAt = t
				worker.BillingStatus = "active"
			}
			if t := parseTime(wm, "podReadyAt"); t != nil {
				worker.PodReadyAt = t
			}
			if t := parseTime(wm, "last_heartbeat"); t != nil {
				worker.LastHeartbeat = t
			}
			if v, ok := wm["cold_start_duration_ms"].(float64); ok {
				ms := int64(v)
				worker.ColdStartDurationMs = &ms
			}

			if err := j.db.Create(&worker).Error; err != nil {
				logger.Infof("[WorkerSync] failed to create worker %s: %v", workerID, err)
			}
		} else if result.Error == nil {
			// 更新现有 worker
			updates := map[string]interface{}{
				"status":                getString(wm, "status"),
				"current_jobs":          getInt(wm, "current_jobs"),
				"total_tasks_completed": getInt64(wm, "total_tasks_completed"),
				"total_tasks_failed":    getInt64(wm, "total_tasks_failed"),
				"last_synced_at":        now,
				"updated_at":            now,
			}
			if t := parseTime(wm, "last_heartbeat"); t != nil {
				updates["last_heartbeat"] = t
			}
			if t := parseTime(wm, "last_task_time"); t != nil {
				updates["last_task_time"] = t
			}
			if t := parseTime(wm, "podStartedAt"); t != nil && worker.PodStartedAt == nil {
				updates["pod_started_at"] = t
				if worker.BillingStatus == "pending" {
					updates["billing_status"] = "active"
					updates["last_billed_at"] = t
				}
			}

			j.db.Model(&worker).Updates(updates)
		}
	}

	// 处理本地存在但远端没返回的 workers
	// 调用 waverless 的 /workers/:id 接口获取真实状态和终止时间
	var existingWorkers []model.Worker
	j.db.Where("endpoint_id = ? AND status NOT IN ?", ep.ID, []string{"OFFLINE"}).Find(&existingWorkers)
	for _, w := range existingWorkers {
		if !seenWorkerIDs[w.WorkerID] {
			// 查询远端 worker 详情
			remoteWorker, err := client.GetWorker(ctx, w.WorkerID)
			if err != nil {
				// 查询失败，用 last_heartbeat 作为终止时间
				terminatedAt := now
				if w.LastHeartbeat != nil {
					terminatedAt = *w.LastHeartbeat
				}
				j.db.Model(&w).Updates(map[string]interface{}{
					"status":            "OFFLINE",
					"pod_terminated_at": terminatedAt,
					"updated_at":        now,
				})
			} else {
				// 使用远端的真实状态和终止时间
				updates := map[string]interface{}{
					"status":     getString(remoteWorker, "status"),
					"updated_at": now,
				}
				if t := parseTime(remoteWorker, "terminated_at"); t != nil {
					updates["pod_terminated_at"] = t
				} else if t := parseTime(remoteWorker, "last_heartbeat"); t != nil {
					updates["pod_terminated_at"] = t
				}
				j.db.Model(&w).Updates(updates)
			}
		}
	}
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}

func parseTime(m map[string]interface{}, key string) *time.Time {
	if v, ok := m[key].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return &t
		}
		if t, err := time.Parse("2006-01-02T15:04:05Z", v); err == nil {
			return &t
		}
	}
	return nil
}
