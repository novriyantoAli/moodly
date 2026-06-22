package entity

import "time"

type UserOAuth struct {
	ID             uint      `gorm:"primaryKey"`
	UserID         uint      `gorm:"not null;index"`
	Provider       string    `gorm:"size:100;not null;uniqueIndex:idx_provider_user"`
	ProviderUserID string    `gorm:"size:255;not null;uniqueIndex:idx_provider_user"`
	Email          string    `gorm:"size:100"`
	Name           string    `gorm:"size:100"`
	Picture        string    `gorm:"size:500"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime:milli"`
}

func (UserOAuth) TableName() string {
	return "user_oauth"
}
