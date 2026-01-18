package mysql

import (
	"context"

	"github.com/wavespeedai/waverless-portal/pkg/store/mysql/model"
	"gorm.io/gorm"
)

type SpecRepo struct {
	db *gorm.DB
}

func NewSpecRepo(db *gorm.DB) *SpecRepo {
	return &SpecRepo{db: db}
}

type SpecWithAvailability struct {
	SpecName          string `json:"spec_name"`
	SpecType          string `json:"spec_type"`
	GPUType           string `json:"gpu_type,omitempty"`
	GPUCount          int    `json:"gpu_count"`
	CPUCores          int    `json:"cpu_cores"`
	RAMGB             int    `json:"ram_gb"`
	DiskGB            int    `json:"disk_gb"`
	PricePerHour      int64  `json:"price_per_hour"`
	Currency          string `json:"currency"`
	Description       string `json:"description"`
	AvailableClusters int    `json:"available_clusters"`
	TotalCapacity     int    `json:"total_capacity"`
	AvailableCapacity int    `json:"available_capacity"`
}

func (r *SpecRepo) ListWithAvailability(ctx context.Context, specType string) ([]SpecWithAvailability, error) {
	query := `
		SELECT 
			sp.spec_name, sp.spec_type, sp.gpu_type, sp.gpu_count, sp.cpu_cores, sp.ram_gb, sp.disk_gb,
			sp.price_per_hour, sp.currency, sp.description,
			COALESCE(agg.available_clusters, 0) as available_clusters,
			COALESCE(agg.total_capacity, 0) as total_capacity,
			COALESCE(agg.available_capacity, 0) as available_capacity
		FROM spec_pricing sp
		LEFT JOIN (
			SELECT spec_name, COUNT(DISTINCT cluster_id) as available_clusters,
				SUM(total_capacity) as total_capacity, SUM(available_capacity) as available_capacity
			FROM cluster_specs WHERE is_available = true GROUP BY spec_name
		) agg ON sp.spec_name = agg.spec_name
		WHERE sp.is_available = true`

	if specType != "" {
		query += " AND sp.spec_type = ?"
	}
	query += " ORDER BY sp.spec_type, sp.price_per_hour"

	var specs []SpecWithAvailability
	var err error
	if specType != "" {
		err = r.db.WithContext(ctx).Raw(query, specType).Scan(&specs).Error
	} else {
		err = r.db.WithContext(ctx).Raw(query).Scan(&specs).Error
	}
	return specs, err
}

func (r *SpecRepo) GetByName(ctx context.Context, specName string) (*model.SpecPricing, error) {
	var spec model.SpecPricing
	err := r.db.WithContext(ctx).Where("spec_name = ?", specName).First(&spec).Error
	return &spec, err
}

func (r *SpecRepo) List(ctx context.Context) ([]model.SpecPricing, error) {
	var specs []model.SpecPricing
	err := r.db.WithContext(ctx).Order("spec_type, price_per_hour").Find(&specs).Error
	return specs, err
}

func (r *SpecRepo) Create(ctx context.Context, spec *model.SpecPricing) error {
	return r.db.WithContext(ctx).Create(spec).Error
}

func (r *SpecRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.SpecPricing{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SpecRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SpecPricing{}).Error
}
