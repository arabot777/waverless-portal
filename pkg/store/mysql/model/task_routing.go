package model

import (
	"time"
)

// TaskRouting 任务路由记录表(仅用于路由查询,不存储任务状态)
type TaskRouting struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TaskID     string `gorm:"column:task_id;type:varchar(255);unique;not null;index:idx_task_id" json:"task_id"`
	UserID     string `gorm:"column:user_id;type:varchar(100);not null;index:idx_user_task" json:"user_id"`       // 主站用户 UUID
	OrgID      string `gorm:"column:org_id;type:varchar(100);index:idx_org_task" json:"org_id"`                   // 组织 ID
	EndpointID int64  `gorm:"column:endpoint_id;not null;index:idx_endpoint_task" json:"endpoint_id"`             // Portal Endpoint ID
	ClusterID  string `gorm:"column:cluster_id;type:varchar(100);not null;index:idx_cluster_task" json:"cluster_id"` // 任务路由到的集群

	// 时间戳
	SubmittedAt time.Time `gorm:"column:submitted_at;autoCreateTime;index:idx_user_task,idx_org_task,idx_endpoint_task,idx_cluster_task" json:"submitted_at"`
}

// TableName 表名
func (TaskRouting) TableName() string {
	return "task_routing"
}
