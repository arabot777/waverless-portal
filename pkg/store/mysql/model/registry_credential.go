package model

import (
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/utils"
)

// RegistryCredential 镜像仓库凭证
type RegistryCredential struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            string    `gorm:"column:user_id;type:varchar(100);not null;uniqueIndex:uk_user_name" json:"user_id"`
	OrgID             string    `gorm:"column:org_id;type:varchar(100)" json:"org_id"`
	Name              string    `gorm:"column:name;type:varchar(100);not null;uniqueIndex:uk_user_name" json:"name"`
	Registry          string    `gorm:"column:registry;type:varchar(255);not null;default:docker.io" json:"registry"`
	Username          string    `gorm:"column:username;type:varchar(255);not null" json:"username"`
	PasswordEncrypted string    `gorm:"column:password_encrypted;type:varchar(1000);not null" json:"-"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (RegistryCredential) TableName() string {
	return "registry_credentials"
}

func (r *RegistryCredential) DecryptPassword() (string, error) {
	return utils.Decrypt(r.PasswordEncrypted)
}
