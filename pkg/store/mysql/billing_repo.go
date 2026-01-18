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

// Billing transactions
func (r *BillingRepo) CreateTransaction(ctx context.Context, tx *model.BillingTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// BillingTransactionWithEndpoint 带 endpoint 信息的计费记录
type BillingTransactionWithEndpoint struct {
	model.BillingTransaction
	EndpointName string `gorm:"column:endpoint_name"`
	SpecName     string `gorm:"column:spec_name"`
}

func (r *BillingRepo) ListTransactions(ctx context.Context, userID string, limit, offset int) ([]BillingTransactionWithEndpoint, int64, error) {
	var records []BillingTransactionWithEndpoint
	var total int64
	r.db.WithContext(ctx).Model(&model.BillingTransaction{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.WithContext(ctx).
		Table("billing_transactions bt").
		Select("bt.*, ue.logical_name as endpoint_name, ue.spec_name").
		Joins("LEFT JOIN user_endpoints ue ON bt.endpoint_id = ue.id").
		Where("bt.user_id = ?", userID).
		Order("bt.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&records).Error
	return records, total, err
}

func (r *BillingRepo) GetUsageStats(ctx context.Context, userID string, from, to time.Time) (totalAmount int64, totalSeconds int64, err error) {
	var result struct {
		TotalAmount  int64 `gorm:"column:total_amount"`
		TotalSeconds int64 `gorm:"column:total_seconds"`
	}
	err = r.db.WithContext(ctx).Model(&model.BillingTransaction{}).
		Select("COALESCE(SUM(amount), 0) as total_amount, COALESCE(SUM(duration_seconds), 0) as total_seconds").
		Where("user_id = ? AND created_at BETWEEN ? AND ? AND status = ?", userID, from, to, "success").
		Scan(&result).Error
	return result.TotalAmount, result.TotalSeconds, err
}

// Transaction support
func (r *BillingRepo) Transaction(ctx context.Context, fn func(tx *BillingRepo, userTx *UserRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&BillingRepo{db: tx}, &UserRepo{db: tx})
	})
}
