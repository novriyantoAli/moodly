package api

import (
	"github.com/novriyantoAli/moodly/internal/application/auth"
	"github.com/novriyantoAli/moodly/internal/application/bill"
	"github.com/novriyantoAli/moodly/internal/application/consultation"
	"github.com/novriyantoAli/moodly/internal/application/oauth"
	"github.com/novriyantoAli/moodly/internal/application/payment"
	"github.com/novriyantoAli/moodly/internal/application/scan"
	"github.com/novriyantoAli/moodly/internal/application/security"
	"github.com/novriyantoAli/moodly/internal/application/subscribe"
	"github.com/novriyantoAli/moodly/internal/application/user"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include all domain modules
	auth.Module,
	oauth.Module,
	user.Module,
	payment.Module,
	subscribe.Module,
	bill.Module,
	scan.Module,
	security.Module,
	consultation.Module,
	// API api
	fx.Provide(NewServer),
)
