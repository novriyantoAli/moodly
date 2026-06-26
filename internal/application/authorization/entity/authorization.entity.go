package entity

import (
	"time"
)

type Role struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Role) TableName() string {
	return "roles"
}

type Permission struct {
	ID          uint      `gorm:"primaryKey"`
	Code        string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (Permission) TableName() string {
	return "permissions"
}

type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey;autoIncrement:false"`
	PermissionID uint      `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

type UserRole struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false"`
	RoleID    uint      `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
