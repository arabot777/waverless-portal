package mysql

import (
	"context"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type BillingRepo struct {
	db *gorm.DB
}

func NewBillingRepo(db *gorm.DB) *BillingRepo {
	return &BillingRepo{db: db}
}

// Worker billing state
func (r *BillingRepo) GetWorkerState(ctx context.Context, workerID string) (*model.WorkerBillingState, error) {
	var state model.WorkerBillingState
	err := r.db.WithContext(ctx).Where("worker_id = ?", workerID).First(&state).Error
	return &state, err
}

func (r *BillingRepo) CreateWorkerState(ctx context.Context, state *model.WorkerBillingState) error {
	return r.db.WithContext(ctx).Create(state).Error
}

func (r *BillingRepo) UpdateWorkerState(ctx context.Context, workerID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.WorkerBillingState{}).Where("worker_id = ?", workerID).Updates(updates).Error
}

func (r *BillingRepo) ListActiveWorkers(ctx context.Context) ([]model.WorkerBillingState, error) {
	var workers []model.WorkerBillingState
	err := r.db.WithContext(ctx).Where("billing_status = ?", "active").Find(&workers).Error
	return workers, err
}

// Billing transactions
func (r *BillingRepo) CreateTransaction(ctx context.Context, tx *model.BillingTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *BillingRepo) ListTransactions(ctx context.Context, userID string, limit, offset int) ([]model.BillingTransaction, int64, error) {
	var records []model.BillingTransaction
	var total int64
	query := r.db.WithContext(ctx).Model(&model.BillingTransaction{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error
	return records, total, err
}

func (r *BillingRepo) GetUsageStats(ctx context.Context, userID string, from, to time.Time) (totalAmount, totalGPUHours float64, totalSeconds int, err error) {
	var result struct {
		TotalAmount   float64 `gorm:"column:total_amount"`
		TotalGPUHours float64 `gorm:"column:total_gpu_hours"`
		TotalSeconds  int     `gorm:"column:total_seconds"`
	}
	err = r.db.WithContext(ctx).Model(&model.BillingTransaction{}).
		Select("SUM(amount) as total_amount, SUM(gpu_hours) as total_gpu_hours, SUM(duration_seconds) as total_seconds").
		Where("user_id = ? AND created_at BETWEEN ? AND ? AND status = ?", userID, from, to, "success").
		Scan(&result).Error
	return result.TotalAmount, result.TotalGPUHours, result.TotalSeconds, err
}

// Task routing
func (r *BillingRepo) CreateTaskRouting(ctx context.Context, routing *model.TaskRouting) error {
	return r.db.WithContext(ctx).Create(routing).Error
}

func (r *BillingRepo) GetTaskRouting(ctx context.Context, taskID, userID string) (*model.TaskRouting, error) {
	var routing model.TaskRouting
	err := r.db.WithContext(ctx).Where("task_id = ? AND user_id = ?", taskID, userID).First(&routing).Error
	return &routing, err
}

// Transaction support
func (r *BillingRepo) Transaction(ctx context.Context, fn func(tx *BillingRepo, userTx *UserRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&BillingRepo{db: tx}, &UserRepo{db: tx})
	})
}
