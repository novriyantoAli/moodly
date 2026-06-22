package publisher

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/novriyantoAli/moodly/internal/application/bill/dto"
	"github.com/novriyantoAli/moodly/internal/application/bill/entity"
	"github.com/novriyantoAli/moodly/internal/application/bill/task"
	"go.uber.org/zap"
)

type AsynqClient interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type BillPublisher interface {
	ScheduleBillPerSubscribe(req dto.CreateBillRequest) error
	ScheduleBillPerSubscribeChangeFromUnpaidOverdue(req entity.Bill) error
}

type billPublisher struct {
	client AsynqClient
	logger *zap.Logger
}

func NewBillPublisher(client AsynqClient, logger *zap.Logger) BillPublisher {
	return &billPublisher{
		client: client,
		logger: logger,
	}
}

func (p *billPublisher) ScheduleBillPerSubscribeChangeFromUnpaidOverdue(req entity.Bill) error {

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(task.TypeChangeBillFromUnpaidToOverdue, payloadBytes)
	_, err = p.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	p.logger.Info("Scheduled bill status update from unpaid to overdue",
		zap.String("id", req.ID.String()),
		zap.String("status", string(req.Status)))

	return nil
}
func (p *billPublisher) ScheduleBillPerSubscribe(req dto.CreateBillRequest) error {
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(task.TypeGenerateBillPerSubscribe, payloadBytes)
	_, err = p.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	p.logger.Info("Scheduled bill generation for subscribe",
		zap.Uint("subscribe_id", req.SubscribeID),
		zap.Int("month", int(req.BillMonth)),
		zap.Int("year", int(req.BillYear)))

	return nil
}
