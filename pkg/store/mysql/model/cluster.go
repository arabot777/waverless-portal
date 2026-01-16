package model

import (
	"time"
)

// Cluster 集群注册表
type Cluster struct {
	ClusterID   string `gorm:"column:cluster_id;primaryKey;type:varchar(100)" json:"cluster_id"`
	ClusterName string `gorm:"column:cluster_name;type:varchar(255);not null" json:"cluster_name"`
	Region      string `gorm:"column:region;type:varchar(100);not null;index:idx_region" json:"region"`

	// 连接信息
	APIEndpoint string `gorm:"column:api_endpoint;type:varchar(500);not null" json:"api_endpoint"`
	APIKey      string `gorm:"column:api_key_hash;type:varchar(255)" json:"api_key"`

	// 集群状态
	Status   string `gorm:"column:status;type:varchar(50);default:'active';index:idx_status" json:"status"`
	Priority int    `gorm:"column:priority;default:100" json:"priority"`

	// 容量统计
	TotalGPUSlots     int `gorm:"column:total_gpu_slots;default:0" json:"total_gpu_slots"`
	AvailableGPUSlots int `gorm:"column:available_gpu_slots;default:0" json:"available_gpu_slots"`

	LastHeartbeatAt *time.Time `gorm:"column:last_heartbeat_at;index:idx_heartbeat" json:"last_heartbeat_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (Cluster) TableName() string {
	return "clusters"
}

// ClusterSpec 集群规格表 - 记录集群支持哪些规格及容量
type ClusterSpec struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClusterID       string `gorm:"column:cluster_id;type:varchar(100);not null;uniqueIndex:uk_cluster_spec" json:"cluster_id"`
	ClusterSpecName string `gorm:"column:cluster_spec_name;type:varchar(100);not null;uniqueIndex:uk_cluster_spec" json:"cluster_spec_name"` // 集群内部的 spec 名称
	SpecName        string `gorm:"column:spec_name;type:varchar(100);not null;index:idx_spec_name" json:"spec_name"`                         // 关联 spec_pricing 的名称

	// 容量信息
	TotalCapacity     int `gorm:"column:total_capacity;default:0" json:"total_capacity"`
	AvailableCapacity int `gorm:"column:available_capacity;default:0" json:"available_capacity"`

	// 可用性
	IsAvailable bool `gorm:"column:is_available;default:true;index:idx_available" json:"is_available"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (ClusterSpec) TableName() string {
	return "cluster_specs"
}
