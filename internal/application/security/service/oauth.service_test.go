package service

import (
	"context"
	"errors"
	"testing"

	"github.com/novriyantoAli/moodly/internal/config"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/idtoken"
)

func setupOAuthService() (
	OAuthService,
	*testutil.MockGoogleTokenValidator,
) {

	logger := testutil.NewSilentLogger()

	mockValidator := &testutil.MockGoogleTokenValidator{}

	cfg := &config.Config{
		OAuth: config.OAuthConfig{
			// Google: config.GoogleConfig{
			// 	ClientID: "google-client-id",
			// },
			Google: struct {
				ClientID     string "mapstructure:\"client_id\""
				ClientSecret string "mapstructure:\"client_secret\""
				RedirectURL  string "mapstructure:\"redirect_url\""
			}{
				ClientID:     "google-client-id",
				ClientSecret: "google-client-secret",
				RedirectURL:  "http://localhost:8080/auth/google/callback",
			},
		},
	}

	svc := NewOAuthService(
		mockValidator,
		cfg,
		logger,
	)

	return svc, mockValidator
}
func TestOAuthService_VerifyGoogleToken(t *testing.T) {

	t.Run("should verify google token successfully", func(t *testing.T) {

		svc, mockValidator := setupOAuthService()

		payload := &idtoken.Payload{
			Subject: "google-user-id",
			Claims: map[string]interface{}{
				"email":   "user@gmail.com",
				"name":    "John Doe",
				"picture": "https://image.jpg",
			},
		}

		mockValidator.On(
			"Validate",
			mock.Anything,
			"valid-token",
			"google-client-id",
		).Return(
			payload,
			nil,
		)

		resp, err := svc.VerifyGoogleToken(
			context.Background(),
			"valid-token",
		)

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Equal(
			t,
			"google-user-id",
			resp.Subject,
		)

		assert.Equal(
			t,
			"user@gmail.com",
			resp.Email,
		)

		assert.Equal(
			t,
			"John Doe",
			resp.Name,
		)

		assert.Equal(
			t,
			"https://image.jpg",
			resp.Picture,
		)

		mockValidator.AssertExpectations(t)
	})

	t.Run("should return error when token invalid", func(t *testing.T) {

		svc, mockValidator := setupOAuthService()

		mockValidator.On(
			"Validate",
			mock.Anything,
			"invalid-token",
			"google-client-id",
		).Return(
			nil,
			errors.New("token invalid"),
		)

		resp, err := svc.VerifyGoogleToken(
			context.Background(),
			"invalid-token",
		)

		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.Contains(
			t,
			err.Error(),
			"invalid google token",
		)

		mockValidator.AssertExpectations(t)
	})

	t.Run("should handle missing optional claims", func(t *testing.T) {

		svc, mockValidator := setupOAuthService()

		payload := &idtoken.Payload{
			Subject: "google-user-id",
			Claims:  map[string]interface{}{},
		}

		mockValidator.On(
			"Validate",
			mock.Anything,
			"valid-token",
			"google-client-id",
		).Return(
			payload,
			nil,
		)

		resp, err := svc.VerifyGoogleToken(
			context.Background(),
			"valid-token",
		)

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Equal(
			t,
			"google-user-id",
			resp.Subject,
		)

		assert.Empty(
			t,
			resp.Email,
		)

		assert.Empty(
			t,
			resp.Name,
		)

		assert.Empty(
			t,
			resp.Picture,
		)

		mockValidator.AssertExpectations(t)
	})
}
