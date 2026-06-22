package service

import (
	"context"

	entity "github.com/novriyantoAli/moodly/internal/application/auth/entity"
	repository "github.com/novriyantoAli/moodly/internal/application/auth/repository"

	"go.uber.org/zap"
)

type LoginAttemptService interface {
	CreateAttempt(
		ctx context.Context,
		attempt *entity.LoginAttempt,
	) error

	GetAttemptsByUserID(
		ctx context.Context,
		userID uint,
	) ([]entity.LoginAttempt, error)

	GetAttemptsByUsername(
		ctx context.Context,
		username string,
	) ([]entity.LoginAttempt, error)

	GetLatestAttemptByUserID(
		ctx context.Context,
		userID uint,
	) (*entity.LoginAttempt, error)

	GetLatestAttemptByUsername(
		ctx context.Context,
		username string,
	) (*entity.LoginAttempt, error)

	GetFailedAttemptCountByUserID(
		ctx context.Context,
		userID uint,
	) (int, error)

	GetFailedAttemptCountByUsername(
		ctx context.Context,
		username string,
	) (int, error)
}

type loginAttemptService struct {
	repo   repository.LoginAttemptRepository
	logger *zap.Logger
}

func NewLoginAttemptService(
	repo repository.LoginAttemptRepository,
	logger *zap.Logger,
) LoginAttemptService {

	return &loginAttemptService{
		repo:   repo,
		logger: logger,
	}
}

func (s *loginAttemptService) CreateAttempt(
	ctx context.Context,
	attempt *entity.LoginAttempt,
) error {

	s.logger.Info(
		"Creating login attempt",
		zap.String("username", attempt.Username),
	)

	return s.repo.Create(
		ctx,
		attempt,
	)
}

func (s *loginAttemptService) GetAttemptsByUserID(
	ctx context.Context,
	userID uint,
) ([]entity.LoginAttempt, error) {

	return s.repo.GetByUserID(
		ctx,
		userID,
	)
}

func (s *loginAttemptService) GetAttemptsByUsername(
	ctx context.Context,
	username string,
) ([]entity.LoginAttempt, error) {

	return s.repo.GetByUsername(
		ctx,
		username,
	)
}

func (s *loginAttemptService) GetLatestAttemptByUserID(
	ctx context.Context,
	userID uint,
) (*entity.LoginAttempt, error) {

	attempts, err := s.repo.GetByUserID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, err
	}

	if len(attempts) == 0 {
		return nil, nil
	}

	return &attempts[0], nil
}

func (s *loginAttemptService) GetLatestAttemptByUsername(
	ctx context.Context,
	username string,
) (*entity.LoginAttempt, error) {

	attempts, err := s.repo.GetByUsername(
		ctx,
		username,
	)
	if err != nil {
		return nil, err
	}

	if len(attempts) == 0 {
		return nil, nil
	}

	return &attempts[0], nil
}

func (s *loginAttemptService) GetFailedAttemptCountByUserID(
	ctx context.Context,
	userID uint,
) (int, error) {

	attempts, err := s.repo.GetByUserID(
		ctx,
		userID,
	)
	if err != nil {
		return 0, err
	}

	count := 0

	for _, attempt := range attempts {
		if !attempt.Success {
			count++
		}
	}

	return count, nil
}

func (s *loginAttemptService) GetFailedAttemptCountByUsername(
	ctx context.Context,
	username string,
) (int, error) {

	attempts, err := s.repo.GetByUsername(
		ctx,
		username,
	)
	if err != nil {
		return 0, err
	}

	count := 0

	for _, attempt := range attempts {
		if !attempt.Success {
			count++
		}
	}

	return count, nil
}
