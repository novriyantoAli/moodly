package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	dto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	entity "github.com/novriyantoAli/moodly/internal/application/security/entity"

	testutil "github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserPINServiceWithMocks() (UserPINService, *testutil.MockUserPINRepository) {
	mockRepo := &testutil.MockUserPINRepository{}
	logger := testutil.NewSilentLogger()
	service := NewUserPINService(mockRepo, logger)
	return service, mockRepo
}

func setupUserPINService(mockRepo *testutil.MockUserPINRepository) UserPINService {
	logger := testutil.NewSilentLogger()
	service := NewUserPINService(mockRepo, logger)
	return service
}

func hashPIN(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return fmt.Sprintf("%x", hash)
}

func TestUserPINService_SetPIN(t *testing.T) {
	ctx := context.Background()

	t.Run("should create new security record when none exists", func(t *testing.T) {
		// Setup
		service, mockRepo := setupUserPINServiceWithMocks()

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "123456",
		}

		expectedPinHash := hashPIN(req.PIN)

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(nil, nil)
		mockRepo.On("Create", ctx, mock.MatchedBy(func(s *entity.UserPIN) bool {
			return s.UserID == req.UserID &&
				s.PinHash == expectedPinHash &&
				s.FailedAttempt == 0 &&
				s.LockedUntil == nil
		})).Return(nil)

		// When
		err := service.SetPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should update existing PIN", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "654321",
		}

		existingSecurity := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 0,
			LockedUntil:   nil,
		}

		expectedPinHash := hashPIN(req.PIN)

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(existingSecurity, nil)
		mockRepo.On("UpdatePIN", ctx, req.UserID, expectedPinHash).Return(nil)

		// When
		err := service.SetPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails to get security", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "123456",
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(nil, errors.New("database error"))

		// When
		err := service.SetPIN(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails to create", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "123456",
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(nil, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.UserPIN")).Return(errors.New("create error"))

		// When
		err := service.SetPIN(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "create error", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails to update", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "654321",
		}

		existingSecurity := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 0,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(existingSecurity, nil)
		mockRepo.On("UpdatePIN", ctx, req.UserID, mock.AnythingOfType("string")).Return(errors.New("update error"))

		// When
		err := service.SetPIN(ctx, req)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "update error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPINService_VerifyPIN(t *testing.T) {
	ctx := context.Background()

	t.Run("should verify PIN successfully and reset failed attempts", func(t *testing.T) {
		// Setup
		service, mockRepo := setupUserPINServiceWithMocks()

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "123456",
		}

		security := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN(req.PIN),
			FailedAttempt: 2,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(security, nil)
		mockRepo.On("ResetFailedAttempt", ctx, req.UserID).Return(nil)

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.True(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when PIN is incorrect", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "654321",
		}

		security := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 0,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(security, nil)
		mockRepo.On("IncrementFailedAttempt", ctx, req.UserID).Return(nil)

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.False(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when security record not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.VerifyPINRequest{
			UserID: 999,
			PIN:    "123456",
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(nil, nil)

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.False(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when account is locked", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "123456",
		}

		lockedUntil := time.Now().Add(10 * time.Minute)
		security := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN(req.PIN),
			FailedAttempt: 3,
			LockedUntil:   &lockedUntil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(security, nil)

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.False(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should lock account after 3 failed attempts", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "wrongpin",
		}

		security := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 2,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(security, nil)
		mockRepo.On("IncrementFailedAttempt", ctx, req.UserID).Return(nil)
		mockRepo.On("LockAccount", ctx, req.UserID, 15*time.Minute).Return(nil)

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.False(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "123456",
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(nil, errors.New("database error"))

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.Error(t, err)
		assert.False(t, isValid)
		assert.Equal(t, "database error", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should continue even if reset failed attempt fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "123456",
		}

		security := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN(req.PIN),
			FailedAttempt: 1,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(security, nil)
		mockRepo.On("ResetFailedAttempt", ctx, req.UserID).Return(errors.New("reset error"))

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.True(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should continue even if increment failed attempt fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "wrongpin",
		}

		security := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 0,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(security, nil)
		mockRepo.On("IncrementFailedAttempt", ctx, req.UserID).Return(errors.New("increment error"))

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.False(t, isValid)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should continue even if lock account fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "wrongpin",
		}

		security := &entity.UserPIN{
			UserID:        1,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 2,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, req.UserID).Return(security, nil)
		mockRepo.On("IncrementFailedAttempt", ctx, req.UserID).Return(nil)
		mockRepo.On("LockAccount", ctx, req.UserID, 15*time.Minute).Return(errors.New("lock error"))

		// When
		isValid, err := service.VerifyPIN(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.False(t, isValid)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPINService_GetSecurity(t *testing.T) {
	ctx := context.Background()

	t.Run("should get security info successfully", func(t *testing.T) {
		// Setup
		service, mockRepo := setupUserPINServiceWithMocks()

		userID := uint(1)
		security := &entity.UserPIN{
			UserID:        userID,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 1,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(security, nil)

		// When
		response, err := service.GetSecurity(ctx, userID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.UserID)
		assert.Equal(t, 1, response.FailedAttempt)
		assert.False(t, response.IsLocked)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return default security info when record not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		userID := uint(999)

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(nil, nil)

		// When
		response, err := service.GetSecurity(ctx, userID)

		// Then
		assert.NoError(t, err)
		assert.Nil(t, response)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should correctly identify locked account", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		userID := uint(1)
		lockedUntil := time.Now().Add(10 * time.Minute)
		security := &entity.UserPIN{
			UserID:        userID,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 3,
			LockedUntil:   &lockedUntil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(security, nil)

		// When
		response, err := service.GetSecurity(ctx, userID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.UserID)
		assert.Equal(t, 3, response.FailedAttempt)
		assert.True(t, response.IsLocked)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should correctly identify unlocked account when lock expired", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		userID := uint(1)
		lockedUntil := time.Now().Add(-10 * time.Minute) // expired lock
		security := &entity.UserPIN{
			UserID:        userID,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 3,
			LockedUntil:   &lockedUntil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(security, nil)

		// When
		response, err := service.GetSecurity(ctx, userID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.UserID)
		assert.Equal(t, 3, response.FailedAttempt)
		assert.False(t, response.IsLocked)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		userID := uint(1)

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("database error"))

		// When
		response, err := service.GetSecurity(ctx, userID)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "database error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPINService_IsAccountLocked(t *testing.T) {
	ctx := context.Background()

	t.Run("should return true when account is locked", func(t *testing.T) {
		// Setup
		service, mockRepo := setupUserPINServiceWithMocks()

		userID := uint(1)
		lockedUntil := time.Now().Add(10 * time.Minute)
		security := &entity.UserPIN{
			UserID:        userID,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 3,
			LockedUntil:   &lockedUntil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(security, nil)

		// When
		isLocked, err := service.IsAccountLocked(ctx, userID)

		// Then
		assert.NoError(t, err)
		assert.True(t, isLocked)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when account is not locked", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		userID := uint(1)
		security := &entity.UserPIN{
			UserID:        userID,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 1,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(security, nil)

		// When
		isLocked, err := service.IsAccountLocked(ctx, userID)

		// Then
		assert.NoError(t, err)
		assert.False(t, isLocked)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when lock has expired", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		userID := uint(1)
		lockedUntil := time.Now().Add(-10 * time.Minute) // expired
		security := &entity.UserPIN{
			UserID:        userID,
			PinHash:       hashPIN("123456"),
			FailedAttempt: 3,
			LockedUntil:   &lockedUntil,
		}

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(security, nil)

		// When
		isLocked, err := service.IsAccountLocked(ctx, userID)

		// Then
		assert.NoError(t, err)
		assert.False(t, isLocked)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when security record not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		userID := uint(999)

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(nil, nil)

		// When
		isLocked, err := service.IsAccountLocked(ctx, userID)

		// Then
		assert.NoError(t, err)
		assert.False(t, isLocked)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPINRepository{}
		service := setupUserPINService(mockRepo)

		userID := uint(1)

		// Mock expectations
		mockRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("database error"))

		// When
		isLocked, err := service.IsAccountLocked(ctx, userID)

		// Then
		assert.Error(t, err)
		assert.False(t, isLocked)
		assert.Equal(t, "database error", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
