package security

import (
	common "github.com/novriyantoAli/moodly/internal/application/common/contract"
	"github.com/novriyantoAli/moodly/internal/application/security/handler"
	"github.com/novriyantoAli/moodly/internal/application/security/repository"
	"github.com/novriyantoAli/moodly/internal/application/security/service"
	"github.com/novriyantoAli/moodly/internal/pkg/google"

	"go.uber.org/fx"
)

// Module provides all user-security domain dependencies
var Module = fx.Options(
	fx.Provide(
		google.NewGoogleTokenValidator,
		func(validator *google.GoogleTokenValidator) common.GoogleTokenValidator { return validator },
		repository.NewUserPINRepository,
		repository.NewUserPasswordRepository,
		repository.NewUserOAuthRepository,
		service.NewUserPINService,
		service.NewUserPasswordService,
		service.NewUserOAuthService,
		service.NewOAuthService,
		handler.NewUserPINHandler,
		handler.NewUserPasswordHandler,
	),
)
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewUserPINRepository,
		repository.NewUserPasswordRepository,
		service.NewUserPINService,
		service.NewUserPasswordService,
	),
)
