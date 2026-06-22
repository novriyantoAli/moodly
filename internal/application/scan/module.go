package scan

import (
	"github.com/novriyantoAli/moodly/internal/application/scan/handler"
	"github.com/novriyantoAli/moodly/internal/application/scan/repository"
	"github.com/novriyantoAli/moodly/internal/application/scan/service"

	"go.uber.org/fx"
)

// Module provides scan domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewScanRepository,
		service.NewScanService,
		handler.NewScanHandler,
	),
)
