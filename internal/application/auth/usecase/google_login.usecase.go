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
	authService "github.com/novriyantoAli/moodly/internal/application/authorization/service"
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

	authSvc      authService.AuthorizationService
	tokenService common.TokenService

	logger *zap.Logger
}

func NewGoogleLoginUseCase(
	userSvc userService.UserService,
	userOAuthSvc securityService.UserOAuthService,
	oauthSvc securityService.OAuthService,

	sessionSvc service.AuthSessionService,
	attemptSvc service.LoginAttemptService,

	authSvc authService.AuthorizationService,
	tokenService common.TokenService,

	logger *zap.Logger,
) GoogleLoginUseCase {

	return &googleLoginUseCase{
		userSvc:      userSvc,
		userOAuthSvc: userOAuthSvc,
		oauthSvc:     oauthSvc,
		sessionSvc:   sessionSvc,
		attemptSvc:   attemptSvc,
		authSvc:      authSvc,
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

	var user *userDto.UserResponse
	isNewUser := false

	if err == nil {
		user, err = uc.userSvc.GetUserByID(
			ctx,
			userOAuth.UserID,
		)
		if err != nil {
			return nil, err
		}
	} else {
		user, err = uc.userSvc.CreateUser(
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
		isNewUser = true
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

	expiredAt := time.Now().
		Add(7 * 24 * time.Hour).
		UTC()

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
			Reason:    "google login",
		},
	)

	return &dto.GoogleLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiredAt:    expiredAt.Unix(),
		UserID:       user.ID,
		IsNewUser:    isNewUser,
		Roles:        roleNames,
		Permissions:  permissions,
	}, nil
}
