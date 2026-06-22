package entity

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ScanStatus represents the status of a scan
type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
)

// DeviceInfo contains device information for the scan
type DeviceInfo struct {
	UserAgent  string `json:"user_agent"`
	Platform   string `json:"platform"`
	Language   string `json:"language"`
	DeviceType string `json:"device_type"`
	Browser    string `json:"browser"`
}

// Scan represents a barcode scan record
type Scan struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	Barcode       string         `gorm:"not null;index;type:varchar(100)" json:"barcode"`
	Timestamp     int64          `gorm:"not null" json:"timestamp"`
	TransactionID string         `gorm:"not null;uniqueIndex;type:varchar(100)" json:"transaction_id"`
	Pin           string         `gorm:"not null;type:varchar(255)" json:"pin"`
	Photo         string         `gorm:"type:text" json:"photo"`
	DeviceInfo    datatypes.JSON `gorm:"type:jsonb" json:"device_info"`
	PhotoSize     string         `gorm:"type:varchar(20)" json:"photo_size"`
	Status        string         `gorm:"not null;default:'pending';type:varchar(20)" json:"status"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Scan
func (s *Scan) TableName() string {
	return "scans"
}

// String returns the string representation of Scan
func (s *Scan) String() string {
	return "Scan{" +
		"id=" + string(rune(s.ID)) +
		", barcode=" + s.Barcode +
		", status=" + s.Status +
		"}"
}

// IsValid validates the scan status
func (s *Scan) IsValid() bool {
	validStatuses := map[string]bool{
		string(ScanStatusPending):   true,
		string(ScanStatusCompleted): true,
		string(ScanStatusFailed):    true,
	}
	return validStatuses[s.Status]
}

// MarshalDeviceInfo converts DeviceInfo to JSON bytes
func MarshalDeviceInfo(device *DeviceInfo) ([]byte, error) {
	return json.Marshal(device)
}

// UnmarshalDeviceInfo converts JSON bytes to DeviceInfo
func UnmarshalDeviceInfo(data []byte) (*DeviceInfo, error) {
	var device DeviceInfo
	err := json.Unmarshal(data, &device)
	if err != nil {
		return nil, err
	}
	return &device, nil
}
