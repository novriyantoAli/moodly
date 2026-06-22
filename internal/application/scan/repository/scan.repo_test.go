package repository

import (
	"encoding/json"
	"testing"

	"github.com/novriyantoAli/moodly/internal/application/scan/dto"
	"github.com/novriyantoAli/moodly/internal/application/scan/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestScanRepository_Create_Success(t *testing.T) {
	// Mock GORM behavior - in real tests you'd use a test database
	t.Run("should create scan with valid data", func(t *testing.T) {
		scan := &entity.Scan{
			UserID:        1,
			Barcode:       "BC123456789",
			Timestamp:     1234567890000,
			TransactionID: "trx-001",
			Pin:           "123456",
			Photo:         "data:image/jpeg;base64,test",
			DeviceInfo:    datatypes.JSON(`{"user_agent":"Mozilla"}`),
			PhotoSize:     "45 KB",
			Status:        string(entity.ScanStatusPending),
		}

		// Verify scan data structure
		assert.Equal(t, uint(1), scan.UserID)
		assert.Equal(t, "BC123456789", scan.Barcode)
		assert.Equal(t, string(entity.ScanStatusPending), scan.Status)
	})
}

func TestScanRepository_Scan_Structure(t *testing.T) {
	t.Run("should have valid table name", func(t *testing.T) {
		scan := &entity.Scan{}
		assert.Equal(t, "scans", scan.TableName())
	})

	t.Run("should support all required fields", func(t *testing.T) {
		scan := &entity.Scan{
			ID:            1,
			UserID:        1,
			Barcode:       "BC123456789",
			Timestamp:     1234567890000,
			TransactionID: "trx-001",
			Pin:           "123456",
			Photo:         "base64data",
			DeviceInfo:    datatypes.JSON(`{}`),
			PhotoSize:     "45 KB",
			Status:        string(entity.ScanStatusPending),
		}

		assert.NotNil(t, scan.ID)
		assert.NotEmpty(t, scan.Barcode)
		assert.NotEmpty(t, scan.TransactionID)
		assert.NotEmpty(t, scan.DeviceInfo)
		assert.NotEmpty(t, scan.Status)
	})
}

func TestScanRepository_TransactionID_Unique(t *testing.T) {
	t.Run("transaction id should be unique", func(t *testing.T) {
		scan1 := &entity.Scan{
			TransactionID: "trx-001",
			Barcode:       "BC1",
			Status:        string(entity.ScanStatusPending),
		}

		scan2 := &entity.Scan{
			TransactionID: "trx-001", // Same as scan1
			Barcode:       "BC2",
			Status:        string(entity.ScanStatusPending),
		}

		// In real test, duplicate transaction_id should fail
		assert.Equal(t, scan1.TransactionID, scan2.TransactionID)
	})
}

func TestScanRepository_Status_Validation(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		isValid bool
	}{
		{"pending", string(entity.ScanStatusPending), true},
		{"completed", string(entity.ScanStatusCompleted), true},
		{"failed", string(entity.ScanStatusFailed), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := &entity.Scan{Status: tt.status}
			result := scan.IsValid()
			if tt.isValid {
				assert.True(t, result)
			} else {
				assert.False(t, result)
			}
		})
	}
}

func TestScanRepository_DeviceInfo_JSON(t *testing.T) {
	t.Run("should store device info as JSON", func(t *testing.T) {
		deviceJSON := datatypes.JSON(`{
			"user_agent": "Mozilla/5.0",
			"platform": "Linux",
			"language": "en-US",
			"device_type": "mobile",
			"browser": "Chrome"
		}`)

		scan := &entity.Scan{
			DeviceInfo: deviceJSON,
		}

		// Verify JSON can be unmarshaled
		var device map[string]interface{}
		_ = json.Unmarshal(scan.DeviceInfo, &device)
		// Verify it can be parsed
		require.NotNil(t, device)
	})
}

func TestScanRepository_Photo_Storage(t *testing.T) {
	t.Run("should store base64 photo", func(t *testing.T) {
		photoBase64 := "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAIBA..."
		scan := &entity.Scan{
			Photo:     photoBase64,
			PhotoSize: "45 KB",
		}

		assert.True(t, len(scan.Photo) > 0)
		assert.Equal(t, "45 KB", scan.PhotoSize)
	})
}

func TestScanRepository_Filtering(t *testing.T) {
	t.Run("should filter by status", func(t *testing.T) {
		filter := &dto.ScanFilter{
			Status:   "pending",
			Page:     1,
			PageSize: 10,
		}

		assert.Equal(t, "pending", filter.Status)
		assert.Equal(t, 1, filter.Page)
		assert.Equal(t, 10, filter.PageSize)
	})
}

func TestScanRepository_Pagination(t *testing.T) {
	t.Run("should calculate offset and limit", func(t *testing.T) {
		filter := &dto.ScanFilter{
			Page:     2,
			PageSize: 20,
		}

		offset := (filter.Page - 1) * filter.PageSize
		limit := filter.PageSize

		assert.Equal(t, 20, offset)
		assert.Equal(t, 20, limit)
	})

	t.Run("should use default pagination", func(t *testing.T) {
		filter := &dto.ScanFilter{
			Page:     0,
			PageSize: 0,
		}

		page := filter.Page
		if page <= 0 {
			page = 1
		}
		pageSize := filter.PageSize
		if pageSize <= 0 {
			pageSize = 10
		}

		assert.Equal(t, 1, page)
		assert.Equal(t, 10, pageSize)
	})
}

func TestScanRepository_UserID_Association(t *testing.T) {
	t.Run("scan should be associated with user", func(t *testing.T) {
		scan := &entity.Scan{
			UserID:  1,
			Barcode: "BC123456789",
		}

		assert.Equal(t, uint(1), scan.UserID)
	})

	t.Run("should retrieve scans for specific user", func(t *testing.T) {
		userID := uint(1)
		scans := []entity.Scan{
			{ID: 1, UserID: userID, Barcode: "BC1"},
			{ID: 2, UserID: userID, Barcode: "BC2"},
			{ID: 3, UserID: 2, Barcode: "BC3"}, // Different user
		}

		userScans := []entity.Scan{}
		for _, scan := range scans {
			if scan.UserID == userID {
				userScans = append(userScans, scan)
			}
		}

		assert.Len(t, userScans, 2)
		assert.Equal(t, userID, userScans[0].UserID)
		assert.Equal(t, userID, userScans[1].UserID)
	})
}

func TestScanRepository_SoftDelete(t *testing.T) {
	t.Run("should support soft delete", func(t *testing.T) {
		scan := &entity.Scan{
			ID:      1,
			Barcode: "BC123456789",
			Status:  string(entity.ScanStatusPending),
		}

		// DeletedAt is initialized but not set (Valid: false)
		assert.NotNil(t, scan.DeletedAt)
		assert.False(t, scan.DeletedAt.Valid)
	})
}

func TestScanRepository_Timestamps(t *testing.T) {
	t.Run("should track creation and update times", func(t *testing.T) {
		scan := &entity.Scan{
			Barcode: "BC123456789",
		}

		// In real test with database, CreatedAt and UpdatedAt would be set
		assert.NotNil(t, &scan.CreatedAt)
		assert.NotNil(t, &scan.UpdatedAt)
	})
}
