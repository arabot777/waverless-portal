package mysql

import (
	"context"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type TaskRepo struct {
	db *gorm.DB
}

func NewTaskRepo(db *gorm.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

// Create 创建任务路由记录
func (r *TaskRepo) Create(ctx context.Context, task *model.TaskRouting) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByTaskID 根据 TaskID 获取
func (r *TaskRepo) GetByTaskID(ctx context.Context, taskID, userID string) (*model.TaskRouting, error) {
	var task model.TaskRouting
	err := r.db.WithContext(ctx).Where("task_id = ? AND user_id = ?", taskID, userID).First(&task).Error
	return &task, err
}

// Update 更新任务状态
func (r *TaskRepo) Update(ctx context.Context, taskID, status, workerID string, executionTimeMs int64) error {
	updates := map[string]interface{}{"status": status}
	if workerID != "" {
		updates["worker_id"] = workerID
	}
	if executionTimeMs > 0 {
		updates["execution_time_ms"] = executionTimeMs
	}
	if status == "COMPLETED" || status == "FAILED" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return r.db.WithContext(ctx).Model(&model.TaskRouting{}).Where("task_id = ?", taskID).Updates(updates).Error
}

// List 列出用户任务 (关联 endpoint 获取 name)
func (r *TaskRepo) List(ctx context.Context, userID, status string, limit, offset int) ([]map[string]interface{}, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&model.TaskRouting{}).Where("task_routing.user_id = ?", userID)
	if status != "" && status != "all" {
		query = query.Where("task_routing.status = ?", status)
	}
	query.Count(&total)

	var results []map[string]interface{}
	err := r.db.WithContext(ctx).Table("task_routing").
		Select("task_routing.*, user_endpoints.logical_name as endpoint").
		Joins("LEFT JOIN user_endpoints ON task_routing.endpoint_id = user_endpoints.id").
		Where("task_routing.user_id = ?", userID).
		Scopes(func(db *gorm.DB) *gorm.DB {
			if status != "" && status != "all" {
				return db.Where("task_routing.status = ?", status)
			}
			return db
		}).
		Order("task_routing.submitted_at DESC").
		Limit(limit).Offset(offset).
		Find(&results).Error

	return results, total, err
}

// GetOverview 获取任务总览统计
func (r *TaskRepo) GetOverview(ctx context.Context, userID string) (map[string]int64, error) {
	var results []struct {
		Status string
		Count  int64
	}
	err := r.db.WithContext(ctx).Model(&model.TaskRouting{}).
		Select("status, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	overview := map[string]int64{"completed": 0, "in_progress": 0, "pending": 0, "failed": 0}
	for _, r := range results {
		switch r.Status {
		case "COMPLETED":
			overview["completed"] = r.Count
		case "IN_PROGRESS":
			overview["in_progress"] = r.Count
		case "PENDING":
			overview["pending"] = r.Count
		case "FAILED":
			overview["failed"] = r.Count
		}
	}
	return overview, nil
}
