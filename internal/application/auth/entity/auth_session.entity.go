package entity

import "time"

type AuthSession struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null;index"`
	AccessToken  string    `gorm:"size:500;not null"`
	RefreshToken string    `gorm:"size:500;not null"`
	IPAddress    string    `gorm:"size:50"`
	UserAgent    string    `gorm:"size:500"`
	ExpiredAt    time.Time `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AuthSession) TableName() string {
	return "auth_sessions"
}
