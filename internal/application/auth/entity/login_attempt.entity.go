package entity

import "time"

type LoginAttempt struct {
	ID        uint `gorm:"primaryKey"`
	UserID    *uint
	Username  string `gorm:"size:100"`
	IPAddress string `gorm:"size:50"`
	Success   bool
	Reason    string `gorm:"size:100"`
	UserAgent string `gorm:"size:500"`
	CreatedAt time.Time
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}
