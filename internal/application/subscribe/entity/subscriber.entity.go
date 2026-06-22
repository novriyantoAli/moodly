package entity

import (
	"time"

	"gorm.io/gorm"
)

type Subscriber struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CallName  string         `json:"call_name" gorm:"type:varchar(255);not null"`
	Username  string         `json:"username" gorm:"type:varchar(255);not null"`
	Password  string         `json:"password" gorm:"type:varchar(255);not null"`
	Plan      string         `json:"plan" gorm:"type:varchar(255);not null;check:plan IN ('pppoe','hotspot')"`
	Price     float64        `json:"price" gorm:"not null"`
	StartDate time.Time      `json:"start_date" gorm:"not null"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (u Subscriber) TableName() string {
	return "subscribers"
}
