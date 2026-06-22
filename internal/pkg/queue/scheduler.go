package queue

import (
	"context"
	"fmt"

	"github.com/novriyantoAli/moodly/internal/config"

	"github.com/hibiken/asynq"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Scheduler struct {
	scheduler *asynq.Scheduler
	logger    *zap.Logger
	cfg       *config.Config
}

func NewScheduler(cfg *config.Config, logger *zap.Logger) *Scheduler {
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	s := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Logger: NewAsynqLogger(logger),
	})

	logger.Info("Scheduler initialized",
		zap.String("redis_addr", redisAddr))

	return &Scheduler{
		scheduler: s,
		logger:    logger,
		cfg:       cfg,
	}
}

func (s *Scheduler) RegisterJobs(jobName string) {
	// tiap 1 menit → cek payment pending
	_, err := s.scheduler.Register(
		"* * * * *",
		asynq.NewTask(jobName, nil),
		asynq.Queue("low"),
	)

	if err != nil {
		s.logger.Fatal("failed to register job", zap.Error(err))
	}

	s.logger.Info("Scheduler jobs registered")
}

func (s *Scheduler) Start(lifecycle fx.Lifecycle) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				s.logger.Info("Starting scheduler")
				if err := s.scheduler.Run(); err != nil {
					s.logger.Fatal("Scheduler failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.logger.Info("Stopping scheduler")
			s.scheduler.Shutdown()
			return nil
		},
	})
}
