package mysql

import (
	"context"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type RegistryCredentialRepo struct {
	db *gorm.DB
}

func NewRegistryCredentialRepo(db *gorm.DB) *RegistryCredentialRepo {
	return &RegistryCredentialRepo{db: db}
}

func (r *RegistryCredentialRepo) Create(ctx context.Context, cred *model.RegistryCredential) error {
	return r.db.WithContext(ctx).Create(cred).Error
}

func (r *RegistryCredentialRepo) GetByID(ctx context.Context, id int64) (*model.RegistryCredential, error) {
	var cred model.RegistryCredential
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&cred).Error
	return &cred, err
}

func (r *RegistryCredentialRepo) GetByName(ctx context.Context, userID, name string) (*model.RegistryCredential, error) {
	var cred model.RegistryCredential
	err := r.db.WithContext(ctx).Where("user_id = ? AND name = ?", userID, name).First(&cred).Error
	return &cred, err
}

func (r *RegistryCredentialRepo) List(ctx context.Context, userID string) ([]model.RegistryCredential, error) {
	var creds []model.RegistryCredential
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&creds).Error
	return creds, err
}

func (r *RegistryCredentialRepo) Delete(ctx context.Context, userID string, name string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND name = ?", userID, name).Delete(&model.RegistryCredential{}).Error
}
