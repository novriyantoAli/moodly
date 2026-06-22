package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;type:varchar(255);not null"`
	Password  string         `json:"password" gorm:"type:varchar(255);not null"`
	FullName  string         `json:"full_name" gorm:"type:varchar(255);not null"`
	Level     string         `json:"level" gorm:"type:varchar(20);default:'user';check:level IN ('user','agent','admin')"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (u User) TableName() string {
	return "users"
}
