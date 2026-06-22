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

func setupLogoutHandler() (
	*LogoutHandler,
	*testutil.MockLogoutUseCase,
) {

	gin.SetMode(gin.TestMode)

	mockUseCase := &testutil.MockLogoutUseCase{}

	logger := testutil.NewSilentLogger()

	handler := NewLogoutHandler(
		mockUseCase,
		logger,
	)

	return handler, mockUseCase
}

func TestLogoutHandler_Logout(t *testing.T) {

	t.Run("should logout successfully", func(t *testing.T) {

		// Setup
		handler, mockUseCase := setupLogoutHandler()

		req := authDto.LogoutRequest{
			RefreshToken: "refresh-token",
		}

		mockUseCase.On(
			"Execute",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(r *authDto.LogoutRequest) bool {
				return r.RefreshToken == "refresh-token"
			}),
		).Return(nil)

		reqBody, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/logout",
			bytes.NewBuffer(reqBody),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.Logout(ctx)

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
			"logout success",
			result["message"],
		)
	})

	t.Run("should return bad request for invalid json", func(t *testing.T) {

		// Setup
		handler, _ := setupLogoutHandler()

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/logout",
			bytes.NewBuffer([]byte("invalid json")),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.Logout(ctx)

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

	t.Run("should return unauthorized when logout failed", func(t *testing.T) {

		// Setup
		handler, mockUseCase := setupLogoutHandler()

		req := authDto.LogoutRequest{
			RefreshToken: "invalid-refresh-token",
		}

		mockUseCase.On(
			"Execute",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(r *authDto.LogoutRequest) bool {
				return r.RefreshToken == "invalid-refresh-token"
			}),
		).Return(
			errors.New("session not found"),
		)

		reqBody, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/logout",
			bytes.NewBuffer(reqBody),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.Logout(ctx)

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
			"session not found",
			result["message"],
		)
	})
}

func TestLogoutHandler_RegisterRoutes(t *testing.T) {

	t.Run("should register logout route correctly", func(t *testing.T) {

		// Setup
		handler, _ := setupLogoutHandler()

		router := gin.New()

		api := router.Group("/api/v1")

		// When
		handler.RegisterRoutes(api)

		// Then
		routes := router.Routes()

		expectedRoutes := []string{
			"POST /api/v1/logout",
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
