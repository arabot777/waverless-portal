package model

import (
	"time"
)

// UserPreferences 用户偏好设置表
type UserPreferences struct {
	UserID string `gorm:"column:user_id;primaryKey;type:varchar(100)" json:"user_id"` // 主站用户 UUID

	// 预算控制
	DailyBudgetLimit   float64 `gorm:"column:daily_budget_limit;type:decimal(12,4)" json:"daily_budget_limit"`     // 每日预算上限
	MonthlyBudgetLimit float64 `gorm:"column:monthly_budget_limit;type:decimal(12,4)" json:"monthly_budget_limit"` // 每月预算上限

	// 自动化设置
	AutoSuspendOnLowBalance bool `gorm:"column:auto_suspend_on_low_balance;default:true" json:"auto_suspend_on_low_balance"` // 余额不足时自动暂停
	AutoMigrateForPrice     bool `gorm:"column:auto_migrate_for_price;default:false" json:"auto_migrate_for_price"`          // 自动迁移到更便宜的集群

	// 通知设置
	EmailNotifications bool `gorm:"column:email_notifications;default:true" json:"email_notifications"`
	LowBalanceAlert    bool `gorm:"column:low_balance_alert;default:true" json:"low_balance_alert"`

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (UserPreferences) TableName() string {
	return "user_preferences"
}
