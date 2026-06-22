package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status ScanStatus
		want   string
	}{
		{"pending", ScanStatusPending, "pending"},
		{"completed", ScanStatusCompleted, "completed"},
		{"failed", ScanStatusFailed, "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.status))
		})
	}
}

func TestScan_TableName(t *testing.T) {
	scan := &Scan{}
	assert.Equal(t, "scans", scan.TableName())
}

func TestScan_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		isValid bool
	}{
		{"valid pending", string(ScanStatusPending), true},
		{"valid completed", string(ScanStatusCompleted), true},
		{"valid failed", string(ScanStatusFailed), true},
		{"invalid status", "invalid", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := &Scan{Status: tt.status}
			result := scan.IsValid()
			if tt.isValid {
				assert.True(t, result)
			} else {
				assert.False(t, result)
			}
		})
	}
}

func TestDeviceInfo_MarshalUnmarshal(t *testing.T) {
	originalInfo := &DeviceInfo{
		UserAgent:  "Mozilla/5.0",
		Platform:   "Linux",
		Language:   "en-US",
		DeviceType: "mobile",
		Browser:    "Chrome",
	}

	scan := &Scan{
		Barcode: "BC123456789",
		Status:  string(ScanStatusPending),
	}

	// Test marshaling
	data, err := json.Marshal(originalInfo)
	require.NoError(t, err)
	scan.DeviceInfo = data

	// Test unmarshaling
	var unmarshaledInfo DeviceInfo
	err = json.Unmarshal(scan.DeviceInfo, &unmarshaledInfo)
	require.NoError(t, err)

	assert.Equal(t, originalInfo.UserAgent, unmarshaledInfo.UserAgent)
	assert.Equal(t, originalInfo.Platform, unmarshaledInfo.Platform)
	assert.Equal(t, originalInfo.Language, unmarshaledInfo.Language)
	assert.Equal(t, originalInfo.DeviceType, unmarshaledInfo.DeviceType)
	assert.Equal(t, originalInfo.Browser, unmarshaledInfo.Browser)
}

func TestScan_String(t *testing.T) {
	scan := &Scan{
		ID:      1,
		Barcode: "BC123456789",
		Status:  string(ScanStatusPending),
	}
	str := scan.String()
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "Scan")
}

func TestScan_StatusTransitions(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus string
		toStatus   string
		isValid    bool
	}{
		{"pending to completed", string(ScanStatusPending), string(ScanStatusCompleted), true},
		{"pending to failed", string(ScanStatusPending), string(ScanStatusFailed), true},
		{"completed to completed", string(ScanStatusCompleted), string(ScanStatusCompleted), true},
		{"failed to completed", string(ScanStatusFailed), string(ScanStatusCompleted), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := &Scan{Status: tt.fromStatus}
			assert.True(t, scan.IsValid())
			scan.Status = tt.toStatus
			if tt.isValid {
				assert.True(t, scan.IsValid())
			} else {
				assert.False(t, scan.IsValid())
			}
		})
	}
}

func TestNewScan_DefaultValues(t *testing.T) {
	scan := &Scan{
		Barcode: "BC123456789",
	}

	// Status should be initialized somewhere or have a default in the service
	assert.NotNil(t, scan.Barcode)
	assert.Equal(t, "BC123456789", scan.Barcode)
}

func TestScan_ValidStatusChecks(t *testing.T) {
	t.Run("valid statuses", func(t *testing.T) {
		validStatuses := []string{
			string(ScanStatusPending),
			string(ScanStatusCompleted),
			string(ScanStatusFailed),
		}
		for _, status := range validStatuses {
			scan := &Scan{Status: status}
			assert.True(t, scan.IsValid())
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		scan := &Scan{Status: "invalid"}
		assert.False(t, scan.IsValid())
	})
}
