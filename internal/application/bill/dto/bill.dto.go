package dto

import (
	"time"

	"github.com/google/uuid"
	paymentDto "github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
)

type CreateBillRequest struct {
	SubscribeID uint      `json:"subscribe_id" binding:"required"`
	BillMonth   uint      `json:"bill_month" binding:"required,min=1,max=12"`
	BillYear    uint      `json:"bill_year" binding:"required,min=2020"`
	Amount      int64     `json:"amount" binding:"required,gt=0"`
	DueDate     time.Time `json:"due_date" binding:"required"`
	Status      string    `json:"status" binding:"omitempty,oneof=unpaid paid overdue"`
}

type UpdateBillRequest struct {
	ID      uuid.UUID `json:"id" binding:"required"`
	Amount  int64     `json:"amount" binding:"omitempty,gt=0"`
	DueDate time.Time `json:"due_date" binding:"omitempty"`
	Status  string    `json:"status" binding:"omitempty,oneof=unpaid paid overdue"`
}

type BillQuickCountOverdueResponse struct {
	Count  int64 `json:"count"`
	Amount int64 `json:"amount"`
}

type BillQuickCountUnpaidResponse struct {
	Count  int64 `json:"count"`
	Amount int64 `json:"amount"`
}

type BillResponse struct {
	ID          uuid.UUID                    `json:"id"`
	BillNumber  string                       `json:"bill_number"`
	SubscribeID uint                         `json:"subscribe_id"`
	BillMonth   uint                         `json:"bill_month"`
	BillYear    uint                         `json:"bill_year"`
	Amount      int64                        `json:"amount"`
	Status      string                       `json:"status"`
	DueDate     time.Time                    `json:"due_date"`
	PaidAt      *time.Time                   `json:"paid_at,omitempty"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
	Subscribe   dto.SubscribeResponse        `json:"subscribe"`
	Payments    []paymentDto.PaymentResponse `json:"payments,omitempty"`
}

type BillListResponse struct {
	Data       []BillResponse `json:"data"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

type CountBillResponse struct {
	Count  int64  `json:"count"`
	Status string `json:"status"`
	Month  uint   `json:"month"`
	Year   uint   `json:"year"`
}

type BillFilter struct {
	SubscribeID uint   `form:"subscribe_id"`
	Status      string `form:"status"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

type SumAmountBillFilter struct {
	Status string `form:"status"`
	Month  uint   `form:"month"`
	Year   uint   `form:"year"`
}

type CountBillFilter struct {
	Status string `form:"status"`
	Month  uint   `form:"month"`
	Year   uint   `form:"year"`
}
