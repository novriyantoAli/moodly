package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/application/scan/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockScanService struct {
	mock.Mock
}

func (m *MockScanService) CreateScan(ctx context.Context, userID uint, req *dto.CreateScanRequest) (*dto.ScanResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ScanResponse), args.Error(1)
}

func (m *MockScanService) GetScanByID(ctx context.Context, id uint) (*dto.ScanResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ScanResponse), args.Error(1)
}

func (m *MockScanService) GetScans(ctx context.Context, filter *dto.ScanFilter) (*dto.ScanListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ScanListResponse), args.Error(1)
}

func (m *MockScanService) GetUserScans(ctx context.Context, userID uint, filter *dto.ScanFilter) (*dto.ScanListResponse, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ScanListResponse), args.Error(1)
}

func (m *MockScanService) UpdateScanStatus(ctx context.Context, id uint, req *dto.UpdateScanRequest) (*dto.ScanResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ScanResponse), args.Error(1)
}

func (m *MockScanService) DeleteScan(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestScanHandler_CreateScan_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	mockService.On("CreateScan", context.Background(), uint(1), mock.AnythingOfType("*dto.CreateScanRequest")).
		Return(&dto.ScanResponse{
			ID:            1,
			Barcode:       "BC123456789",
			Status:        "pending",
			TransactionID: "trx-001",
		}, nil)

	handler := NewScanHandler(mockService, logger)

	req := &dto.CreateScanRequest{
		Barcode:       "BC123456789",
		Timestamp:     1234567890000,
		TransactionID: "trx-001",
		Pin:           "123456",
		Photo:         "data:image/jpeg;base64,test",
		Device: dto.DeviceInfo{
			UserAgent:  "Mozilla/5.0",
			Platform:   "Linux",
			Language:   "en-US",
			DeviceType: "mobile",
			Browser:    "Chrome",
		},
		PhotoSize: "45 KB",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/scans", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Set("user_id", uint(1))

	handler.CreateScan(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestScanHandler_GetScan_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	mockService.On("GetScanByID", context.Background(), uint(1)).
		Return(&dto.ScanResponse{
			ID:      1,
			Barcode: "BC123456789",
			Status:  "pending",
		}, nil)

	handler := NewScanHandler(mockService, logger)

	httpReq := httptest.NewRequest("GET", "/api/v1/scans/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetScan(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanHandler_GetScan_NotFound(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	mockService.On("GetScanByID", context.Background(), uint(999)).Return(nil, assert.AnError)

	handler := NewScanHandler(mockService, logger)

	httpReq := httptest.NewRequest("GET", "/api/v1/scans/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler.GetScan(c)

	assert.True(t, w.Code >= 400)
}

func TestScanHandler_GetScans_WithFilters(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	mockService.On("GetScans", context.Background(), mock.AnythingOfType("*dto.ScanFilter")).
		Return(&dto.ScanListResponse{
			Data: []dto.ScanResponse{
				{ID: 1, Barcode: "BC1", Status: "pending"},
				{ID: 2, Barcode: "BC2", Status: "completed"},
			},
			TotalCount: 2,
			Page:       1,
			PageSize:   10,
		}, nil)

	handler := NewScanHandler(mockService, logger)

	httpReq := httptest.NewRequest("GET", "/api/v1/scans?page=1&page_size=10&status=pending", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	handler.GetScans(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanHandler_GetUserScans_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	mockService.On("GetUserScans", context.Background(), uint(1), mock.AnythingOfType("*dto.ScanFilter")).
		Return(&dto.ScanListResponse{
			Data: []dto.ScanResponse{
				{ID: 1, UserID: 1, Barcode: "BC1", Status: "pending"},
			},
			TotalCount: 1,
			Page:       1,
			PageSize:   10,
		}, nil)

	handler := NewScanHandler(mockService, logger)

	httpReq := httptest.NewRequest("GET", "/api/v1/users/1/scans?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{{Key: "user_id", Value: "1"}}

	handler.GetUserScans(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanHandler_UpdateScanStatus_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	mockService.On("UpdateScanStatus", context.Background(), uint(1), mock.AnythingOfType("*dto.UpdateScanRequest")).
		Return(&dto.ScanResponse{
			ID:     1,
			Status: "completed",
		}, nil)

	handler := NewScanHandler(mockService, logger)

	req := &dto.UpdateScanRequest{Status: "completed"}
	body, _ := json.Marshal(req)

	httpReq := httptest.NewRequest("PUT", "/api/v1/scans/1", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.UpdateScanStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanHandler_DeleteScan_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	mockService.On("DeleteScan", context.Background(), uint(1)).Return(nil)

	handler := NewScanHandler(mockService, logger)

	httpReq := httptest.NewRequest("DELETE", "/api/v1/scans/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.DeleteScan(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanHandler_CreateScan_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	handler := NewScanHandler(mockService, logger)

	httpReq := httptest.NewRequest("POST", "/api/v1/scans", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	handler.CreateScan(c)

	// Should return error status
	assert.True(t, w.Code >= 400)
}

func TestScanHandler_GetScans_DefaultPagination(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockScanService)
	mockService.On("GetScans", context.Background(), mock.AnythingOfType("*dto.ScanFilter")).
		Return(&dto.ScanListResponse{
			Data:       []dto.ScanResponse{},
			TotalCount: 0,
			Page:       1,
			PageSize:   10,
		}, nil)

	handler := NewScanHandler(mockService, logger)

	httpReq := httptest.NewRequest("GET", "/api/v1/scans", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	handler.GetScans(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
