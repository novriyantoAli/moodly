package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/novriyantoAli/moodly/internal/application/scan/dto"
	"github.com/novriyantoAli/moodly/internal/application/scan/entity"
	"github.com/novriyantoAli/moodly/internal/application/scan/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ScanService defines the interface for scan business logic
type ScanService interface {
	CreateScan(ctx context.Context, userID uint, req *dto.CreateScanRequest) (*dto.ScanResponse, error)
	GetScanByID(ctx context.Context, id uint) (*dto.ScanResponse, error)
	GetScans(ctx context.Context, filter *dto.ScanFilter) (*dto.ScanListResponse, error)
	GetUserScans(ctx context.Context, userID uint, filter *dto.ScanFilter) (*dto.ScanListResponse, error)
	UpdateScanStatus(ctx context.Context, id uint, req *dto.UpdateScanRequest) (*dto.ScanResponse, error)
	DeleteScan(ctx context.Context, id uint) error
}

type scanService struct {
	repo   repository.ScanRepository
	logger *zap.Logger
}

// NewScanService creates a new scan service
func NewScanService(repo repository.ScanRepository, logger *zap.Logger) ScanService {
	return &scanService{
		repo:   repo,
		logger: logger,
	}
}

// CreateScan creates a new scan
func (s *scanService) CreateScan(ctx context.Context, userID uint, req *dto.CreateScanRequest) (*dto.ScanResponse, error) {
	// Validate request
	if req.Barcode == "" {
		return nil, errors.New("barcode is required")
	}
	if req.TransactionID == "" {
		return nil, errors.New("transaction_id is required")
	}
	if req.Pin == "" {
		return nil, errors.New("pin is required")
	}

	// Marshal device info to JSON
	deviceInfoJSON, err := json.Marshal(req.Device)
	if err != nil {
		s.logger.Error("Failed to marshal device info", zap.Error(err))
		return nil, errors.New("invalid device info")
	}

	// Create scan entity
	scan := &entity.Scan{
		UserID:        userID,
		Barcode:       req.Barcode,
		Timestamp:     req.Timestamp,
		TransactionID: req.TransactionID,
		Pin:           req.Pin,
		Photo:         req.Photo,
		DeviceInfo:    deviceInfoJSON,
		PhotoSize:     req.PhotoSize,
		Status:        string(entity.ScanStatusPending),
	}

	// Save to database
	if err := s.repo.Create(ctx, scan); err != nil {
		s.logger.Error("Failed to create scan", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(scan)
}

// GetScanByID retrieves a scan by ID
func (s *scanService) GetScanByID(ctx context.Context, id uint) (*dto.ScanResponse, error) {
	scan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("scan not found")
		}
		s.logger.Error("Failed to get scan", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(scan)
}

// GetScans retrieves all scans with pagination
func (s *scanService) GetScans(ctx context.Context, filter *dto.ScanFilter) (*dto.ScanListResponse, error) {
	// Set default pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	scans, total, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get scans", zap.Error(err))
		return nil, err
	}

	responses := make([]dto.ScanResponse, 0, len(scans))
	for _, scan := range scans {
		resp, err := s.entityToResponse(&scan)
		if err != nil {
			s.logger.Warn("Failed to convert scan to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return &dto.ScanListResponse{
		Data:       responses,
		TotalCount: total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

// GetUserScans retrieves scans for a specific user
func (s *scanService) GetUserScans(ctx context.Context, userID uint, filter *dto.ScanFilter) (*dto.ScanListResponse, error) {
	// Set default pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	scans, total, err := s.repo.GetByUserID(ctx, userID, filter)
	if err != nil {
		s.logger.Error("Failed to get user scans", zap.Uint("user_id", userID), zap.Error(err))
		return nil, err
	}

	responses := make([]dto.ScanResponse, 0, len(scans))
	for _, scan := range scans {
		resp, err := s.entityToResponse(&scan)
		if err != nil {
			s.logger.Warn("Failed to convert scan to response", zap.Error(err))
			continue
		}
		responses = append(responses, *resp)
	}

	return &dto.ScanListResponse{
		Data:       responses,
		TotalCount: total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

// UpdateScanStatus updates the status of a scan
func (s *scanService) UpdateScanStatus(ctx context.Context, id uint, req *dto.UpdateScanRequest) (*dto.ScanResponse, error) {
	// Validate status
	validStatuses := map[string]bool{
		string(entity.ScanStatusPending):   true,
		string(entity.ScanStatusCompleted): true,
		string(entity.ScanStatusFailed):    true,
	}
	if !validStatuses[req.Status] {
		return nil, errors.New("invalid status")
	}

	// Get existing scan
	scan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("scan not found")
		}
		return nil, err
	}

	// Update status
	scan.Status = req.Status
	if err := s.repo.Update(ctx, scan); err != nil {
		s.logger.Error("Failed to update scan status", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(scan)
}

// DeleteScan deletes a scan
func (s *scanService) DeleteScan(ctx context.Context, id uint) error {
	// Check if scan exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("scan not found")
		}
		return err
	}

	// Delete scan
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete scan", zap.Uint("id", id), zap.Error(err))
		return err
	}

	return nil
}

// entityToResponse converts a scan entity to a response DTO
func (s *scanService) entityToResponse(scan *entity.Scan) (*dto.ScanResponse, error) {
	// Unmarshal device info
	var device dto.DeviceInfo
	if err := json.Unmarshal(scan.DeviceInfo, &device); err != nil {
		s.logger.Warn("Failed to unmarshal device info", zap.Error(err))
		device = dto.DeviceInfo{}
	}

	return &dto.ScanResponse{
		ID:            scan.ID,
		UserID:        scan.UserID,
		Barcode:       scan.Barcode,
		Timestamp:     scan.Timestamp,
		TransactionID: scan.TransactionID,
		Pin:           scan.Pin,
		Photo:         scan.Photo,
		Device:        device,
		PhotoSize:     scan.PhotoSize,
		Status:        scan.Status,
		CreatedAt:     scan.CreatedAt,
		UpdatedAt:     scan.UpdatedAt,
	}, nil
}
