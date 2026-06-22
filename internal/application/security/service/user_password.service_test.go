package service

import (
	"context"
	"errors"
	"testing"
	"time"

	dto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	entity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"golang.org/x/crypto/bcrypt"
)

func setupUserPasswordServiceWithMocks() (
	UserPasswordService,
	*testutil.MockUserPasswordRepository,
) {

	mockRepo := &testutil.MockUserPasswordRepository{}
	logger := testutil.NewSilentLogger()

	service := NewUserPasswordService(
		mockRepo,
		logger,
	)

	return service, mockRepo
}

func setupUserPasswordService(
	mockRepo *testutil.MockUserPasswordRepository,
) UserPasswordService {

	logger := testutil.NewSilentLogger()

	return NewUserPasswordService(
		mockRepo,
		logger,
	)
}

func hashPassword(
	password string,
) string {

	hash, _ := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	return string(hash)
}

func TestUserPasswordService_SetPassword(t *testing.T) {
	ctx := context.Background()
	t.Run(
		"should create password when not exists",
		func(t *testing.T) {
			service, mockRepo := setupUserPasswordServiceWithMocks()
			req := &dto.SetPasswordRequest{
				UserID:   1,
				Username: "john",
				Password: "Password123",
			}

			mockRepo.On("GetByUserID", ctx, req.UserID).Return(nil, nil)

			mockRepo.On("Create", ctx, mock.MatchedBy(func(p *entity.UserPassword) bool {
				return p.UserID == req.UserID && p.Username == req.Username
			})).Return(nil)

			err := service.SetPassword(
				ctx,
				req,
			)

			assert.NoError(t, err)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should update existing password",
		func(t *testing.T) {

			service, mockRepo :=
				setupUserPasswordServiceWithMocks()

			req := &dto.SetPasswordRequest{
				UserID:   1,
				Username: "john",
				Password: "NewPassword123",
			}

			existing := &entity.UserPassword{
				UserID:       1,
				Username:     "john",
				PasswordHash: hashPassword("old"),
			}

			mockRepo.
				On(
					"GetByUserID",
					ctx,
					req.UserID,
				).
				Return(existing, nil)

			mockRepo.
				On(
					"UpdatePasswordHash",
					ctx,
					req.UserID,
					mock.AnythingOfType("string"),
				).
				Return(nil)

			err := service.SetPassword(
				ctx,
				req,
			)

			assert.NoError(t, err)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestUserPasswordService_VerifyPassword(t *testing.T) {

	ctx := context.Background()

	t.Run(
		"should verify password successfully",
		func(t *testing.T) {

			service, mockRepo :=
				setupUserPasswordServiceWithMocks()

			req := &dto.VerifyPasswordRequest{
				Username: "john",
				Password: "Password123",
			}

			userPassword := &entity.UserPassword{
				UserID:        1,
				Username:      "john",
				PasswordHash:  hashPassword("Password123"),
				FailedAttempt: 2,
				LockedUntil:   nil,
			}

			mockRepo.
				On(
					"GetByUsername",
					ctx,
					req.Username,
				).
				Return(userPassword, nil)

			mockRepo.
				On(
					"ResetFailedAttempt",
					ctx,
					uint(1),
				).
				Return(nil)

			mockRepo.
				On(
					"UpdateLastLogin",
					ctx,
					uint(1),
					mock.AnythingOfType("time.Time"),
				).
				Return(nil)

			isValid, err :=
				service.VerifyPassword(
					ctx,
					req,
				)

			assert.NoError(t, err)
			assert.True(t, isValid)

			mockRepo.AssertExpectations(t)
		},
	)

	t.Run(
		"should increment failed attempt when password invalid",
		func(t *testing.T) {

			service, mockRepo :=
				setupUserPasswordServiceWithMocks()

			req := &dto.VerifyPasswordRequest{
				Username: "john",
				Password: "wrong",
			}

			userPassword := &entity.UserPassword{
				UserID:        1,
				Username:      "john",
				PasswordHash:  hashPassword("Password123"),
				FailedAttempt: 0,
			}

			mockRepo.
				On(
					"GetByUsername",
					ctx,
					req.Username,
				).
				Return(userPassword, nil)

			mockRepo.
				On(
					"IncrementFailedAttempt",
					ctx,
					uint(1),
				).
				Return(nil)

			isValid, err :=
				service.VerifyPassword(
					ctx,
					req,
				)

			assert.NoError(t, err)
			assert.False(t, isValid)

			mockRepo.AssertExpectations(t)
		})

	t.Run(
		"should lock account after 5 failed attempts",
		func(t *testing.T) {

			service, mockRepo :=
				setupUserPasswordServiceWithMocks()

			req := &dto.VerifyPasswordRequest{
				Username: "john",
				Password: "wrong",
			}

			userPassword := &entity.UserPassword{
				UserID:        1,
				Username:      "john",
				PasswordHash:  hashPassword("Password123"),
				FailedAttempt: 4,
			}

			mockRepo.
				On(
					"GetByUsername",
					ctx,
					req.Username,
				).
				Return(userPassword, nil)

			mockRepo.
				On(
					"IncrementFailedAttempt",
					ctx,
					uint(1),
				).
				Return(nil)

			mockRepo.
				On(
					"LockAccount",
					ctx,
					uint(1),
					15*time.Minute,
				).
				Return(nil)

			isValid, err :=
				service.VerifyPassword(
					ctx,
					req,
				)

			assert.NoError(t, err)
			assert.False(t, isValid)

			mockRepo.AssertExpectations(t)
		},
	)

}

func TestUserPasswordService_ChangePassword(
	t *testing.T,
) {

	ctx := context.Background()

	t.Run(
		"should change password successfully",
		func(t *testing.T) {

			service, mockRepo :=
				setupUserPasswordServiceWithMocks()

			req := &dto.ChangePasswordRequest{
				UserID:          1,
				CurrentPassword: "OldPassword",
				NewPassword:     "NewPassword",
			}

			userPassword := &entity.UserPassword{
				UserID:       1,
				Username:     "john",
				PasswordHash: hashPassword("OldPassword"),
			}

			mockRepo.
				On(
					"GetByUserID",
					ctx,
					req.UserID,
				).
				Return(userPassword, nil)

			mockRepo.
				On(
					"UpdatePasswordHash",
					ctx,
					req.UserID,
					mock.AnythingOfType("string"),
				).
				Return(nil)

			err := service.ChangePassword(
				ctx,
				req,
			)

			assert.NoError(t, err)

			mockRepo.AssertExpectations(t)
		},
	)
}

func TestUserPasswordService_GetPasswordInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("should get password info successfully", func(t *testing.T) {
		// Setup
		service, mockRepo := setupUserPasswordServiceWithMocks()

		userID := uint(1)

		password := &entity.UserPassword{
			UserID:        userID,
			Username:      "john",
			PasswordHash:  hashPassword("Password123"),
			FailedAttempt: 1,
			LockedUntil:   nil,
		}

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(password, nil)

		// When
		response, err := service.GetPasswordInfo(
			ctx,
			userID,
		)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.UserID)
		assert.Equal(t, "john", response.Username)
		assert.Equal(t, 1, response.FailedAttempt)
		assert.False(t, response.IsLocked)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return nil when password record not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPasswordRepository{}
		service := setupUserPasswordService(mockRepo)

		userID := uint(999)

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(nil, nil)

		// When
		response, err := service.GetPasswordInfo(
			ctx,
			userID,
		)

		// Then
		assert.NoError(t, err)
		assert.Nil(t, response)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should identify locked account correctly", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPasswordRepository{}
		service := setupUserPasswordService(mockRepo)

		userID := uint(1)

		lockedUntil := time.Now().Add(
			10 * time.Minute,
		)

		password := &entity.UserPassword{
			UserID:        userID,
			Username:      "john",
			PasswordHash:  hashPassword("Password123"),
			FailedAttempt: 5,
			LockedUntil:   &lockedUntil,
		}

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(password, nil)

		// When
		response, err := service.GetPasswordInfo(
			ctx,
			userID,
		)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.UserID)
		assert.Equal(t, "john", response.Username)
		assert.Equal(t, 5, response.FailedAttempt)
		assert.True(t, response.IsLocked)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should identify expired lock as unlocked", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPasswordRepository{}
		service := setupUserPasswordService(mockRepo)

		userID := uint(1)

		expiredLock := time.Now().Add(
			-10 * time.Minute,
		)

		password := &entity.UserPassword{
			UserID:        userID,
			Username:      "john",
			PasswordHash:  hashPassword("Password123"),
			FailedAttempt: 5,
			LockedUntil:   &expiredLock,
		}

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(password, nil)

		// When
		response, err := service.GetPasswordInfo(
			ctx,
			userID,
		)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.UserID)
		assert.False(t, response.IsLocked)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPasswordRepository{}
		service := setupUserPasswordService(mockRepo)

		userID := uint(1)

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(nil, errors.New("database error"))

		// When
		response, err := service.GetPasswordInfo(
			ctx,
			userID,
		)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(
			t,
			"database error",
			err.Error(),
		)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserPasswordService_IsAccountLocked(t *testing.T) {
	ctx := context.Background()

	t.Run("should return true when account is locked", func(t *testing.T) {
		// Setup
		service, mockRepo := setupUserPasswordServiceWithMocks()

		userID := uint(1)

		lockedUntil := time.Now().Add(
			10 * time.Minute,
		)

		password := &entity.UserPassword{
			UserID:       userID,
			Username:     "john",
			PasswordHash: hashPassword("Password123"),
			LockedUntil:  &lockedUntil,
		}

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(password, nil)

		// When
		isLocked, err := service.IsAccountLocked(
			ctx,
			userID,
		)

		// Then
		assert.NoError(t, err)
		assert.True(t, isLocked)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when account is not locked", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPasswordRepository{}
		service := setupUserPasswordService(mockRepo)

		userID := uint(1)

		password := &entity.UserPassword{
			UserID:       userID,
			Username:     "john",
			PasswordHash: hashPassword("Password123"),
			LockedUntil:  nil,
		}

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(password, nil)

		// When
		isLocked, err := service.IsAccountLocked(
			ctx,
			userID,
		)

		// Then
		assert.NoError(t, err)
		assert.False(t, isLocked)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when lock already expired", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPasswordRepository{}
		service := setupUserPasswordService(mockRepo)

		userID := uint(1)

		expiredLock := time.Now().Add(
			-10 * time.Minute,
		)

		password := &entity.UserPassword{
			UserID:       userID,
			Username:     "john",
			PasswordHash: hashPassword("Password123"),
			LockedUntil:  &expiredLock,
		}

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(password, nil)

		// When
		isLocked, err := service.IsAccountLocked(
			ctx,
			userID,
		)

		// Then
		assert.NoError(t, err)
		assert.False(t, isLocked)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when password record not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPasswordRepository{}
		service := setupUserPasswordService(mockRepo)

		userID := uint(999)

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(nil, nil)

		// When
		isLocked, err := service.IsAccountLocked(
			ctx,
			userID,
		)

		// Then
		assert.NoError(t, err)
		assert.False(t, isLocked)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserPasswordRepository{}
		service := setupUserPasswordService(mockRepo)

		userID := uint(1)

		// Mock expectations
		mockRepo.
			On("GetByUserID", ctx, userID).
			Return(nil, errors.New("database error"))

		// When
		isLocked, err := service.IsAccountLocked(
			ctx,
			userID,
		)

		// Then
		assert.Error(t, err)
		assert.False(t, isLocked)
		assert.Equal(
			t,
			"database error",
			err.Error(),
		)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserPasswordService_DeletePassword(
	t *testing.T,
) {

	ctx := context.Background()

	service, mockRepo :=
		setupUserPasswordServiceWithMocks()

	mockRepo.
		On(
			"Delete",
			ctx,
			uint(1),
		).
		Return(nil)

	err := service.DeletePassword(
		ctx,
		1,
	)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
