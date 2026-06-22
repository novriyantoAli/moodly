package entity

import (
	"time"
)

type UserPassword struct {
	UserID        uint       `gorm:"primaryKey" json:"user_id"`
	Username      string     `gorm:"uniqueIndex" json:"username"`
	PasswordHash  string     `json:"password_hash"`
	FailedAttempt int        `json:"failed_attempt"`
	RetryCount    int        `json:"retry_count"`
	LockedUntil   *time.Time `json:"locked_until"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime:milli"`
}

func (UserPassword) TableName() string {
	return "user_passwords"
}
