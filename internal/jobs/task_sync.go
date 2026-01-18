package jobs

import (
	"context"
	"log"
	"time"

	"github.com/wavespeedai/waverless-portal/internal/service"
	"github.com/wavespeedai/waverless-portal/pkg/waverless"
	"gorm.io/gorm"
)

// TaskSyncJob 同步未完成任务的状态
type TaskSyncJob struct {
	db              *gorm.DB
	clusterService  *service.ClusterService
	endpointService *service.EndpointService
}

func NewTaskSyncJob(db *gorm.DB, clusterService *service.ClusterService, endpointService *service.EndpointService) *TaskSyncJob {
	return &TaskSyncJob{db: db, clusterService: clusterService, endpointService: endpointService}
}

// Start 启动 Task 同步任务 (每 10 秒执行一次)
func (j *TaskSyncJob) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.sync(ctx)
		}
	}
}

func (j *TaskSyncJob) sync(ctx context.Context) {
	// 查询未完成的任务 (PENDING, IN_PROGRESS)
	var tasks []struct {
		ID        int64
		TaskID    string
		ClusterID string
		Status    string
	}
	if err := j.db.Table("task_routing").
		Select("id, task_id, cluster_id, status").
		Where("status IN ?", []string{"PENDING", "IN_PROGRESS"}).
		Limit(100).
		Find(&tasks).Error; err != nil {
		log.Printf("[TaskSync] failed to list tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	// 按集群分组
	clusterTasks := make(map[string][]struct {
		ID        int64
		TaskID    string
		ClusterID string
		Status    string
	})
	for _, t := range tasks {
		clusterTasks[t.ClusterID] = append(clusterTasks[t.ClusterID], t)
	}

	// 遍历每个集群同步任务状态
	for clusterID, taskList := range clusterTasks {
		cluster, err := j.clusterService.GetCluster(ctx, clusterID)
		if err != nil {
			continue
		}

		client := waverless.NewClient(cluster.APIEndpoint, cluster.APIKey)
		for _, t := range taskList {
			j.syncTask(ctx, client, t.TaskID)
		}
	}
}

func (j *TaskSyncJob) syncTask(ctx context.Context, client *waverless.Client, taskID string) {
	resp, err := client.GetTaskStatus(ctx, taskID)
	if err != nil {
		return
	}

	updates := map[string]interface{}{
		"status": resp.Status,
	}
	if resp.WorkerID != "" {
		updates["worker_id"] = resp.WorkerID
	}
	if resp.ExecutionTime > 0 {
		updates["execution_time_ms"] = resp.ExecutionTime
	}
	if resp.Status == "COMPLETED" || resp.Status == "FAILED" {
		updates["completed_at"] = time.Now()
	}

	j.db.Table("task_routing").Where("task_id = ?", taskID).Updates(updates)
}
