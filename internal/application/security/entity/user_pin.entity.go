package entity

import "time"

type UserPIN struct {
	UserID        uint       `gorm:"primaryKey" json:"user_id"`
	PinHash       string     `json:"pin_hash"`
	FailedAttempt int        `json:"failed_attempt"`
	RetryCount    int        `json:"retry_count"`
	LockedUntil   *time.Time `json:"locked_until"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime:milli"`
}

func (UserPIN) TableName() string {
	return "user_pins"
}
