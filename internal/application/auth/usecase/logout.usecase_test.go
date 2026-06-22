package usecase

import (
	"context"
	"errors"
	"testing"

	authDto "github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupLogoutUseCase() (
	LogoutUseCase,
	*testutil.MockAuthSessionService,
) {

	mockSessionSvc := &testutil.MockAuthSessionService{}

	useCase := NewLogoutUseCase(
		mockSessionSvc,
	)

	return useCase, mockSessionSvc
}

func TestLogoutUseCase_Execute(t *testing.T) {

	t.Run("should logout successfully", func(t *testing.T) {

		// Setup
		useCase, mockSessionSvc := setupLogoutUseCase()

		req := &authDto.LogoutRequest{
			RefreshToken: "refresh-token",
		}

		mockSessionSvc.On(
			"LogoutByRefreshToken",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			"refresh-token",
		).Return(nil)

		// When
		err := useCase.Execute(
			context.Background(),
			req,
		)

		// Then
		assert.NoError(
			t,
			err,
		)

		mockSessionSvc.AssertExpectations(t)
	})

	t.Run("should return error when logout failed", func(t *testing.T) {

		// Setup
		useCase, mockSessionSvc := setupLogoutUseCase()

		req := &authDto.LogoutRequest{
			RefreshToken: "invalid-refresh-token",
		}

		mockSessionSvc.On(
			"LogoutByRefreshToken",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			"invalid-refresh-token",
		).Return(
			errors.New("session not found"),
		)

		// When
		err := useCase.Execute(
			context.Background(),
			req,
		)

		// Then
		assert.Error(
			t,
			err,
		)

		assert.Equal(
			t,
			"session not found",
			err.Error(),
		)

		mockSessionSvc.AssertExpectations(t)
	})

	t.Run("should pass refresh token correctly", func(t *testing.T) {

		// Setup
		useCase, mockSessionSvc := setupLogoutUseCase()

		req := &authDto.LogoutRequest{
			RefreshToken: "test-refresh-token",
		}

		mockSessionSvc.On(
			"LogoutByRefreshToken",
			mock.Anything,
			mock.MatchedBy(func(token string) bool {
				return token == "test-refresh-token"
			}),
		).Return(nil)

		// When
		err := useCase.Execute(
			context.Background(),
			req,
		)

		// Then
		assert.NoError(
			t,
			err,
		)

		mockSessionSvc.AssertExpectations(t)
	})
}
