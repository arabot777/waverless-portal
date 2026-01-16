package mysql

import (
	"context"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type ClusterRepo struct {
	db *gorm.DB
}

func NewClusterRepo(db *gorm.DB) *ClusterRepo {
	return &ClusterRepo{db: db}
}

func (r *ClusterRepo) GetByID(ctx context.Context, clusterID string) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.WithContext(ctx).Where("cluster_id = ?", clusterID).First(&cluster).Error
	return &cluster, err
}

func (r *ClusterRepo) List(ctx context.Context) ([]model.Cluster, error) {
	var clusters []model.Cluster
	err := r.db.WithContext(ctx).Find(&clusters).Error
	return clusters, err
}

func (r *ClusterRepo) ListActive(ctx context.Context) ([]model.Cluster, error) {
	var clusters []model.Cluster
	err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&clusters).Error
	return clusters, err
}

func (r *ClusterRepo) Create(ctx context.Context, cluster *model.Cluster) error {
	return r.db.WithContext(ctx).Create(cluster).Error
}

func (r *ClusterRepo) Update(ctx context.Context, clusterID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Cluster{}).Where("cluster_id = ?", clusterID).Updates(updates).Error
}

func (r *ClusterRepo) Delete(ctx context.Context, clusterID string) error {
	return r.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Delete(&model.Cluster{}).Error
}

func (r *ClusterRepo) MarkOffline(ctx context.Context, timeout time.Duration) error {
	threshold := time.Now().Add(-timeout)
	return r.db.WithContext(ctx).Model(&model.Cluster{}).
		Where("status = ? AND last_heartbeat_at < ?", "active", threshold).
		Update("status", "offline").Error
}

// ClusterSpec operations
func (r *ClusterRepo) GetSpecByClusterAndName(ctx context.Context, clusterID, specName string) (*model.ClusterSpec, error) {
	var spec model.ClusterSpec
	err := r.db.WithContext(ctx).Where("cluster_id = ? AND spec_name = ?", clusterID, specName).First(&spec).Error
	return &spec, err
}

func (r *ClusterRepo) ListSpecsByCluster(ctx context.Context, clusterID string) ([]model.ClusterSpec, error) {
	var specs []model.ClusterSpec
	err := r.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Find(&specs).Error
	return specs, err
}

func (r *ClusterRepo) ListAvailableSpecsByName(ctx context.Context, specName string) ([]model.ClusterSpec, error) {
	var specs []model.ClusterSpec
	err := r.db.WithContext(ctx).
		Where("spec_name = ? AND is_available = ? AND available_capacity > 0", specName, true).
		Find(&specs).Error
	return specs, err
}

func (r *ClusterRepo) CreateSpec(ctx context.Context, spec *model.ClusterSpec) error {
	return r.db.WithContext(ctx).Create(spec).Error
}

func (r *ClusterRepo) UpdateSpec(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.ClusterSpec{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ClusterRepo) DeleteSpec(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ClusterSpec{}).Error
}

func (r *ClusterRepo) UpdateSpecCapacity(ctx context.Context, clusterID, specName string, availableCapacity int) error {
	return r.db.WithContext(ctx).Model(&model.ClusterSpec{}).
		Where("cluster_id = ? AND spec_name = ?", clusterID, specName).
		Update("available_capacity", availableCapacity).Error
}

// Transaction support
func (r *ClusterRepo) Transaction(ctx context.Context, fn func(tx *ClusterRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&ClusterRepo{db: tx})
	})
}

// GetPricing gets spec pricing
func (r *ClusterRepo) GetPricing(ctx context.Context, specName string) (*model.SpecPricing, error) {
	var pricing model.SpecPricing
	err := r.db.WithContext(ctx).Where("spec_name = ?", specName).First(&pricing).Error
	return &pricing, err
}
