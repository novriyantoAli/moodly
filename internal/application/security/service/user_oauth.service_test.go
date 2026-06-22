package service

import (
	"context"
	"errors"
	"testing"

	securityEntity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserOAuthService() (
	UserOAuthService,
	*testutil.MockUserOAuthRepository,
) {

	mockRepo := &testutil.MockUserOAuthRepository{}
	logger := testutil.NewSilentLogger()

	svc := NewUserOAuthService(
		mockRepo,
		logger,
	)

	return svc, mockRepo
}

func TestUserOAuthService_GetByProviderAndProviderUserID(t *testing.T) {

	t.Run("should get oauth successfully", func(t *testing.T) {

		// Setup
		svc, mockRepo := setupUserOAuthService()

		expected := &securityEntity.UserOAuth{
			UserID:         1,
			Provider:       "google",
			ProviderUserID: "google-user-id",
		}

		mockRepo.On(
			"GetByProviderAndUserID",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			"google",
			"google-user-id",
		).Return(
			expected,
			nil,
		)

		// When
		result, err := svc.GetByProviderAndProviderUserID(
			context.Background(),
			"google",
			"google-user-id",
		)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(
			t,
			uint(1),
			result.UserID,
		)

		assert.Equal(
			t,
			"google",
			result.Provider,
		)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when provider empty", func(t *testing.T) {

		// Setup
		svc, _ := setupUserOAuthService()

		// When
		result, err := svc.GetByProviderAndProviderUserID(
			context.Background(),
			"",
			"google-user-id",
		)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)

		assert.Equal(
			t,
			"provider is required",
			err.Error(),
		)
	})

	t.Run("should return error when provider user id empty", func(t *testing.T) {

		// Setup
		svc, _ := setupUserOAuthService()

		// When
		result, err := svc.GetByProviderAndProviderUserID(
			context.Background(),
			"google",
			"",
		)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)

		assert.Equal(
			t,
			"provider user id is required",
			err.Error(),
		)
	})

	t.Run("should return repository error", func(t *testing.T) {

		// Setup
		svc, mockRepo := setupUserOAuthService()

		mockRepo.On(
			"GetByProviderAndUserID",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			"google",
			"google-user-id",
		).Return(
			nil,
			errors.New("oauth not found"),
		)

		// When
		result, err := svc.GetByProviderAndProviderUserID(
			context.Background(),
			"google",
			"google-user-id",
		)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)

		assert.Equal(
			t,
			"oauth not found",
			err.Error(),
		)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserOAuthService_Create(t *testing.T) {

	t.Run("should create oauth successfully", func(t *testing.T) {

		// Setup
		svc, mockRepo := setupUserOAuthService()

		oauth := &securityEntity.UserOAuth{
			UserID:         1,
			Provider:       "google",
			ProviderUserID: "google-user-id",
		}

		mockRepo.On(
			"Create",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(o *securityEntity.UserOAuth) bool {
				return o.UserID == 1 &&
					o.Provider == "google" &&
					o.ProviderUserID == "google-user-id"
			}),
		).Return(nil)

		// When
		err := svc.Create(
			context.Background(),
			oauth,
		)

		// Then
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when oauth nil", func(t *testing.T) {

		// Setup
		svc, _ := setupUserOAuthService()

		// When
		err := svc.Create(
			context.Background(),
			nil,
		)

		// Then
		assert.Error(t, err)

		assert.Equal(
			t,
			"oauth data is required",
			err.Error(),
		)
	})

	t.Run("should return error when user id empty", func(t *testing.T) {

		// Setup
		svc, _ := setupUserOAuthService()

		oauth := &securityEntity.UserOAuth{
			UserID:         0,
			Provider:       "google",
			ProviderUserID: "google-user-id",
		}

		// When
		err := svc.Create(
			context.Background(),
			oauth,
		)

		// Then
		assert.Error(t, err)

		assert.Equal(
			t,
			"user id is required",
			err.Error(),
		)
	})

	t.Run("should return error when provider empty", func(t *testing.T) {

		// Setup
		svc, _ := setupUserOAuthService()

		oauth := &securityEntity.UserOAuth{
			UserID:         1,
			Provider:       "",
			ProviderUserID: "google-user-id",
		}

		// When
		err := svc.Create(
			context.Background(),
			oauth,
		)

		// Then
		assert.Error(t, err)

		assert.Equal(
			t,
			"provider is required",
			err.Error(),
		)
	})

	t.Run("should return error when provider user id empty", func(t *testing.T) {

		// Setup
		svc, _ := setupUserOAuthService()

		oauth := &securityEntity.UserOAuth{
			UserID:         1,
			Provider:       "google",
			ProviderUserID: "",
		}

		// When
		err := svc.Create(
			context.Background(),
			oauth,
		)

		// Then
		assert.Error(t, err)

		assert.Equal(
			t,
			"provider user id is required",
			err.Error(),
		)
	})

	t.Run("should return repository error", func(t *testing.T) {

		// Setup
		svc, mockRepo := setupUserOAuthService()

		oauth := &securityEntity.UserOAuth{
			UserID:         1,
			Provider:       "google",
			ProviderUserID: "google-user-id",
		}

		mockRepo.On(
			"Create",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.Anything,
		).Return(
			errors.New("database error"),
		)

		// When
		err := svc.Create(
			context.Background(),
			oauth,
		)

		// Then
		assert.Error(t, err)

		assert.Equal(
			t,
			"database error",
			err.Error(),
		)

		mockRepo.AssertExpectations(t)
	})
}
