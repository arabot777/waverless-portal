package mysql

import (
	"context"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type WorkerRepo struct {
	db *gorm.DB
}

func NewWorkerRepo(db *gorm.DB) *WorkerRepo {
	return &WorkerRepo{db: db}
}

// ListByEndpoint 获取 Endpoint 的 Workers (不包含 OFFLINE)
func (r *WorkerRepo) ListByEndpoint(ctx context.Context, endpointID int64) ([]model.Worker, error) {
	var workers []model.Worker
	err := r.db.WithContext(ctx).Where("endpoint_id = ? AND status != ?", endpointID, "OFFLINE").Order("created_at DESC").Find(&workers).Error
	return workers, err
}

// ListByUser 获取用户的所有 Workers
func (r *WorkerRepo) ListByUser(ctx context.Context, userID string, status string, limit, offset int) ([]model.Worker, int64, error) {
	var workers []model.Worker
	var total int64
	query := r.db.WithContext(ctx).Model(&model.Worker{}).Where("user_id = ?", userID)
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&workers).Error
	return workers, total, err
}

// GetByWorkerID 根据 WorkerID 获取
func (r *WorkerRepo) GetByWorkerID(ctx context.Context, workerID string) (*model.Worker, error) {
	var worker model.Worker
	err := r.db.WithContext(ctx).Where("worker_id = ?", workerID).First(&worker).Error
	return &worker, err
}

// Create 创建 Worker
func (r *WorkerRepo) Create(ctx context.Context, worker *model.Worker) error {
	return r.db.WithContext(ctx).Create(worker).Error
}

// Update 更新 Worker
func (r *WorkerRepo) Update(ctx context.Context, workerID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Worker{}).Where("worker_id = ?", workerID).Updates(updates).Error
}

// GetBillableWorkers 获取需要计费的 Workers
func (r *WorkerRepo) GetBillableWorkers(ctx context.Context) ([]model.Worker, error) {
	var workers []model.Worker
	err := r.db.WithContext(ctx).
		Where("billing_status = ? AND pod_started_at IS NOT NULL", "active").
		Find(&workers).Error
	return workers, err
}
