package mysql

import (
	"context"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetBalance(ctx context.Context, userID string) (*model.UserBalance, error) {
	var balance model.UserBalance
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&balance).Error
	return &balance, err
}

func (r *UserRepo) CreateBalance(ctx context.Context, balance *model.UserBalance) error {
	return r.db.WithContext(ctx).Create(balance).Error
}

func (r *UserRepo) UpdateBalance(ctx context.Context, userID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.UserBalance{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (r *UserRepo) AddBalance(ctx context.Context, userID string, amount float64) error {
	return r.db.WithContext(ctx).Model(&model.UserBalance{}).
		Where("user_id = ?", userID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

func (r *UserRepo) DeductBalance(ctx context.Context, userID string, amount float64) error {
	return r.db.WithContext(ctx).Model(&model.UserBalance{}).
		Where("user_id = ?", userID).
		Update("balance", gorm.Expr("balance - ?", amount)).Error
}

// Recharge records
func (r *UserRepo) CreateRechargeRecord(ctx context.Context, record *model.RechargeRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *UserRepo) GetRechargeRecord(ctx context.Context, id int64) (*model.RechargeRecord, error) {
	var record model.RechargeRecord
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, "pending").First(&record).Error
	return &record, err
}

func (r *UserRepo) UpdateRechargeRecord(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.RechargeRecord{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepo) ListRechargeRecords(ctx context.Context, userID string, limit, offset int) ([]model.RechargeRecord, int64, error) {
	var records []model.RechargeRecord
	var total int64
	query := r.db.WithContext(ctx).Model(&model.RechargeRecord{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error
	return records, total, err
}

// Transaction support
func (r *UserRepo) Transaction(ctx context.Context, fn func(tx *UserRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&UserRepo{db: tx})
	})
}
