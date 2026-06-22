package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	dto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	entity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	repository "github.com/novriyantoAli/moodly/internal/application/security/repository"

	"go.uber.org/zap"
)

type UserPINService interface {
	SetPIN(ctx context.Context, req *dto.SetPINRequest) error
	VerifyPIN(ctx context.Context, req *dto.VerifyPINRequest) (bool, error)
	GetSecurity(ctx context.Context, userID uint) (*dto.UserPINResponse, error)
	IsAccountLocked(ctx context.Context, userID uint) (bool, error)
}

type userPINService struct {
	repo   repository.UserPINRepository
	logger *zap.Logger
}

func NewUserPINService(repo repository.UserPINRepository, logger *zap.Logger) UserPINService {
	return &userPINService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userPINService) SetPIN(ctx context.Context, req *dto.SetPINRequest) error {
	s.logger.Info("Setting PIN for user", zap.Uint("user_id", req.UserID))

	pinHash := s.hashPIN(req.PIN)

	security, err := s.repo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}

	if security == nil {
		// Create new security record
		return s.repo.Create(ctx, &entity.UserPIN{
			UserID:        req.UserID,
			PinHash:       pinHash,
			FailedAttempt: 0,
			LockedUntil:   nil,
		})
	}

	// Update existing PIN
	return s.repo.UpdatePIN(ctx, req.UserID, pinHash)
}

func (s *userPINService) VerifyPIN(ctx context.Context, req *dto.VerifyPINRequest) (bool, error) {
	s.logger.Info("Verifying PIN for user", zap.Uint("user_id", req.UserID))

	security, err := s.repo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return false, err
	}

	if security == nil {
		s.logger.Warn("Security record not found", zap.Uint("user_id", req.UserID))
		return false, nil
	}

	// Check if account is locked
	if security.LockedUntil != nil && security.LockedUntil.After(time.Now()) {
		s.logger.Warn("Account is locked", zap.Uint("user_id", req.UserID))
		return false, nil
	}

	pinHash := s.hashPIN(req.PIN)
	if security.PinHash == pinHash {
		// Reset failed attempts on successful verification
		if err := s.repo.ResetFailedAttempt(ctx, req.UserID); err != nil {
			s.logger.Error("Failed to reset failed attempts", zap.Error(err))
		}
		return true, nil
	}

	// Increment failed attempts
	if err := s.repo.IncrementFailedAttempt(ctx, req.UserID); err != nil {
		s.logger.Error("Failed to increment failed attempt", zap.Error(err))
	}

	// Lock account after 3 failed attempts for 15 minutes
	if security.FailedAttempt+1 >= 3 {
		if err := s.repo.LockAccount(ctx, req.UserID, 15*time.Minute); err != nil {
			s.logger.Error("Failed to lock account", zap.Error(err))
		}
	}

	return false, nil
}

func (s *userPINService) GetSecurity(ctx context.Context, userID uint) (*dto.UserPINResponse, error) {
	s.logger.Info("Getting security info for user", zap.Uint("user_id", userID))

	security, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user security", zap.Error(err))
		return nil, err
	}

	if security == nil {
		s.logger.Info("User security not found", zap.Uint("user_id", userID))
		s.logger.Info("Returning nil for user security", zap.Uint("user_id", userID))
		return nil, nil
	}

	isLocked := security.LockedUntil != nil && security.LockedUntil.After(time.Now())

	return &dto.UserPINResponse{
		UserID:        security.UserID,
		FailedAttempt: security.FailedAttempt,
		IsLocked:      isLocked,
	}, nil
}

func (s *userPINService) IsAccountLocked(ctx context.Context, userID uint) (bool, error) {
	security, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	if security == nil {
		return false, nil
	}

	return security.LockedUntil != nil && security.LockedUntil.After(time.Now()), nil
}

func (s *userPINService) hashPIN(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return fmt.Sprintf("%x", hash)
}
