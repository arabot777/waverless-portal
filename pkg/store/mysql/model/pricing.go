package model

import (
	"time"
)

// SpecPricing 规格价格配置表(支持 GPU 和 CPU 规格)
type SpecPricing struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SpecName string `gorm:"column:spec_name;type:varchar(100);not null;unique" json:"spec_name"` // GPU-A100-40GB, GPU-H100-80GB, CPU-16C-32G

	// 规格类型
	SpecType string `gorm:"column:spec_type;type:varchar(20);not null;index:idx_spec_type" json:"spec_type"` // GPU, CPU

	// 资源配置(用于展示和说明)
	GPUType  string `gorm:"column:gpu_type;type:varchar(100);index:idx_gpu_type" json:"gpu_type"` // A100-40GB, H100-80GB (仅 GPU 规格)
	GPUCount int    `gorm:"column:gpu_count;default:0" json:"gpu_count"`                           // GPU 数量
	CPUCores int    `gorm:"column:cpu_cores;not null" json:"cpu_cores"`                            // CPU 核心数
	RAMGB    int    `gorm:"column:ram_gb;not null" json:"ram_gb"`                                  // 内存大小
	DiskGB   int    `gorm:"column:disk_gb" json:"disk_gb"`                                         // 磁盘大小

	// 价格 (单位: 1/1000000 USD)
	PricePerHour int64  `gorm:"column:price_per_hour;type:bigint;not null" json:"price_per_hour"`
	Currency     string `gorm:"column:currency;type:varchar(10);default:'USD'" json:"currency"`

	// 价格范围(可选)
	MinPrice int64 `gorm:"column:min_price;type:bigint" json:"min_price"`
	MaxPrice int64 `gorm:"column:max_price;type:bigint" json:"max_price"`

	// 描述
	Description string `gorm:"column:description;type:text" json:"description"`

	// 可用性
	IsAvailable bool `gorm:"column:is_available;default:true;index:idx_available" json:"is_available"`

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (SpecPricing) TableName() string {
	return "spec_pricing"
}

// ClusterPricingOverride 集群价格覆盖表(特殊定价)
type ClusterPricingOverride struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClusterID string `gorm:"column:cluster_id;type:varchar(100);not null;uniqueIndex:uk_cluster_spec" json:"cluster_id"`
	SpecName  string `gorm:"column:spec_name;type:varchar(100);not null;uniqueIndex:uk_cluster_spec" json:"spec_name"` // 改为 spec_name,支持 GPU 和 CPU

	// 覆盖价格 (单位: 1/1000000 USD)
	PricePerHour int64  `gorm:"column:price_per_hour;type:bigint;not null" json:"price_per_hour"`
	Currency     string `gorm:"column:currency;type:varchar(10);default:'USD'" json:"currency"`

	// 生效时间
	EffectiveFrom  *time.Time `gorm:"column:effective_from;index:idx_effective" json:"effective_from"`
	EffectiveUntil *time.Time `gorm:"column:effective_until;index:idx_effective" json:"effective_until"`

	// 原因
	Reason string `gorm:"column:reason;type:varchar(255)" json:"reason"`

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (ClusterPricingOverride) TableName() string {
	return "cluster_pricing_overrides"
}
