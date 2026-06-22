package grpc

import (
	"github.com/novriyantoAli/moodly/internal/application/payment"
	paymentHandler "github.com/novriyantoAli/moodly/internal/application/payment/handler"
	"github.com/novriyantoAli/moodly/internal/application/user"
	userHandler "github.com/novriyantoAli/moodly/internal/application/user/handler"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include domain modules
	user.Module,
	payment.Module,

	// gRPC handlers
	fx.Provide(
		userHandler.NewUserGrpcHandler,
		paymentHandler.NewPaymentGrpcHandler,
		NewServer,
	),
)
