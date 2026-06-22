package dto

import (
	"time"
)

type CreatePaymentRequest struct {
	Method     string `json:"method" binding:"required,oneof=cash transfer qris credit_card virtual_account"`
	BillNumber string `json:"bill_number" binding:"required"`
}

type UpdatePaymentRequest struct {
	Status      string `json:"status" binding:"required,oneof=pending completed failed canceled"`
	Description string `json:"description"`
}

// ID               uint            `gorm:"primaryKey"`
// PaymentNumber    string          `gorm:"size:50;uniqueIndex;not null"`
// BillID           uuid.UUID       `gorm:"type:uuid"`
// Bill             billEntity.Bill `gorm:"foreignKey:BillID"`
// CreatedBy        *uint
// CreatedByUser    *userEntity.User `gorm:"foreignKey:CreatedBy"`
// Amount           int64            `gorm:"not null"`
// Currency         string           `gorm:"size:3;default:'IDR';not null"`
// Status           PaymentStatus    `gorm:"size:20;default:'pending'"`
// Method           PaymentMethod    `gorm:"size:30"`
// GatewayReference string           `gorm:"size:100"`
// Description      string           `gorm:"size:500"`
// PaidAt           *time.Time
// CreatedAt        time.Time
// UpdatedAt        time.Time
// DeletedAt        gorm.DeletedAt `gorm:"index"`
type PaymentResponse struct {
	ID               uint       `json:"id"`
	BillID           string     `json:"bill_id"`
	CreatedBy        uint       `json:"created_by"`
	PaymentNumber    string     `json:"payment_number"`
	Amount           int64      `json:"amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	Method           string     `json:"method"`
	GatewayReference string     `json:"gateway_reference"`
	Description      string     `json:"description"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type PaymentListResponse struct {
	Data       []PaymentResponse `json:"data"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

type PaymentFilter struct {
	Page          int
	PageSize      int
	Status        string
	Currency      string
	BillID        string
	PaymentNumber string
	Method        string
}
