package authorization

import (
	"github.com/novriyantoAli/moodly/internal/application/authorization/repository"
	"github.com/novriyantoAli/moodly/internal/application/authorization/service"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		repository.NewAuthorizationRepository,
		service.NewAuthorizationService,
	),
)
