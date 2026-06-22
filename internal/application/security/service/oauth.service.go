package service

import (
	"context"
	"fmt"

	"github.com/novriyantoAli/moodly/internal/application/common/contract"
	"github.com/novriyantoAli/moodly/internal/application/security/dto"
	"github.com/novriyantoAli/moodly/internal/config"
	"go.uber.org/zap"
)

type OAuthService interface {
	VerifyGoogleToken(ctx context.Context, token string) (*dto.GoogleUserInfo, error)
}

type oauthService struct {
	logger    *zap.Logger
	clientID  string
	validator contract.GoogleTokenValidator
}

func NewOAuthService(
	validator contract.GoogleTokenValidator,
	cfg *config.Config,
	logger *zap.Logger,
) OAuthService {
	return &oauthService{
		logger:    logger,
		clientID:  cfg.OAuth.Google.ClientID,
		validator: validator,
	}
}

func (s *oauthService) VerifyGoogleToken(ctx context.Context, token string) (*dto.GoogleUserInfo, error) {
	s.logger.Info(
		"Verifying google token",
	)

	payload, err := s.validator.Validate(ctx, token, s.clientID)

	if err != nil {
		s.logger.Error(
			"Failed verify google token",
			zap.Error(err),
		)

		return nil, fmt.Errorf(
			"invalid google token: %w",
			err,
		)
	}

	info := &dto.GoogleUserInfo{
		Subject: payload.Subject,
	}

	if email, ok := payload.Claims["email"].(string); ok {
		info.Email = email
	}

	if name, ok := payload.Claims["name"].(string); ok {
		info.Name = name
	}

	if picture, ok := payload.Claims["picture"].(string); ok {
		info.Picture = picture
	}

	return info, nil
}
