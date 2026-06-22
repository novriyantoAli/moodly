package usecase

import (
	"context"

	"github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/application/auth/service"
)

type LogoutUseCase interface {
	Execute(ctx context.Context, req *dto.LogoutRequest) error
}

type logoutUseCase struct {
	sessionSvc service.AuthSessionService
}

func NewLogoutUseCase(
	sessionSvc service.AuthSessionService,
) LogoutUseCase {

	return &logoutUseCase{
		sessionSvc: sessionSvc,
	}
}

func (uc *logoutUseCase) Execute(ctx context.Context, req *dto.LogoutRequest) error {
	return uc.sessionSvc.LogoutByRefreshToken(
		ctx,
		req.RefreshToken,
	)
}
