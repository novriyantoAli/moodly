package service

import (
	"context"
	"errors"
	"testing"

	"github.com/novriyantoAli/moodly/internal/application/user/dto"
	"github.com/novriyantoAli/moodly/internal/application/user/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("should create user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		req := &dto.CreateUserRequest{
			Email:    "test@example.com",
			FullName: "Test User",
		}

		// Mock expectations
		mockRepo.On("EmailExists", ctx, req.Email).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(1).(*entity.User)
			user.ID = 1
		})

		// When
		response, err := service.CreateUser(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, req.FullName, response.FullName)
		assert.Equal(t, req.Email, response.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when email already exists", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		req := &dto.CreateUserRequest{
			Email:    "existing@example.com",
			FullName: "Existing User",
		}

		// Mock expectations
		mockRepo.On("EmailExists", ctx, req.Email).Return(true, nil)

		// When
		response, err := service.CreateUser(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "email already exists", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		req := &dto.CreateUserRequest{
			Email:    "test@example.com",
			FullName: "Test User",
		}

		// Mock expectations
		mockRepo.On("EmailExists", ctx, req.Email).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(errors.New("database error"))

		// When
		response, err := service.CreateUser(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get user by ID successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		user := &entity.User{
			ID:       1,
			Email:    "test@example.com",
			FullName: "Test User",
			Level:    "user",
			IsActive: true,
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)

		// When
		response, err := service.GetUserByID(ctx, 1)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Email, response.Email)
		assert.Equal(t, user.FullName, response.FullName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetUserByID(ctx, 999)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("should get user by email successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		user := &entity.User{
			ID:       1,
			Email:    "test@example.com",
			FullName: "Test User",
			Level:    "user",
			IsActive: true,
		}

		// Mock expectations
		mockRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

		// When
		response, err := service.GetUserByEmail(ctx, "test@example.com")

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, user.Email, response.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when email not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		// Mock expectations
		mockRepo.On("GetByEmail", ctx, "nonexistent@example.com").Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetUserByEmail(ctx, "nonexistent@example.com")

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("should get users with pagination", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		users := []entity.User{
			{
				ID:       1,
				Email:    "user1@example.com",
				FullName: "User 1",
				Level:    "user",
				IsActive: true,
			},
			{
				ID:       2,
				Email:    "user2@example.com",
				FullName: "User 2",
				Level:    "user",
				IsActive: true,
			},
		}

		filter := &dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", ctx, filter).Return(users, int64(2), nil)

		// When
		response, err := service.GetUsers(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, int64(2), response.TotalCount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		filter := &dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", ctx, filter).Return(nil, int64(0), errors.New("database error"))

		// When
		response, err := service.GetUsers(ctx, filter)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("should update user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		existingUser := &entity.User{
			ID:       1,
			Email:    "test@example.com",
			FullName: "Test User",
			Level:    "user",
			IsActive: true,
		}

		updateReq := &dto.UpdateUserRequest{
			FullName: "Updated User",
			Level:    "admin",
			IsActive: false,
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(1)).Return(existingUser, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(nil)

		// When
		response, err := service.UpdateUser(ctx, 1, updateReq)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Updated User", response.FullName)
		assert.Equal(t, "admin", response.Level)
		assert.False(t, response.IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		updateReq := &dto.UpdateUserRequest{
			FullName: "Updated User",
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.UpdateUser(ctx, 999, updateReq)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		user := &entity.User{
			ID:       1,
			Email:    "test@example.com",
			FullName: "Test User",
			Level:    "user",
			IsActive: true,
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
		mockRepo.On("Delete", ctx, uint(1)).Return(nil)

		// When
		err := service.DeleteUser(ctx, 1)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, logger)

		// Mock expectations
		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// When
		err := service.DeleteUser(ctx, 999)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
