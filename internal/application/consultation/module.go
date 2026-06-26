package consultation

import (
	"github.com/novriyantoAli/moodly/internal/application/consultation/handler"
	"github.com/novriyantoAli/moodly/internal/application/consultation/repository"
	"github.com/novriyantoAli/moodly/internal/application/consultation/service"
	"github.com/novriyantoAli/moodly/internal/application/consultation/usecase"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		repository.NewConsultationRepository,
		service.NewConsultationService,
		usecase.NewConsultationUsecase,
		handler.NewConsultationHandler,
	),
)
