package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authDto "github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupGoogleLoginHandler() (
	*GoogleLoginHandler,
	*testutil.MockGoogleLoginUseCase,
) {

	gin.SetMode(gin.TestMode)

	mockUseCase := &testutil.MockGoogleLoginUseCase{}
	logger := testutil.NewSilentLogger()

	handler := NewGoogleLoginHandler(
		mockUseCase,
		logger,
	)

	return handler, mockUseCase
}

func TestGoogleLoginHandler_GoogleLogin(t *testing.T) {

	t.Run("should google login successfully", func(t *testing.T) {

		// Setup
		handler, mockUseCase := setupGoogleLoginHandler()

		req := authDto.GoogleLoginRequest{
			IDToken: "google-id-token",
		}

		response := &authDto.GoogleLoginResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiredAt:    9999999999,
		}

		mockUseCase.On(
			"Execute",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(r *authDto.GoogleLoginRequest) bool {
				return r.IDToken == req.IDToken
			}),
		).Return(
			response,
			nil,
		)

		reqBody, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/google/login",
			bytes.NewBuffer(reqBody),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.GoogleLogin(ctx)

		// Then
		assert.Equal(
			t,
			http.StatusOK,
			w.Code,
		)

		mockUseCase.AssertExpectations(t)

		var result map[string]interface{}

		_ = json.Unmarshal(
			w.Body.Bytes(),
			&result,
		)

		assert.Equal(
			t,
			true,
			result["success"],
		)

		assert.Equal(
			t,
			"google login success",
			result["message"],
		)

		assert.Contains(
			t,
			result,
			"data",
		)
	})

	t.Run("should return bad request for invalid json", func(t *testing.T) {

		// Setup
		handler, _ := setupGoogleLoginHandler()

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/google/login",
			bytes.NewBuffer([]byte("invalid json")),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.GoogleLogin(ctx)

		// Then
		assert.Equal(
			t,
			http.StatusBadRequest,
			w.Code,
		)

		var result map[string]interface{}

		_ = json.Unmarshal(
			w.Body.Bytes(),
			&result,
		)

		assert.Equal(
			t,
			false,
			result["success"],
		)

		assert.Equal(
			t,
			"invalid request",
			result["message"],
		)
	})

	t.Run("should return unauthorized when google login failed", func(t *testing.T) {

		// Setup
		handler, mockUseCase := setupGoogleLoginHandler()

		req := authDto.GoogleLoginRequest{
			IDToken: "invalid-token",
		}

		mockUseCase.On(
			"Execute",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(r *authDto.GoogleLoginRequest) bool {
				return r.IDToken == req.IDToken
			}),
		).Return(
			nil,
			errors.New("invalid google token"),
		)

		reqBody, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/google/login",
			bytes.NewBuffer(reqBody),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.GoogleLogin(ctx)

		// Then
		assert.Equal(
			t,
			http.StatusUnauthorized,
			w.Code,
		)

		mockUseCase.AssertExpectations(t)

		var result map[string]interface{}

		_ = json.Unmarshal(
			w.Body.Bytes(),
			&result,
		)

		assert.Equal(
			t,
			false,
			result["success"],
		)

		assert.Equal(
			t,
			"invalid google token",
			result["message"],
		)
	})

	t.Run("should pass client ip and user agent to usecase", func(t *testing.T) {

		// Setup
		handler, mockUseCase := setupGoogleLoginHandler()

		req := authDto.GoogleLoginRequest{
			IDToken: "google-id-token",
		}

		response := &authDto.GoogleLoginResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiredAt:    9999999999,
		}

		mockUseCase.On(
			"Execute",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(r *authDto.GoogleLoginRequest) bool {

				return r.IPAddress == "192.168.1.100" &&
					r.UserAgent == "Mozilla Test Agent"
			}),
		).Return(
			response,
			nil,
		)

		reqBody, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/google/login",
			bytes.NewBuffer(reqBody),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		ctx.Request.Header.Set(
			"User-Agent",
			"Mozilla Test Agent",
		)

		ctx.Request.RemoteAddr = "192.168.1.100:12345"

		// When
		handler.GoogleLogin(ctx)

		// Then
		assert.Equal(
			t,
			http.StatusOK,
			w.Code,
		)

		mockUseCase.AssertExpectations(t)
	})
}

func TestGoogleLoginHandler_RegisterRoutes(t *testing.T) {

	t.Run("should register google login route correctly", func(t *testing.T) {

		// Setup
		handler, _ := setupGoogleLoginHandler()

		router := gin.New()

		api := router.Group("/api/v1")

		// When
		handler.RegisterRoutes(api)

		// Then
		routes := router.Routes()

		expectedRoutes := []string{
			"POST /api/v1/google/login",
		}

		for _, expectedRoute := range expectedRoutes {

			found := false

			for _, route := range routes {

				if route.Method+" "+route.Path == expectedRoute {
					found = true
					break
				}
			}

			assert.True(
				t,
				found,
				"Route %s not found",
				expectedRoute,
			)
		}
	})
}
