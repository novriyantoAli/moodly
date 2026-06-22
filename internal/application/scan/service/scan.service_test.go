package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/novriyantoAli/moodly/internal/application/scan/dto"
	"github.com/novriyantoAli/moodly/internal/application/scan/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type MockScanRepository struct {
	mock.Mock
}

func (m *MockScanRepository) Create(ctx context.Context, scan *entity.Scan) error {
	args := m.Called(ctx, scan)
	return args.Error(0)
}

func (m *MockScanRepository) GetByID(ctx context.Context, id uint) (*entity.Scan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Scan), args.Error(1)
}

func (m *MockScanRepository) GetByTransactionID(ctx context.Context, transactionID string) (*entity.Scan, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Scan), args.Error(1)
}

func (m *MockScanRepository) GetAll(ctx context.Context, filter *dto.ScanFilter) ([]entity.Scan, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Scan), args.Get(1).(int64), args.Error(2)
}

func (m *MockScanRepository) GetByUserID(ctx context.Context, userID uint, filter *dto.ScanFilter) ([]entity.Scan, int64, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Scan), args.Get(1).(int64), args.Error(2)
}

func (m *MockScanRepository) Update(ctx context.Context, scan *entity.Scan) error {
	args := m.Called(ctx, scan)
	return args.Error(0)
}

func (m *MockScanRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestScanService_CreateScan_Valid(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockScanRepository)
	mockRepo.On("Create", context.Background(), mock.AnythingOfType("*entity.Scan")).Return(nil)

	service := NewScanService(mockRepo, logger)
	ctx := context.Background()

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

	resp, err := service.CreateScan(ctx, 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "BC123456789", resp.Barcode)
	assert.Equal(t, "pending", resp.Status)
	mockRepo.AssertExpectations(t)
}

func TestScanService_CreateScan_MissingBarcode(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockScanRepository)
	service := NewScanService(mockRepo, logger)
	ctx := context.Background()

	req := &dto.CreateScanRequest{
		Barcode:       "",
		TransactionID: "trx-001",
		Pin:           "123456",
	}

	resp, err := service.CreateScan(ctx, 1, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestScanService_GetScanByID_Found(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockScanRepository)

	deviceJSON := datatypes.JSON(`{"user_agent":"Mozilla"}`)
	expectedScan := &entity.Scan{
		ID:         1,
		UserID:     1,
		Barcode:    "BC123456789",
		Status:     string(entity.ScanStatusPending),
		DeviceInfo: deviceJSON,
	}
	mockRepo.On("GetByID", context.Background(), uint(1)).Return(expectedScan, nil)

	service := NewScanService(mockRepo, logger)
	ctx := context.Background()
	resp, err := service.GetScanByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
	mockRepo.AssertExpectations(t)
}

func TestScanService_UpdateScanStatus_Valid(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockScanRepository)
	mockRepo.On("GetByID", context.Background(), uint(1)).Return(&entity.Scan{
		ID:     1,
		Status: string(entity.ScanStatusPending),
	}, nil)
	mockRepo.On("Update", context.Background(), mock.AnythingOfType("*entity.Scan")).Return(nil)

	service := NewScanService(mockRepo, logger)
	ctx := context.Background()
	req := &dto.UpdateScanRequest{
		Status: "completed",
	}

	resp, err := service.UpdateScanStatus(ctx, 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestScanService_UpdateScanStatus_InvalidStatus(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockScanRepository)
	mockRepo.On("GetByID", context.Background(), uint(1)).Return(&entity.Scan{
		ID:     1,
		Status: string(entity.ScanStatusPending),
	}, nil)

	service := NewScanService(mockRepo, logger)
	ctx := context.Background()
	req := &dto.UpdateScanRequest{
		Status: "invalid_status",
	}

	resp, err := service.UpdateScanStatus(ctx, 1, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestScanService_DeleteScan_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockScanRepository)
	mockRepo.On("GetByID", context.Background(), uint(1)).Return(&entity.Scan{ID: 1}, nil)
	mockRepo.On("Delete", context.Background(), uint(1)).Return(nil)

	service := NewScanService(mockRepo, logger)
	ctx := context.Background()
	err := service.DeleteScan(ctx, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestScanService_GetScans_DefaultPagination(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockScanRepository)
	mockRepo.On("GetAll", context.Background(), mock.AnythingOfType("*dto.ScanFilter")).Return([]entity.Scan{
		{ID: 1, Barcode: "BC123", Status: string(entity.ScanStatusPending)},
	}, int64(1), nil)

	service := NewScanService(mockRepo, logger)
	ctx := context.Background()
	resp, err := service.GetScans(ctx, &dto.ScanFilter{Page: 0, PageSize: 0})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestScanService_GetUserScans_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockScanRepository)
	mockRepo.On("GetByUserID", context.Background(), uint(1), mock.AnythingOfType("*dto.ScanFilter")).
		Return([]entity.Scan{
			{ID: 1, UserID: 1, Barcode: "BC123", Status: string(entity.ScanStatusPending)},
		}, int64(1), nil)

	service := NewScanService(mockRepo, logger)
	ctx := context.Background()
	filter := &dto.ScanFilter{Page: 1, PageSize: 10}
	resp, err := service.GetUserScans(ctx, 1, filter)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestScanService_DeviceInfoJSON(t *testing.T) {
	t.Run("should correctly marshal device info", func(t *testing.T) {
		device := dto.DeviceInfo{
			UserAgent:  "Mozilla/5.0",
			Platform:   "Linux",
			Language:   "en-US",
			DeviceType: "mobile",
			Browser:    "Chrome",
		}

		data, err := json.Marshal(device)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var unmarshaled dto.DeviceInfo
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, device.UserAgent, unmarshaled.UserAgent)
	})
}
