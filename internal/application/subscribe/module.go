package subscribe

import (
	"github.com/novriyantoAli/moodly/internal/application/subscribe/handler"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/repository"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/service"

	"go.uber.org/fx"
)

// Module provides all subscribe domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewSubscribeRepository,
		service.NewSubscribeService,
		handler.NewSubscribeHandler,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewSubscribeRepository,
		service.NewSubscribeService,
	),
)
