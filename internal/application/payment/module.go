package payment

import (
	"github.com/novriyantoAli/moodly/internal/application/payment/handler"
	"github.com/novriyantoAli/moodly/internal/application/payment/repository"
	"github.com/novriyantoAli/moodly/internal/application/payment/service"
	"github.com/novriyantoAli/moodly/internal/application/payment/worker"
	"github.com/novriyantoAli/moodly/internal/pkg/queue"

	"go.uber.org/fx"
	"go.uber.org/zap"

	common "github.com/novriyantoAli/moodly/internal/application/common/contract"
	generator "github.com/novriyantoAli/moodly/internal/pkg/generator"
)

// Module provides all payment domain dependencies
var Module = fx.Options(
	fx.Provide(
		generator.NewPaymentNumberGenerator,
		func(gen *generator.PaymentNumberGenerator) common.PaymentNumberGenerator { return gen },

		repository.NewPaymentRepository,
		service.NewPaymentService,
		handler.NewPaymentHandler,
		// Provide the queue client as AsynqClient interface
		func(client *queue.Client, logger *zap.Logger) worker.AsynqClient {
			return client
		},
		worker.NewPaymentWorker,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewPaymentRepository,
		service.NewPaymentService,
		func(client *queue.Client, logger *zap.Logger) worker.AsynqClient {
			return client
		},
		worker.NewPaymentWorker,
	),
)
