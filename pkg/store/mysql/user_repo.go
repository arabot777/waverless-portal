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
	var user model.UserBalance
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	return &user, err
}

func (r *UserRepo) CreateBalance(ctx context.Context, user *model.UserBalance) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) UpdateBalance(ctx context.Context, userID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.UserBalance{}).Where("user_id = ?", userID).Updates(updates).Error
}
