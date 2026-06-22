package dto

import (
	"time"
)

type CreateSubscriberRequest struct {
	CallName  string    `json:"call_name" binding:"required"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	Plan      string    `json:"plan" binding:"required,oneof=pppoe hotspot"`
	Price     float64   `json:"price" binding:"required"`
	StartDate time.Time `json:"start_date" binding:"required"`
}

type UpdateSubscriberRequest struct {
	CallName  string    `json:"call_name"`
	Password  string    `json:"password"`
	Plan      string    `json:"plan" binding:"omitempty,oneof=pppoe hotspot"`
	Price     float64   `json:"price"`
	StartDate time.Time `json:"start_date"`
	IsActive  bool      `json:"is_active"`
}

type SubscriberResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	CallName  string    `json:"call_name"`
	Plan      string    `json:"plan"`
	Price     float64   `json:"price"`
	StartDate time.Time `json:"start_date"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriberListResponse struct {
	Data       []SubscriberResponse `json:"data"`
	TotalCount int64                `json:"total_count"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
}

type SubscribeFilter struct {
	CallName string `form:"call_name"`
	Username string `form:"username"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type CountFilter struct {
	IsActive bool   `form:"is_active"`
	Plan     string `form:"plan"`
}

// Legacy DTOs for backward compatibility
type CreateSubscribeRequest struct {
	CallName  string    `json:"callname" binding:"required"`
	Username  string    `json:"username" binding:"required"`
	Password  string    `json:"password" binding:"required"`
	Plan      string    `json:"plan" binding:"required,oneof=pppoe hotspot"`
	Price     float64   `json:"price" binding:"required"`
	StartDate time.Time `json:"startdate" binding:"required"`
}

type UpdateActive struct {
	IsActive bool `json:"is_active" binding:"required"`
}

type UpdateSubscribeRequest struct {
	CallName  string    `json:"callname"`
	Password  string    `json:"password"`
	Plan      string    `json:"plan" binding:"omitempty,oneof=pppoe hotspot"`
	Price     float64   `json:"price"`
	StartDate time.Time `json:"startdate"`
}

type SubscribeResponse struct {
	ID        uint      `json:"id"`
	Callname  string    `json:"callname"`
	Username  string    `json:"username"`
	Plan      string    `json:"plan"`
	Price     float64   `json:"price"`
	IsActive  bool      `json:"is_active"`
	StartDate time.Time `json:"startdate"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CountResponse struct {
	Plan     string `json:"plan"`
	IsActive bool   `json:"is_active"`
	Count    int64  `json:"count"`
}

type SubscribeListResponse struct {
	Data       []SubscribeResponse `json:"data"`
	TotalCount int64               `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}
