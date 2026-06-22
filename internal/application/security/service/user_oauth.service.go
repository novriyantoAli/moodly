package service

import (
	"context"
	"errors"

	entity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	repository "github.com/novriyantoAli/moodly/internal/application/security/repository"
	"go.uber.org/zap"
)

type UserOAuthService interface {
	GetByProviderAndProviderUserID(ctx context.Context, provider string, providerUserID string) (*entity.UserOAuth, error)
	Create(ctx context.Context, oauth *entity.UserOAuth) error
}
type userOAuthService struct {
	repo   repository.UserOAuthRepository
	logger *zap.Logger
}

func NewUserOAuthService(
	repo repository.UserOAuthRepository,
	logger *zap.Logger,
) UserOAuthService {

	return &userOAuthService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userOAuthService) GetByProviderAndProviderUserID(
	ctx context.Context,
	provider string,
	providerUserID string,
) (*entity.UserOAuth, error) {

	if provider == "" {
		return nil, errors.New("provider is required")
	}

	if providerUserID == "" {
		return nil, errors.New("provider user id is required")
	}

	return s.repo.GetByProviderAndUserID(
		ctx,
		provider,
		providerUserID,
	)
}

func (s *userOAuthService) Create(ctx context.Context, oauth *entity.UserOAuth) error {

	if oauth == nil {
		return errors.New("oauth data is required")
	}

	if oauth.UserID == 0 {
		return errors.New("user id is required")
	}

	if oauth.Provider == "" {
		return errors.New("provider is required")
	}

	if oauth.ProviderUserID == "" {
		return errors.New("provider user id is required")
	}

	return s.repo.Create(
		ctx,
		oauth,
	)
}
