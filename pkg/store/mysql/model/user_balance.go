package model

import (
	"time"
)

// UserBalance 用户余额表
type UserBalance struct {
	UserID string `gorm:"column:user_id;primaryKey;type:varchar(100)" json:"user_id"`

	// 余额信息
	Balance     float64 `gorm:"column:balance;type:decimal(12,4);not null;default:0" json:"balance"`
	CreditLimit float64 `gorm:"column:credit_limit;type:decimal(12,4);default:0" json:"credit_limit"`
	Currency    string  `gorm:"column:currency;type:varchar(10);default:'USD'" json:"currency"`

	// 状态
	Status string `gorm:"column:status;type:varchar(50);default:'active'" json:"status"` // active, suspended, debt

	// 预警阈值
	LowBalanceThreshold float64 `gorm:"column:low_balance_threshold;type:decimal(12,4);default:10.00" json:"low_balance_threshold"`

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

// RechargeRecord 充值记录表
type RechargeRecord struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID string `gorm:"column:user_id;type:varchar(100);not null;index:idx_user_recharge" json:"user_id"`

	// 充值信息
	Amount   float64 `gorm:"column:amount;type:decimal(12,4);not null" json:"amount"`
	Currency string  `gorm:"column:currency;type:varchar(10);default:'USD'" json:"currency"`

	// 支付信息
	PaymentMethod string `gorm:"column:payment_method;type:varchar(50)" json:"payment_method"` // credit_card, paypal, stripe
	TransactionID string `gorm:"column:transaction_id;type:varchar(255)" json:"transaction_id"`

	// 状态
	Status string `gorm:"column:status;type:varchar(50);default:'pending';index:idx_status" json:"status"` // pending, completed, failed, refunded

	// 备注
	Note string `gorm:"column:note;type:text" json:"note"`

	// 时间戳
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime;index:idx_user_recharge" json:"created_at"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completed_at"`
}

// TableName 表名
func (RechargeRecord) TableName() string {
	return "recharge_records"
}
