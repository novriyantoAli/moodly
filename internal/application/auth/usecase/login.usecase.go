package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/application/auth/entity"
	"github.com/novriyantoAli/moodly/internal/application/auth/service"
	common "github.com/novriyantoAli/moodly/internal/application/common/contract"
	securityDto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	securityService "github.com/novriyantoAli/moodly/internal/application/security/service"
	userService "github.com/novriyantoAli/moodly/internal/application/user/service"
	"go.uber.org/zap"
)

type LoginUseCase interface {
	Execute(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
}

type loginUseCase struct {
	userSvc         userService.UserService
	userPasswordSvc securityService.UserPasswordService
	sessionSvc      service.AuthSessionService
	attemptSvc      service.LoginAttemptService
	tokenService    common.TokenService
	logger          *zap.Logger
}

func NewLoginUseCase(
	userSvc userService.UserService,
	userPasswordSvc securityService.UserPasswordService,
	sessionSvc service.AuthSessionService,
	attemptSvc service.LoginAttemptService,
	tokenService common.TokenService,
	logger *zap.Logger,
) LoginUseCase {
	return &loginUseCase{
		userSvc:         userSvc,
		userPasswordSvc: userPasswordSvc,
		sessionSvc:      sessionSvc,
		attemptSvc:      attemptSvc,
		tokenService:    tokenService,
		logger:          logger,
	}
}

func (uc *loginUseCase) Execute(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {

	valid, err := uc.userPasswordSvc.VerifyPassword(
		ctx,
		&securityDto.VerifyPasswordRequest{
			Username: req.Username,
			Password: req.Password,
		},
	)
	if err != nil {
		return nil, err
	}

	if !valid {
		_ = uc.attemptSvc.CreateAttempt(
			ctx,
			&entity.LoginAttempt{
				UserID:    nil,
				Username:  req.Username,
				Success:   false,
				IPAddress: req.IPAddress,
				UserAgent: req.UserAgent,
				Reason:    "invalid username or password",
				// Error:     err.Error(), --- IGNORE ---
			},
		)

		return nil, fmt.Errorf(
			"invalid username or password",
		)
	}

	// ambil user dari user service
	user, err := uc.userSvc.GetUserByEmail(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	accessToken, err := uc.tokenService.GenerateToken(
		user.ID,
		user.Email,
		user.Level,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(
		user.ID,
		user.Email,
		user.Level,
	)
	if err != nil {
		return nil, err
	}

	expiredAt := time.Now().Add(7 * 24 * time.Hour).UTC()

	err = uc.sessionSvc.CreateSession(
		ctx,
		&entity.AuthSession{
			UserID:       user.ID,
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
			UserID:    &user.ID,
			Username:  user.Email,
			Success:   true,
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
		},
	)

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiredAt:    expiredAt.Unix(),
		UserID:       user.ID,
	}, nil
}
