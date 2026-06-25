package scheduler

import (
	"github.com/novriyantoAli/moodly/internal/application/bill/task"
	"github.com/novriyantoAli/moodly/internal/pkg/queue"

	"go.uber.org/zap"
)

type Scheduler struct {
	queueScheduler *queue.Scheduler
	logger         *zap.Logger
}

func NewScheduler(
	queueScheduler *queue.Scheduler,
	logger *zap.Logger,
) *Scheduler {
	return &Scheduler{
		queueScheduler: queueScheduler,
		logger:         logger,
	}
}

func (s *Scheduler) RegisterJobs() {
	s.logger.Info("Registering worker handlers")

	// Register bill workers
	s.queueScheduler.RegisterJobs(task.TypeGenerateMonthlyBills)

	s.queueScheduler.RegisterJobs(task.TypeCheckUnpaidBills)

	s.logger.Info("Worker handlers registered successfully")
}
