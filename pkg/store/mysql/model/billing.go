package model

import (
	"time"
)

// BillingTransaction 计费流水表(Worker级别)
type BillingTransaction struct {
	ID int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	// 关联信息
	UserID     string `gorm:"column:user_id;type:varchar(100);not null;index:idx_user_billing" json:"user_id"`       // 主站用户 UUID
	OrgID      string `gorm:"column:org_id;type:varchar(100);index:idx_org_billing" json:"org_id"`                   // 组织 ID
	EndpointID int64  `gorm:"column:endpoint_id;not null;index:idx_endpoint_billing" json:"endpoint_id"`             // Portal Endpoint ID
	ClusterID  string `gorm:"column:cluster_id;type:varchar(100);not null;index:idx_cluster_billing" json:"cluster_id"` // 集群 ID
	WorkerID   string `gorm:"column:worker_id;type:varchar(255);not null;index:idx_worker_billing" json:"worker_id"`    // Worker ID (Pod name)

	// GPU 规格信息
	GPUType  string `gorm:"column:gpu_type;type:varchar(100);not null" json:"gpu_type"` // GPU 型号
	GPUCount int    `gorm:"column:gpu_count;not null" json:"gpu_count"`                  // GPU 数量

	// 计费周期
	BillingPeriodStart time.Time `gorm:"column:billing_period_start;not null;index:idx_worker_billing" json:"billing_period_start"` // 计费开始时间(last_billed_at)
	BillingPeriodEnd   time.Time `gorm:"column:billing_period_end;not null" json:"billing_period_end"`                              // 计费结束时间(pod_terminated_at 或 now)
	DurationSeconds    int       `gorm:"column:duration_seconds;not null" json:"duration_seconds"`                                   // 本次计费时长(秒)

	// 计费信息
	PricePerGPUHour float64 `gorm:"column:price_per_gpu_hour;type:decimal(10,4);not null" json:"price_per_gpu_hour"` // 单价(每 GPU 小时)
	GPUHours        float64 `gorm:"column:gpu_hours;type:decimal(10,4);not null" json:"gpu_hours"`                   // GPU 小时数 = (duration_seconds / 3600) * gpu_count
	Amount          float64 `gorm:"column:amount;type:decimal(12,4);not null" json:"amount"`                         // 本次扣费金额 = gpu_hours * price_per_gpu_hour
	Currency        string  `gorm:"column:currency;type:varchar(10);default:'USD'" json:"currency"`

	// 余额信息
	BalanceBefore float64 `gorm:"column:balance_before;type:decimal(12,4)" json:"balance_before"` // 扣费前余额
	BalanceAfter  float64 `gorm:"column:balance_after;type:decimal(12,4)" json:"balance_after"`   // 扣费后余额

	// 扣费状态
	Status       string `gorm:"column:status;type:varchar(50);default:'success'" json:"status"` // success, failed, insufficient_balance
	ErrorMessage string `gorm:"column:error_message;type:text" json:"error_message"`

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_user_billing,idx_org_billing,idx_endpoint_billing,idx_cluster_billing" json:"created_at"`
}

// TableName 表名
func (BillingTransaction) TableName() string {
	return "billing_transactions"
}

// WorkerBillingState Worker 计费状态表
type WorkerBillingState struct {
	WorkerID string `gorm:"column:worker_id;primaryKey;type:varchar(255)" json:"worker_id"` // Worker ID (Pod name)

	// 关联信息
	UserID     string `gorm:"column:user_id;type:varchar(100);not null;index:idx_user_worker" json:"user_id"`   // 主站用户 UUID
	OrgID      string `gorm:"column:org_id;type:varchar(100);index:idx_org_worker" json:"org_id"`               // 组织 ID
	EndpointID int64  `gorm:"column:endpoint_id;not null;index:idx_endpoint" json:"endpoint_id"`                // Portal Endpoint ID
	ClusterID  string `gorm:"column:cluster_id;type:varchar(100);not null;index:idx_cluster" json:"cluster_id"` // 集群 ID

	// GPU 规格信息(创建时从 Waverless 获取并锁定)
	GPUType         string  `gorm:"column:gpu_type;type:varchar(100);not null" json:"gpu_type"`                 // GPU 型号
	GPUCount        int     `gorm:"column:gpu_count;not null" json:"gpu_count"`                                  // GPU 数量
	PricePerGPUHour float64 `gorm:"column:price_per_gpu_hour;type:decimal(10,4);not null" json:"price_per_gpu_hour"` // 单价(创建时锁定)

	// Worker 生命周期(从 Waverless 查询)
	PodStartedAt    time.Time  `gorm:"column:pod_started_at;not null" json:"pod_started_at"`       // Pod 启动时间(首次计费起点)
	PodTerminatedAt *time.Time `gorm:"column:pod_terminated_at" json:"pod_terminated_at"`          // Pod 终止时间(NULL 表示仍在运行)

	// 计费状态
	LastBilledAt       time.Time `gorm:"column:last_billed_at;not null;index:idx_billing_status" json:"last_billed_at"`       // 上次计费时间(初始值 = pod_started_at)
	TotalBilledSeconds int       `gorm:"column:total_billed_seconds;default:0" json:"total_billed_seconds"`                    // 已计费总时长(秒)
	TotalBilledAmount  float64   `gorm:"column:total_billed_amount;type:decimal(12,4);default:0" json:"total_billed_amount"`   // 已扣费总金额
	BillingStatus      string    `gorm:"column:billing_status;type:varchar(50);default:'active';index:idx_billing_status" json:"billing_status"` // active, terminated, final_billed

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (WorkerBillingState) TableName() string {
	return "worker_billing_state"
}
