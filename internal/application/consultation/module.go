package consultation

import (
	"github.com/novriyantoAli/moodly/internal/application/consultation/handler"
	"github.com/novriyantoAli/moodly/internal/application/consultation/repository"
	"github.com/novriyantoAli/moodly/internal/application/consultation/service"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		repository.NewConsultationRepository,
		service.NewConsultationService,
		handler.NewConsultationHandler,
	),
)
