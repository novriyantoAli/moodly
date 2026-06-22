package repository

import (
	"context"

	"github.com/novriyantoAli/moodly/internal/application/scan/dto"
	"github.com/novriyantoAli/moodly/internal/application/scan/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ScanRepository defines the interface for scan data access
type ScanRepository interface {
	Create(ctx context.Context, scan *entity.Scan) error
	GetByID(ctx context.Context, id uint) (*entity.Scan, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*entity.Scan, error)
	GetAll(ctx context.Context, filter *dto.ScanFilter) ([]entity.Scan, int64, error)
	GetByUserID(ctx context.Context, userID uint, filter *dto.ScanFilter) ([]entity.Scan, int64, error)
	Update(ctx context.Context, scan *entity.Scan) error
	Delete(ctx context.Context, id uint) error
}

type scanRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewScanRepository creates a new scan repository
func NewScanRepository(db *gorm.DB, logger *zap.Logger) ScanRepository {
	return &scanRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new scan record
func (r *scanRepository) Create(ctx context.Context, scan *entity.Scan) error {
	r.logger.Info("Creating scan", zap.String("barcode", scan.Barcode), zap.String("transaction_id", scan.TransactionID))
	db := database.GetDB(ctx, r.db)
	return db.Create(scan).Error
}

// GetByID retrieves a scan by ID
func (r *scanRepository) GetByID(ctx context.Context, id uint) (*entity.Scan, error) {
	var scan entity.Scan
	db := database.GetDB(ctx, r.db)
	err := db.First(&scan, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Scan not found", zap.Uint("id", id))
			return nil, err
		}
		r.logger.Error("Failed to get scan by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &scan, nil
}

// GetByTransactionID retrieves a scan by transaction ID
func (r *scanRepository) GetByTransactionID(ctx context.Context, transactionID string) (*entity.Scan, error) {
	var scan entity.Scan
	db := database.GetDB(ctx, r.db)
	err := db.Where("transaction_id = ?", transactionID).First(&scan).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Scan not found", zap.String("transaction_id", transactionID))
			return nil, err
		}
		r.logger.Error("Failed to get scan by transaction_id", zap.String("transaction_id", transactionID), zap.Error(err))
		return nil, err
	}
	return &scan, nil
}

// GetAll retrieves all scans with pagination and filtering
func (r *scanRepository) GetAll(ctx context.Context, filter *dto.ScanFilter) ([]entity.Scan, int64, error) {
	var scans []entity.Scan
	var total int64
	db := database.GetDB(ctx, r.db)

	// Apply filters
	query := db
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Get total count
	if err := query.Model(&entity.Scan{}).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count scans", zap.Error(err))
		return nil, 0, err
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Find(&scans).Error
	if err != nil {
		r.logger.Error("Failed to get scans", zap.Error(err))
		return nil, 0, err
	}

	return scans, total, nil
}

// GetByUserID retrieves scans by user ID with pagination
func (r *scanRepository) GetByUserID(ctx context.Context, userID uint, filter *dto.ScanFilter) ([]entity.Scan, int64, error) {
	var scans []entity.Scan
	var total int64
	db := database.GetDB(ctx, r.db)

	// Apply filters
	query := db.Where("user_id = ?", userID)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Get total count
	if err := query.Model(&entity.Scan{}).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count scans for user", zap.Uint("user_id", userID), zap.Error(err))
		return nil, 0, err
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Find(&scans).Error
	if err != nil {
		r.logger.Error("Failed to get scans for user", zap.Uint("user_id", userID), zap.Error(err))
		return nil, 0, err
	}

	return scans, total, nil
}

// Update updates an existing scan record
func (r *scanRepository) Update(ctx context.Context, scan *entity.Scan) error {
	r.logger.Info("Updating scan", zap.Uint("id", scan.ID), zap.String("status", scan.Status))
	db := database.GetDB(ctx, r.db)
	return db.Save(scan).Error
}

// Delete deletes a scan record (soft delete)
func (r *scanRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Info("Deleting scan", zap.Uint("id", id))
	db := database.GetDB(ctx, r.db)
	return db.Delete(&entity.Scan{}, id).Error
}
