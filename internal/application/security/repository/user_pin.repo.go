package repository

import (
	"context"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/security/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserPINRepository interface {
	GetByUserID(ctx context.Context, userID uint) (*entity.UserPIN, error)
	Create(ctx context.Context, security *entity.UserPIN) error
	UpdatePIN(ctx context.Context, userID uint, pinHash string) error
	IncrementFailedAttempt(ctx context.Context, userID uint) error
	ResetFailedAttempt(ctx context.Context, userID uint) error
	LockAccount(ctx context.Context, userID uint, duration time.Duration) error
	Unlock(ctx context.Context, userID uint) error
	Delete(ctx context.Context, userID uint) error
}

type userPINRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserPINRepository(db *gorm.DB, logger *zap.Logger) UserPINRepository {
	return &userPINRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userPINRepository) GetByUserID(ctx context.Context, userID uint) (*entity.UserPIN, error) {
	r.logger.Info("Getting user PIN", zap.Uint("user_id", userID))
	db := database.GetDB(ctx, r.db)
	var pin entity.UserPIN
	if err := db.Where("user_id = ?", userID).First(&pin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &pin, nil
}

func (r *userPINRepository) Create(ctx context.Context, security *entity.UserPIN) error {
	r.logger.Info("Creating user PIN", zap.Uint("user_id", security.UserID))
	db := database.GetDB(ctx, r.db)
	return db.Create(security).Error
}

func (r *userPINRepository) UpdatePIN(ctx context.Context, userID uint, pinHash string) error {
	r.logger.Info("Updating user PIN", zap.Uint("user_id", userID))
	db := database.GetDB(ctx, r.db)
	return db.Model(&entity.UserPIN{}).Where("user_id = ?", userID).Update("pin_hash", pinHash).Error
}

func (r *userPINRepository) IncrementFailedAttempt(ctx context.Context, userID uint) error {
	r.logger.Info("Incrementing failed attempt", zap.Uint("user_id", userID))
	db := database.GetDB(ctx, r.db)
	return db.Model(&entity.UserPIN{}).Where("user_id = ?", userID).Update("failed_attempt", gorm.Expr("failed_attempt + ?", 1)).Error
}

func (r *userPINRepository) ResetFailedAttempt(ctx context.Context, userID uint) error {
	r.logger.Info("Resetting failed attempt", zap.Uint("user_id", userID))
	db := database.GetDB(ctx, r.db)
	return db.Model(&entity.UserPIN{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"failed_attempt": 0,
		"locked_until":   nil,
	}).Error
}

func (r *userPINRepository) LockAccount(ctx context.Context, userID uint, duration time.Duration) error {
	r.logger.Info("Locking user account", zap.Uint("user_id", userID))
	db := database.GetDB(ctx, r.db)
	lockedUntil := time.Now().Add(duration)
	return db.Model(&entity.UserPIN{}).Where("user_id = ?", userID).Update("locked_until", lockedUntil).Error
}

func (r *userPINRepository) Unlock(ctx context.Context, userID uint) error {
	r.logger.Info("Unlocking user account", zap.Uint("user_id", userID))
	db := database.GetDB(ctx, r.db)
	return db.Model(&entity.UserPIN{}).Where("user_id = ?", userID).Update("locked_until", nil).Error
}

func (r *userPINRepository) Delete(ctx context.Context, userID uint) error {
	r.logger.Info("Deleting user PIN", zap.Uint("user_id", userID))
	db := database.GetDB(ctx, r.db)
	return db.Where("user_id = ?", userID).Delete(&entity.UserPIN{}).Error
}
