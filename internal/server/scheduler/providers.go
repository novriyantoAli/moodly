package scheduler

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	// Scheduler provider
	fx.Provide(NewScheduler),
)
