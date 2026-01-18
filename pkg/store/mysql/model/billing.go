package model

import (
	"time"
)

// BillingTransaction 计费流水表(Worker级别)
type BillingTransaction struct {
	ID int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	// 关联信息
	UserID     string `gorm:"column:user_id;type:varchar(100);not null;index:idx_user_billing" json:"user_id"`
	OrgID      string `gorm:"column:org_id;type:varchar(100);index:idx_org_billing" json:"org_id"`
	EndpointID int64  `gorm:"column:endpoint_id;not null;index:idx_endpoint_billing" json:"endpoint_id"`
	ClusterID  string `gorm:"column:cluster_id;type:varchar(100);not null;index:idx_cluster_billing" json:"cluster_id"`
	WorkerID   string `gorm:"column:worker_id;type:varchar(255);not null;index:idx_worker_billing" json:"worker_id"`

	// GPU 规格信息
	GPUType  string `gorm:"column:gpu_type;type:varchar(100)" json:"gpu_type"`
	GPUCount int    `gorm:"column:gpu_count" json:"gpu_count"`

	// 计费周期
	BillingPeriodStart time.Time `gorm:"column:billing_period_start;not null;index:idx_worker_billing" json:"billing_period_start"`
	BillingPeriodEnd   time.Time `gorm:"column:billing_period_end;not null" json:"billing_period_end"`
	DurationSeconds    int64     `gorm:"column:duration_seconds;not null" json:"duration_seconds"`

	// 计费信息 (单位: 1/1000000 USD, 即 1000000 = 1 USD)
	PricePerHour int64 `gorm:"column:price_per_hour;type:bigint;not null" json:"price_per_hour"` // 每小时价格
	Amount       int64 `gorm:"column:amount;type:bigint;not null" json:"amount"`                 // 本次扣费金额

	// 扣费状态
	Status       string `gorm:"column:status;type:varchar(50);default:'success'" json:"status"`
	ErrorMessage string `gorm:"column:error_message;type:text" json:"error_message"`

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_user_billing,idx_org_billing,idx_endpoint_billing,idx_cluster_billing" json:"created_at"`
}

// TableName 表名
func (BillingTransaction) TableName() string {
	return "billing_transactions"
}
