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
	authService "github.com/novriyantoAli/moodly/internal/application/authorization/service"
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
	authSvc         authService.AuthorizationService
	tokenService    common.TokenService
	logger          *zap.Logger
}

func NewLoginUseCase(
	userSvc userService.UserService,
	userPasswordSvc securityService.UserPasswordService,
	sessionSvc service.AuthSessionService,
	attemptSvc service.LoginAttemptService,
	authSvc authService.AuthorizationService,
	tokenService common.TokenService,
	logger *zap.Logger,
) LoginUseCase {
	return &loginUseCase{
		userSvc:         userSvc,
		userPasswordSvc: userPasswordSvc,
		sessionSvc:      sessionSvc,
		attemptSvc:      attemptSvc,
		authSvc:         authSvc,
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

	// Ambil roles dan permissions
	rolesEntities, err := uc.authSvc.GetRolesByUserID(ctx, user.ID)
	if err != nil {
		uc.logger.Warn("Failed to fetch roles", zap.Error(err))
	}
	var roleNames []string
	for _, r := range rolesEntities {
		roleNames = append(roleNames, r.Name)
	}

	permissions, err := uc.authSvc.GetPermissionsByRoles(ctx, roleNames)
	if err != nil {
		uc.logger.Warn("Failed to fetch permissions", zap.Error(err))
	}
	if permissions == nil {
		permissions = []string{}
	}
	if roleNames == nil {
		roleNames = []string{}
	}

	accessToken, err := uc.tokenService.GenerateToken(
		user.ID,
		user.Email,
		roleNames,
		permissions,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(
		user.ID,
		user.Email,
		roleNames,
		permissions,
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
		Roles:        roleNames,
		Permissions:  permissions,
	}, nil
}
