package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestSubscribeService_CreateSubscriber(t *testing.T) {
	ctx := context.Background()

	t.Run("should create subscriber successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		req := &dto.CreateSubscriberRequest{
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
		}

		// Mock expectations
		mockRepo.On("UsernameExists", ctx, req.Username).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Subscriber")).Return(nil).Run(func(args mock.Arguments) {
			subscriber := args.Get(1).(*entity.Subscriber)
			subscriber.ID = 1
		})

		// When
		response, err := service.CreateSubscriber(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, req.Username, response.Username)
		assert.Equal(t, req.CallName, response.CallName)
		assert.Equal(t, req.Plan, response.Plan)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when username already exists", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		req := &dto.CreateSubscriberRequest{
			Username:  "existing_user",
			CallName:  "Existing User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
		}

		// Mock expectations
		mockRepo.On("UsernameExists", ctx, req.Username).Return(true, nil)

		// When
		response, err := service.CreateSubscriber(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "username already exists", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when plan is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		req := &dto.CreateSubscriberRequest{
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "invalid_plan",
			Price:     50000.0,
			StartDate: time.Now(),
		}

		// Mock expectations
		mockRepo.On("UsernameExists", ctx, req.Username).Return(false, nil)

		// When
		response, err := service.CreateSubscriber(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "invalid plan, must be 'pppoe' or 'hotspot'", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		req := &dto.CreateSubscriberRequest{
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
		}

		// Mock expectations
		mockRepo.On("UsernameExists", ctx, req.Username).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Subscriber")).Return(errors.New("database error"))

		// When
		response, err := service.CreateSubscriber(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestSubscribeService_GetSubscriberByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get subscriber by ID successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		subscriber := &entity.Subscriber{
			ID:        1,
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
			IsActive:  true,
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(1)).Return(subscriber, nil)

		// When
		response, err := service.GetSubscriberByID(ctx, 1)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, subscriber.ID, response.ID)
		assert.Equal(t, subscriber.Username, response.Username)
		assert.Equal(t, subscriber.CallName, response.CallName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when subscriber not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetSubscriberByID(ctx, 999)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "subscriber not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestSubscribeService_GetSubscriberByUsername(t *testing.T) {
	ctx := context.Background()

	t.Run("should get subscriber by username successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		subscriber := &entity.Subscriber{
			ID:        1,
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
			IsActive:  true,
		}

		// Mock expectations
		mockRepo.On("GetByUsername", ctx, "testuser").Return(subscriber, nil)

		// When
		response, err := service.GetSubscriberByUsername(ctx, "testuser")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, subscriber.Username, response.Username)
		assert.Equal(t, subscriber.CallName, response.CallName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when username not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		// Mock expectations
		mockRepo.On("GetByUsername", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetSubscriberByUsername(ctx, "nonexistent")

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "subscriber not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestSubscribeService_GetSubscribers(t *testing.T) {
	ctx := context.Background()

	t.Run("should get subscribers with pagination", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		subscribers := []entity.Subscriber{
			{
				ID:        1,
				Username:  "user1",
				CallName:  "User 1",
				Plan:      "pppoe",
				Price:     50000.0,
				IsActive:  true,
				StartDate: time.Now(),
			},
			{
				ID:        2,
				Username:  "user2",
				CallName:  "User 2",
				Plan:      "hotspot",
				Price:     30000.0,
				IsActive:  true,
				StartDate: time.Now(),
			},
		}

		filter := &dto.SubscribeFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", ctx, filter).Return(subscribers, int64(2), nil)

		// When
		response, err := service.GetSubscribers(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, int64(2), response.TotalCount)
		assert.Equal(t, "user1", response.Data[0].Username)
		assert.Equal(t, "user2", response.Data[1].Username)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should set default page and page size", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		subscribers := []entity.Subscriber{}
		filter := &dto.SubscribeFilter{
			Page:     0,
			PageSize: 0,
		}

		// Mock expectations - verify the filter is corrected to page 1, pagesize 10
		mockRepo.On("GetAll", ctx, mock.MatchedBy(func(f *dto.SubscribeFilter) bool {
			return f.Page == 1 && f.PageSize == 10
		})).Return(subscribers, int64(0), nil)

		// When
		response, err := service.GetSubscribers(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		filter := &dto.SubscribeFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", ctx, filter).Return(nil, int64(0), errors.New("database error"))

		// When
		response, err := service.GetSubscribers(ctx, filter)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestSubscribeService_UpdateSubscriber(t *testing.T) {
	ctx := context.Background()

	t.Run("should update subscriber successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		existingSubscriber := &entity.Subscriber{
			ID:        1,
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
			IsActive:  true,
		}

		updateReq := &dto.UpdateSubscriberRequest{
			CallName: "Updated User",
			Plan:     "hotspot",
			Price:    30000.0,
			IsActive: false,
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(1)).Return(existingSubscriber, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.Subscriber")).Return(nil)

		// When
		response, err := service.UpdateSubscriber(ctx, 1, updateReq)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Updated User", response.CallName)
		assert.Equal(t, "hotspot", response.Plan)
		assert.Equal(t, 30000.0, response.Price)
		assert.False(t, response.IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when plan is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		existingSubscriber := &entity.Subscriber{
			ID:        1,
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
			IsActive:  true,
		}

		updateReq := &dto.UpdateSubscriberRequest{
			Plan: "invalid_plan",
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(1)).Return(existingSubscriber, nil)

		// When
		response, err := service.UpdateSubscriber(ctx, 1, updateReq)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "invalid plan, must be 'pppoe' or 'hotspot'", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when subscriber not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		updateReq := &dto.UpdateSubscriberRequest{
			CallName: "Updated User",
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.UpdateSubscriber(ctx, 999, updateReq)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "subscriber not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestSubscribeService_DeleteSubscriber(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete subscriber successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		subscriber := &entity.Subscriber{
			ID:        1,
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
			IsActive:  true,
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(1)).Return(subscriber, nil)
		mockRepo.On("Delete", ctx, uint(1)).Return(nil)

		// When
		err := service.DeleteSubscriber(ctx, 1)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when subscriber not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// When
		err := service.DeleteSubscriber(ctx, 999)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "subscriber not found", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when delete fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockSubscribeRepository{}
		logger := testutil.NewSilentLogger()
		service := NewSubscribeService(mockRepo, logger)

		subscriber := &entity.Subscriber{
			ID:        1,
			Username:  "testuser",
			CallName:  "Test User",
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
			IsActive:  true,
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(1)).Return(subscriber, nil)
		mockRepo.On("Delete", ctx, uint(1)).Return(errors.New("delete failed"))

		// When
		err := service.DeleteSubscriber(ctx, 1)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "delete failed", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
