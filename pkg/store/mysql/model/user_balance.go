package model

import (
	"time"
)

// UserBalance 用户信息表 (余额从主站获取)
type UserBalance struct {
	UserID string `gorm:"column:user_id;primaryKey;type:varchar(100)" json:"user_id"`

	// 用户信息快照(从 JWT 同步)
	OrgID    string `gorm:"column:org_id;type:varchar(100);index:idx_org" json:"org_id"`
	UserName string `gorm:"column:user_name;type:varchar(255)" json:"user_name"`
	Email    string `gorm:"column:email;type:varchar(255);index:idx_email" json:"email"`

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (UserBalance) TableName() string {
	return "user_balances"
}
