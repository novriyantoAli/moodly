package service

import (
	"context"
	"errors"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/entity"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SubscribeService interface {
	CreateSubscriber(ctx context.Context, req *dto.CreateSubscriberRequest) (*dto.SubscriberResponse, error)
	GetSubscriberByID(ctx context.Context, id uint) (*dto.SubscriberResponse, error)
	GetSubscriberByUsername(ctx context.Context, username string) (*dto.SubscriberResponse, error)
	GetSubscribers(ctx context.Context, filter *dto.SubscribeFilter) (*dto.SubscriberListResponse, error)
	UpdateSubscriber(ctx context.Context, id uint, req *dto.UpdateSubscriberRequest) (*dto.SubscriberResponse, error)
	DeleteSubscriber(ctx context.Context, id uint) error
	CountFilter(ctx context.Context, filter *dto.CountFilter) (*dto.CountResponse, error)
}

type subscribeService struct {
	repo   repository.SubscribeRepository
	logger *zap.Logger
}

func NewSubscribeService(repo repository.SubscribeRepository, logger *zap.Logger) SubscribeService {
	return &subscribeService{
		repo:   repo,
		logger: logger,
	}
}

func (s *subscribeService) CreateSubscriber(ctx context.Context, req *dto.CreateSubscriberRequest) (*dto.SubscriberResponse, error) {
	// Validate username doesn't already exist
	exists, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		s.logger.Error("Failed to check username existence", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Validate plan value
	if !isValidPlan(req.Plan) {
		return nil, errors.New("invalid plan, must be 'pppoe' or 'hotspot'")
	}

	subscriber := &entity.Subscriber{
		Username:  req.Username,
		CallName:  req.CallName,
		Password:  req.Password,
		Plan:      req.Plan,
		Price:     req.Price,
		StartDate: req.StartDate,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.Create(ctx, subscriber)
	if err != nil {
		s.logger.Error("Failed to create subscriber", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(subscriber), nil
}

func (s *subscribeService) GetSubscriberByID(ctx context.Context, id uint) (*dto.SubscriberResponse, error) {
	subscriber, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subscriber not found")
		}
		s.logger.Error("Failed to get subscriber by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(subscriber), nil
}

func (s *subscribeService) GetSubscriberByUsername(ctx context.Context, username string) (*dto.SubscriberResponse, error) {
	subscriber, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subscriber not found")
		}
		s.logger.Error("Failed to get subscriber by username", zap.String("username", username), zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(subscriber), nil
}

func (s *subscribeService) GetSubscribers(ctx context.Context, filter *dto.SubscribeFilter) (*dto.SubscriberListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	subscribers, totalCount, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get subscribers", zap.Error(err))
		return nil, err
	}

	responses := make([]dto.SubscriberResponse, 0, len(subscribers))
	for _, subscriber := range subscribers {
		responses = append(responses, *s.entityToResponse(&subscriber))
	}

	return &dto.SubscriberListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *subscribeService) UpdateSubscriber(ctx context.Context, id uint, req *dto.UpdateSubscriberRequest) (*dto.SubscriberResponse, error) {
	subscriber, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subscriber not found")
		}
		s.logger.Error("Failed to get subscriber", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	// Update fields if provided
	if req.CallName != "" {
		subscriber.CallName = req.CallName
	}
	if req.Password != "" {
		subscriber.Password = req.Password
	}
	if req.Plan != "" {
		if !isValidPlan(req.Plan) {
			return nil, errors.New("invalid plan, must be 'pppoe' or 'hotspot'")
		}
		subscriber.Plan = req.Plan
	}
	if req.Price > 0 {
		subscriber.Price = req.Price
	}
	if !req.StartDate.IsZero() {
		subscriber.StartDate = req.StartDate
	}
	subscriber.IsActive = req.IsActive
	subscriber.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, subscriber)
	if err != nil {
		s.logger.Error("Failed to update subscriber", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(subscriber), nil
}

func (s *subscribeService) DeleteSubscriber(ctx context.Context, id uint) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("subscriber not found")
		}
		s.logger.Error("Failed to get subscriber", zap.Uint("id", id), zap.Error(err))
		return err
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete subscriber", zap.Uint("id", id), zap.Error(err))
		return err
	}

	return nil
}

func (s *subscribeService) CountFilter(ctx context.Context, filter *dto.CountFilter) (*dto.CountResponse, error) {
	count, err := s.repo.Count(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to count subscribers", zap.Error(err))
		return nil, err
	}

	return &dto.CountResponse{
		Plan:     filter.Plan,
		IsActive: filter.IsActive,
		Count:    count,
	}, nil
}

func (s *subscribeService) entityToResponse(subscriber *entity.Subscriber) *dto.SubscriberResponse {
	return &dto.SubscriberResponse{
		ID:        subscriber.ID,
		Username:  subscriber.Username,
		CallName:  subscriber.CallName,
		Plan:      subscriber.Plan,
		Price:     subscriber.Price,
		StartDate: subscriber.StartDate,
		IsActive:  subscriber.IsActive,
		CreatedAt: subscriber.CreatedAt,
		UpdatedAt: subscriber.UpdatedAt,
	}
}

func isValidPlan(plan string) bool {
	validPlans := map[string]bool{
		"pppoe":   true,
		"hotspot": true,
	}
	return validPlans[plan]
}
