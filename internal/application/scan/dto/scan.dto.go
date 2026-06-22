package dto

import "time"

// DeviceInfo contains device information for the scan
type DeviceInfo struct {
	UserAgent  string `json:"user_agent"`
	Platform   string `json:"platform"`
	Language   string `json:"language"`
	DeviceType string `json:"device_type"`
	Browser    string `json:"browser"`
}

// CreateScanRequest is the request for creating a new scan
type CreateScanRequest struct {
	Barcode       string     `json:"barcode" binding:"required"`
	Timestamp     int64      `json:"timestamp" binding:"required"`
	TransactionID string     `json:"transaction_id" binding:"required"`
	Pin           string     `json:"pin" binding:"required"`
	Photo         string     `json:"photo" binding:"required"`
	Device        DeviceInfo `json:"device" binding:"required"`
	PhotoSize     string     `json:"photo_size" binding:"required"`
}

// UpdateScanRequest is the request for updating a scan
type UpdateScanRequest struct {
	Status string `json:"status" binding:"required,oneof=pending completed failed"`
}

// ScanResponse is the response for scan operations
type ScanResponse struct {
	ID            uint       `json:"id"`
	UserID        uint       `json:"user_id"`
	Barcode       string     `json:"barcode"`
	Timestamp     int64      `json:"timestamp"`
	TransactionID string     `json:"transaction_id"`
	Pin           string     `json:"pin"`
	Photo         string     `json:"photo"`
	Device        DeviceInfo `json:"device"`
	PhotoSize     string     `json:"photo_size"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ScanListResponse is the response for list scan operations
type ScanListResponse struct {
	Data       []ScanResponse `json:"data"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

// ScanFilter is the filter for listing scans
type ScanFilter struct {
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
