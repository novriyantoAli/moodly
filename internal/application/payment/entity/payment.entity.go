package entity

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	billEntity "github.com/novriyantoAli/moodly/internal/application/bill/entity"
	userEntity "github.com/novriyantoAli/moodly/internal/application/user/entity"
)

type Payment struct {
	ID               uint            `gorm:"primaryKey"`
	PaymentNumber    string          `gorm:"size:50;uniqueIndex;not null"`
	BillID           uuid.UUID       `gorm:"type:uuid"`
	Bill             billEntity.Bill `gorm:"foreignKey:BillID"`
	CreatedBy        *uint
	CreatedByUser    *userEntity.User `gorm:"foreignKey:CreatedBy"`
	Amount           int64            `gorm:"not null"`
	Currency         string           `gorm:"size:3;default:'IDR';not null"`
	Status           PaymentStatus    `gorm:"size:20;default:'pending'"`
	Method           PaymentMethod    `gorm:"size:30"`
	GatewayReference string           `gorm:"size:100"`
	Description      string           `gorm:"size:500"`
	PaidAt           *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

type PaymentMethod string

const (
	PaymentMethodCash           PaymentMethod = "cash"
	PaymentMethodTransfer       PaymentMethod = "transfer"
	PaymentMethodQRIS           PaymentMethod = "qris"
	PaymentMethodCreditCard     PaymentMethod = "credit_card"
	PaymentMethodVirtualAccount PaymentMethod = "virtual_account"
)

func (pm PaymentMethod) IsValid() bool {
	switch pm {
	case PaymentMethodCash, PaymentMethodTransfer, PaymentMethodQRIS, PaymentMethodCreditCard, PaymentMethodVirtualAccount:
		return true
	default:
		return false
	}
}

// type Payment struct {
// 	ID          uint                       `json:"id" gorm:"primaryKey"`
// 	Amount      float64                    `json:"amount" gorm:"not null"`
// 	Currency    string                     `json:"currency" gorm:"size:3;default:'IDR';not null"`
// 	Status      PaymentStatus              `json:"status" gorm:"type:varchar(20);default:'pending'"`
// 	Description string                     `json:"description" gorm:"size:500"`
// 	UserID      uint                       `json:"user_id" gorm:"not null"`
// 	SubscribeID uint                       `json:"subscribe_id" gorm:"not null"`
// 	CreatedAt   time.Time                  `json:"created_at"`
// 	UpdatedAt   time.Time                  `json:"updated_at"`
// 	DeletedAt   gorm.DeletedAt             `json:"deleted_at,omitempty" gorm:"index"`
// 	User        userEntity.User            `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
// 	Subscribe   subscribeEntity.Subscriber `json:"subscribe" gorm:"foreignKey:SubscribeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
// }

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCanceled   PaymentStatus = "canceled"
	PaymentStatusExpired    PaymentStatus = "expired"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)

func (p Payment) TableName() string {
	return "payments"
}

func (ps PaymentStatus) String() string {
	return string(ps)
}

func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusProcessing, PaymentStatusCompleted, PaymentStatusFailed, PaymentStatusCanceled, PaymentStatusExpired, PaymentStatusRefunded:
		return true
	default:
		return false
	}
}
