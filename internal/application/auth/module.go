package auth

import (
	"github.com/novriyantoAli/moodly/internal/application/auth/handler"
	"github.com/novriyantoAli/moodly/internal/application/auth/repository"
	"github.com/novriyantoAli/moodly/internal/application/auth/service"
	"github.com/novriyantoAli/moodly/internal/application/auth/usecase"
	common "github.com/novriyantoAli/moodly/internal/application/common/contract"
	"github.com/novriyantoAli/moodly/internal/pkg/jwt"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		func(manager *jwt.JWTManager) common.TokenService { return manager },
		repository.NewAuthSessionRepository,
		repository.NewLoginAttemptRepository,
		service.NewAuthSessionService,
		service.NewLoginAttemptService,
		usecase.NewLoginUseCase,
		usecase.NewGoogleLoginUseCase,
		usecase.NewLogoutUseCase,
		handler.NewLoginHandler,
		handler.NewGoogleLoginHandler,
		handler.NewLogoutHandler,
	),
)
