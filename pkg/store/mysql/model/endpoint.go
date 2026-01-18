package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSONMap 自定义 JSON 类型
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// UserEndpoint 用户 Endpoint 表
type UserEndpoint struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID string `gorm:"column:user_id;type:varchar(100);not null;uniqueIndex:uk_user_endpoint;index:idx_user" json:"user_id"` // 主站用户 UUID
	OrgID  string `gorm:"column:org_id;type:varchar(100);index:idx_org" json:"org_id"`                                          // 组织 ID

	// Endpoint 名称
	LogicalName  string `gorm:"column:logical_name;type:varchar(255);not null;uniqueIndex:uk_user_endpoint" json:"logical_name"` // 用户看到的名字: my-model
	PhysicalName string `gorm:"column:physical_name;type:varchar(255);not null" json:"physical_name"`                            // Waverless 中的实际名字: user-{uuid}-my-model

	// 规格信息(支持 GPU 和 CPU)
	SpecName string `gorm:"column:spec_name;type:varchar(100);not null;index:idx_spec_name" json:"spec_name"` // GPU-A100-40GB 或 CPU-16C-32G
	SpecType string `gorm:"column:spec_type;type:varchar(20);not null;index:idx_spec_type" json:"spec_type"`  // GPU 或 CPU

	// GPU 信息(仅 GPU 规格有值)
	GPUType  string `gorm:"column:gpu_type;type:varchar(100)" json:"gpu_type"` // A100-40GB(CPU 规格为 NULL)
	GPUCount int    `gorm:"column:gpu_count;default:0" json:"gpu_count"`       // GPU 数量(CPU 规格为 0)

	// 资源信息(所有规格都有)
	CPUCores int `gorm:"column:cpu_cores;not null" json:"cpu_cores"` // CPU 核心数
	RAMGB    int `gorm:"column:ram_gb;not null" json:"ram_gb"`       // 内存大小

	// 部署集群(Portal 自动选择)
	ClusterID string `gorm:"column:cluster_id;type:varchar(100);not null;index:idx_cluster" json:"cluster_id"`

	// 副本配置
	Replicas        int `gorm:"column:replicas;default:0" json:"replicas"` // 目标副本数
	MinReplicas     int `gorm:"column:min_replicas;not null" json:"min_replicas"`
	MaxReplicas     int `gorm:"column:max_replicas;not null" json:"max_replicas"`
	CurrentReplicas int `gorm:"column:current_replicas;default:0" json:"current_replicas"`

	// 镜像和配置
	Image                string  `gorm:"column:image;type:varchar(500);not null" json:"image"`
	RegistryCredentialID *int64  `gorm:"column:registry_credential_id" json:"registry_credential_id"`
	TaskTimeout          int     `gorm:"column:task_timeout;default:3600" json:"task_timeout"`
	Env                  JSONMap `gorm:"column:env;type:json" json:"env"` // 环境变量

	// 价格信息(创建时锁定, 单位: 1/1000000 USD)
	PricePerHour int64  `gorm:"column:price_per_hour;type:bigint" json:"price_per_hour"`
	Currency     string `gorm:"column:currency;type:varchar(10);default:'USD'" json:"currency"`

	// 调度偏好
	PreferRegion string `gorm:"column:prefer_region;type:varchar(100)" json:"prefer_region"` // 用户偏好区域

	// 状态
	Status string `gorm:"column:status;type:varchar(50);default:'deploying';index:idx_status" json:"status"` // deploying, running, suspended, deleted

	// 时间戳
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName 表名
func (UserEndpoint) TableName() string {
	return "user_endpoints"
}
