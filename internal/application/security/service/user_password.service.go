package service

import (
	"context"
	"fmt"
	"time"

	dto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	entity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	repository "github.com/novriyantoAli/moodly/internal/application/security/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserPasswordService interface {
	SetPassword(
		ctx context.Context,
		req *dto.SetPasswordRequest,
	) error

	VerifyPassword(
		ctx context.Context,
		req *dto.VerifyPasswordRequest,
	) (bool, error)

	ChangePassword(
		ctx context.Context,
		req *dto.ChangePasswordRequest,
	) error

	GetPasswordInfo(
		ctx context.Context,
		userID uint,
	) (*dto.UserPasswordResponse, error)

	IsAccountLocked(
		ctx context.Context,
		userID uint,
	) (bool, error)

	DeletePassword(
		ctx context.Context,
		userID uint,
	) error
}

type userPasswordService struct {
	repo   repository.UserPasswordRepository
	logger *zap.Logger
}

func NewUserPasswordService(
	repo repository.UserPasswordRepository,
	logger *zap.Logger,
) UserPasswordService {
	return &userPasswordService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userPasswordService) SetPassword(
	ctx context.Context,
	req *dto.SetPasswordRequest,
) error {

	s.logger.Info(
		"Setting password",
		zap.Uint("user_id", req.UserID),
	)

	hash, err := s.hashPassword(req.Password)
	if err != nil {
		return err
	}

	password, err := s.repo.GetByUserID(
		ctx,
		req.UserID,
	)
	if err != nil {
		return err
	}

	if password == nil {
		return s.repo.Create(
			ctx,
			&entity.UserPassword{
				UserID:        req.UserID,
				Username:      req.Username,
				PasswordHash:  hash,
				FailedAttempt: 0,
			},
		)
	}

	return s.repo.UpdatePasswordHash(
		ctx,
		req.UserID,
		hash,
	)
}

func (s *userPasswordService) VerifyPassword(
	ctx context.Context,
	req *dto.VerifyPasswordRequest,
) (bool, error) {

	password, err := s.repo.GetByUsername(
		ctx,
		req.Username,
	)
	if err != nil {
		return false, err
	}

	if password == nil {
		return false, nil
	}

	if password.LockedUntil != nil &&
		password.LockedUntil.After(time.Now()) {

		return false, nil
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(password.PasswordHash),
		[]byte(req.Password),
	)

	if err == nil {

		_ = s.repo.ResetFailedAttempt(
			ctx,
			password.UserID,
		)

		_ = s.repo.UpdateLastLogin(
			ctx,
			password.UserID,
			time.Now(),
		)

		return true, nil
	}

	_ = s.repo.IncrementFailedAttempt(
		ctx,
		password.UserID,
	)

	if password.FailedAttempt+1 >= 5 {

		_ = s.repo.LockAccount(
			ctx,
			password.UserID,
			15*time.Minute,
		)
	}

	return false, nil
}

func (s *userPasswordService) ChangePassword(
	ctx context.Context,
	req *dto.ChangePasswordRequest,
) error {

	password, err := s.repo.GetByUserID(
		ctx,
		req.UserID,
	)
	if err != nil {
		return err
	}

	if password == nil {
		return fmt.Errorf("password not found")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(password.PasswordHash),
		[]byte(req.CurrentPassword),
	)
	if err != nil {
		return fmt.Errorf("invalid current password")
	}

	hash, err := s.hashPassword(
		req.NewPassword,
	)
	if err != nil {
		return err
	}

	return s.repo.UpdatePasswordHash(
		ctx,
		req.UserID,
		hash,
	)
}

func (s *userPasswordService) GetPasswordInfo(
	ctx context.Context,
	userID uint,
) (*dto.UserPasswordResponse, error) {

	password, err := s.repo.GetByUserID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, err
	}

	if password == nil {
		return nil, nil
	}

	return &dto.UserPasswordResponse{
		UserID:        password.UserID,
		Username:      password.Username,
		FailedAttempt: password.FailedAttempt,
		IsLocked: password.LockedUntil != nil &&
			password.LockedUntil.After(time.Now()),
	}, nil
}

func (s *userPasswordService) IsAccountLocked(
	ctx context.Context,
	userID uint,
) (bool, error) {

	password, err := s.repo.GetByUserID(
		ctx,
		userID,
	)
	if err != nil {
		return false, err
	}

	if password == nil {
		return false, nil
	}

	return password.LockedUntil != nil &&
			password.LockedUntil.After(time.Now()),
		nil
}

func (s *userPasswordService) DeletePassword(
	ctx context.Context,
	userID uint,
) error {

	return s.repo.Delete(
		ctx,
		userID,
	)
}

func (s *userPasswordService) hashPassword(
	password string,
) (string, error) {

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
