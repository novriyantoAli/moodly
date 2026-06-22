package bill

import (
	"github.com/novriyantoAli/moodly/internal/application/bill/handler"
	"github.com/novriyantoAli/moodly/internal/application/bill/publisher"
	"github.com/novriyantoAli/moodly/internal/application/bill/repository"
	"github.com/novriyantoAli/moodly/internal/application/bill/service"
	"github.com/novriyantoAli/moodly/internal/application/bill/worker"
	"github.com/novriyantoAli/moodly/internal/pkg/queue"

	"go.uber.org/fx"
)

// Module provides all bill domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewBillRepository,
		// Provide the queue client as AsynqClient interface
		func(client *queue.Client) publisher.AsynqClient {
			return client
		},
		publisher.NewBillPublisher,
		service.NewBillService,
		handler.NewBillHandler,
	),
	// fx.Provide(
	// 	// Provide the queue client as AsynqClient interface
	// 	func(client *queue.Client) publisher.AsynqClient {
	// 		return client
	// 	},
	// ),
	// fx.Provide(
	// 	publisher.NewBillPublisher,
	// ),
	// fx.Provide(
	// 	service.NewBillService,
	// ),
	// fx.Provide(
	// 	handler.NewBillHandler,
	// ),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	// Then provide bill dependencies
	fx.Provide(
		repository.NewBillRepository,
		func(client *queue.Client) publisher.AsynqClient {
			return client
		},
		publisher.NewBillPublisher,
		service.NewBillService,
		worker.NewBillWorker,
	),
	// fx.Provide(
	// 	func(client *queue.Client) publisher.AsynqClient {
	// 		return client
	// 	},
	// ),
	// fx.Provide(
	// 	publisher.NewBillPublisher,
	// ),
	// fx.Provide(
	// 	service.NewBillService,
	// ),
	// fx.Provide(
	// 	worker.NewBillWorker,
	// ),
)
