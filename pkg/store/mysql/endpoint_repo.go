package mysql

import (
	"context"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type EndpointRepo struct {
	db *gorm.DB
}

func NewEndpointRepo(db *gorm.DB) *EndpointRepo {
	return &EndpointRepo{db: db}
}

func (r *EndpointRepo) GetByID(ctx context.Context, id int64) (*model.UserEndpoint, error) {
	var endpoint model.UserEndpoint
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&endpoint).Error
	return &endpoint, err
}

func (r *EndpointRepo) GetByLogicalName(ctx context.Context, userID, logicalName string) (*model.UserEndpoint, error) {
	var endpoint model.UserEndpoint
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND logical_name = ? AND deleted_at IS NULL", userID, logicalName).
		First(&endpoint).Error
	return &endpoint, err
}

func (r *EndpointRepo) GetByPhysicalName(ctx context.Context, physicalName, clusterID string) (*model.UserEndpoint, error) {
	var endpoint model.UserEndpoint
	err := r.db.WithContext(ctx).
		Where("physical_name = ? AND cluster_id = ?", physicalName, clusterID).
		First(&endpoint).Error
	return &endpoint, err
}

func (r *EndpointRepo) ListByUser(ctx context.Context, userID string) ([]model.UserEndpoint, error) {
	var endpoints []model.UserEndpoint
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&endpoints).Error
	return endpoints, err
}

func (r *EndpointRepo) ListAll(ctx context.Context) ([]model.UserEndpoint, error) {
	var endpoints []model.UserEndpoint
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND status != 'deleted'").
		Find(&endpoints).Error
	return endpoints, err
}

func (r *EndpointRepo) Create(ctx context.Context, endpoint *model.UserEndpoint) error {
	return r.db.WithContext(ctx).Create(endpoint).Error
}

func (r *EndpointRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.UserEndpoint{}).Where("id = ?", id).Updates(updates).Error
}

func (r *EndpointRepo) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.UserEndpoint{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()")).Error
}
