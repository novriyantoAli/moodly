package usecase

import (
	"context"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/application/auth/entity"
	"github.com/novriyantoAli/moodly/internal/application/auth/service"
	common "github.com/novriyantoAli/moodly/internal/application/common/contract"
	securityEntity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	securityService "github.com/novriyantoAli/moodly/internal/application/security/service"
	userDto "github.com/novriyantoAli/moodly/internal/application/user/dto"
	userService "github.com/novriyantoAli/moodly/internal/application/user/service"
	"go.uber.org/zap"
)

type GoogleLoginUseCase interface {
	Execute(ctx context.Context, req *dto.GoogleLoginRequest) (*dto.GoogleLoginResponse, error)
}

type googleLoginUseCase struct {
	userSvc      userService.UserService
	userOAuthSvc securityService.UserOAuthService
	oauthSvc     securityService.OAuthService

	sessionSvc service.AuthSessionService
	attemptSvc service.LoginAttemptService

	tokenService common.TokenService

	logger *zap.Logger
}

func NewGoogleLoginUseCase(
	userSvc userService.UserService,
	userOAuthSvc securityService.UserOAuthService,
	oauthSvc securityService.OAuthService,

	sessionSvc service.AuthSessionService,
	attemptSvc service.LoginAttemptService,

	tokenService common.TokenService,

	logger *zap.Logger,
) GoogleLoginUseCase {

	return &googleLoginUseCase{
		userSvc:      userSvc,
		userOAuthSvc: userOAuthSvc,
		oauthSvc:     oauthSvc,
		sessionSvc:   sessionSvc,
		attemptSvc:   attemptSvc,
		tokenService: tokenService,
		logger:       logger,
	}
}

func (uc *googleLoginUseCase) Execute(
	ctx context.Context,
	req *dto.GoogleLoginRequest,
) (*dto.GoogleLoginResponse, error) {

	googleUser, err := uc.oauthSvc.VerifyGoogleToken(
		ctx,
		req.IDToken,
	)
	if err != nil {
		return nil, err
	}

	userOAuth, err := uc.userOAuthSvc.GetByProviderAndProviderUserID(
		ctx,
		"google",
		googleUser.Subject,
	)

	var userID uint
	var email string
	var level string

	if err == nil {

		user, err := uc.userSvc.GetUserByID(
			ctx,
			userOAuth.UserID,
		)
		if err != nil {
			return nil, err
		}

		userID = user.ID
		email = user.Email
		level = user.Level

	} else {

		user, err := uc.userSvc.CreateUser(
			ctx,
			&userDto.CreateUserRequest{
				Email:    googleUser.Email,
				FullName: googleUser.Name,
			},
		)
		if err != nil {
			return nil, err
		}

		err = uc.userOAuthSvc.Create(
			ctx,
			&securityEntity.UserOAuth{
				UserID:         user.ID,
				Provider:       "google",
				ProviderUserID: googleUser.Subject,
				Email:          googleUser.Email,
				Name:           googleUser.Name,
				Picture:        googleUser.Picture,
			},
		)
		if err != nil {
			return nil, err
		}

		userID = user.ID
		email = user.Email
		level = user.Level
	}

	accessToken, err := uc.tokenService.GenerateToken(
		userID,
		email,
		level,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(
		userID,
		email,
		level,
	)
	if err != nil {
		return nil, err
	}

	expiredAt := time.Now().
		Add(7 * 24 * time.Hour).
		UTC()

	err = uc.sessionSvc.CreateSession(
		ctx,
		&entity.AuthSession{
			UserID:       userID,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiredAt:    expiredAt,
			IPAddress:    req.IPAddress,
			UserAgent:    req.UserAgent,
		},
	)
	if err != nil {
		return nil, err
	}

	_ = uc.attemptSvc.CreateAttempt(
		ctx,
		&entity.LoginAttempt{
			UserID:    &userID,
			Username:  email,
			Success:   true,
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
			Reason:    "google login",
		},
	)

	return &dto.GoogleLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiredAt:    expiredAt.Unix(),
	}, nil
}
