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

func setupLoginHandler() (*LoginHandler, *testutil.MockLoginUseCase) {
	gin.SetMode(gin.TestMode)

	mockUseCase := &testutil.MockLoginUseCase{}
	logger := testutil.NewSilentLogger()

	handler := NewLoginHandler(
		mockUseCase,
		logger,
	)

	return handler, mockUseCase
}

func TestLoginHandler_Login(t *testing.T) {

	t.Run("should login successfully", func(t *testing.T) {

		// Setup
		handler, mockUseCase := setupLoginHandler()

		req := authDto.LoginRequest{
			Username: "admin@example.com",
			Password: "password123",
		}

		response := &authDto.LoginResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiredAt:    9999999999,
		}

		mockUseCase.On(
			"Execute",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(r *authDto.LoginRequest) bool {
				return r.Username == req.Username &&
					r.Password == req.Password
			}),
		).Return(response, nil)

		reqBody, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/login",
			bytes.NewBuffer(reqBody),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.Login(ctx)

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
			"login success",
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
		handler, _ := setupLoginHandler()

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/login",
			bytes.NewBuffer([]byte("invalid json")),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.Login(ctx)

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

	t.Run("should return unauthorized when login failed", func(t *testing.T) {

		// Setup
		handler, mockUseCase := setupLoginHandler()

		req := authDto.LoginRequest{
			Username: "admin@example.com",
			Password: "wrong-password",
		}

		mockUseCase.On(
			"Execute",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(r *authDto.LoginRequest) bool {
				return r.Username == req.Username
			}),
		).Return(
			nil,
			errors.New("invalid username or password"),
		)

		reqBody, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/login",
			bytes.NewBuffer(reqBody),
		)

		ctx.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		// When
		handler.Login(ctx)

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
			"invalid username or password",
			result["message"],
		)
	})

	t.Run("should pass client ip and user agent to usecase", func(t *testing.T) {

		// Setup
		handler, mockUseCase := setupLoginHandler()

		req := authDto.LoginRequest{
			Username: "admin@example.com",
			Password: "password123",
		}

		response := &authDto.LoginResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiredAt:    9999999999,
		}

		mockUseCase.On(
			"Execute",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			mock.MatchedBy(func(r *authDto.LoginRequest) bool {

				return r.IPAddress == "192.168.1.100" &&
					r.UserAgent == "Mozilla Test Agent"
			}),
		).Return(response, nil)

		reqBody, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(
			"POST",
			"/login",
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
		handler.Login(ctx)

		// Then
		assert.Equal(
			t,
			http.StatusOK,
			w.Code,
		)

		mockUseCase.AssertExpectations(t)
	})
}

func TestLoginHandler_RegisterRoutes(t *testing.T) {

	t.Run("should register login route correctly", func(t *testing.T) {

		// Setup
		handler, _ := setupLoginHandler()

		router := gin.New()

		api := router.Group("/api/v1")

		// When
		handler.RegisterRoutes(api)

		// Then
		routes := router.Routes()

		expectedRoutes := []string{
			"POST /api/v1/login",
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
