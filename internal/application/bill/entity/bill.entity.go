package entity

import (
	"time"

	"github.com/google/uuid"

	subscribeEntity "github.com/novriyantoAli/moodly/internal/application/subscribe/entity"
)

type BillStatus string

const (
	BillStatusUnpaid  BillStatus = "unpaid"
	BillStatusPaid    BillStatus = "paid"
	BillStatusOverdue BillStatus = "overdue"
	BillStatusVoid    BillStatus = "void"
)

type Bill struct {
	ID          uuid.UUID                  `gorm:"type:uuid;primaryKey"`
	BillNumber  string                     `gorm:"size:50;uniqueIndex;not null"`
	SubscribeID uint                       `gorm:"not null;uniqueIndex:idx_bill_unique"`
	BillMonth   uint                       `gorm:"not null;uniqueIndex:idx_bill_unique"`
	BillYear    uint                       `gorm:"not null;uniqueIndex:idx_bill_unique"`
	Amount      int64                      `gorm:"not null"`
	Status      BillStatus                 `gorm:"type:varchar(20);not null;default:'unpaid';index:idx_status_due,priority:1"`
	Description string                     `gorm:"size:255"`
	DueDate     time.Time                  `gorm:"not null;index:idx_status_due,priority:2"`
	PaidAt      *time.Time                 `gorm:"index"`
	GeneratedAt time.Time                  `gorm:"not null"`
	CreatedAt   time.Time                  `gorm:"autoCreateTime:milli"`
	UpdatedAt   time.Time                  `gorm:"autoUpdateTime:milli"`
	Subscribe   subscribeEntity.Subscriber `gorm:"foreignKey:SubscribeID"`
}

func (e Bill) TableName() string {
	return "bills"
}
