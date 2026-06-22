package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/application/oauth/dto"
	"github.com/novriyantoAli/moodly/internal/application/user/entity"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOAuthService is a mock implementation of service.OAuthService
type MockOAuthService struct {
	mock.Mock
}

func (m *MockOAuthService) GetAuthorizationURL(ctx context.Context, provider dto.OAuthProvider, redirectURI string) (*dto.OAuthAuthorizationURLResponse, error) {
	args := m.Called(ctx, provider, redirectURI)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OAuthAuthorizationURLResponse), args.Error(1)
}

func (m *MockOAuthService) ExchangeCodeForToken(ctx context.Context, provider dto.OAuthProvider, code string) (*dto.OAuthTokenResponse, error) {
	args := m.Called(ctx, provider, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OAuthTokenResponse), args.Error(1)
}

func (m *MockOAuthService) GetUserInfo(ctx context.Context, provider dto.OAuthProvider, accessToken string) (*dto.OAuthUserInfo, error) {
	args := m.Called(ctx, provider, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OAuthUserInfo), args.Error(1)
}

func (m *MockOAuthService) Authenticate(ctx context.Context, provider dto.OAuthProvider, code string, state string) (*dto.OAuthLoginResponse, error) {
	args := m.Called(ctx, provider, code, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OAuthLoginResponse), args.Error(1)
}

func (m *MockOAuthService) RefreshToken(ctx context.Context, provider dto.OAuthProvider, refreshToken string) (*dto.OAuthTokenResponse, error) {
	args := m.Called(ctx, provider, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OAuthTokenResponse), args.Error(1)
}

func (m *MockOAuthService) GetCurrentUser(ctx context.Context, token string) (*entity.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestOAuthHandler_GetAuthorizationURL(t *testing.T) {
	logger := testutil.NewTestLogger(t)
	mockService := new(MockOAuthService)

	t.Run("should return authorization URL successfully", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("GetAuthorizationURL", ctx, dto.GoogleProvider, "http://localhost:3000/callback").
			Return(&dto.OAuthAuthorizationURLResponse{
				AuthorizationURL: "https://accounts.google.com/oauth/authorize?...",
				State:            "state123",
			}, nil).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		req := dto.OAuthAuthorizationURLRequest{
			Provider:    dto.GoogleProvider,
			RedirectURI: "http://localhost:3000/callback",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/authorization-url", bytes.NewBuffer(body))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.GetAuthorizationURL(ginCtx)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.OAuthAuthorizationURLResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotEmpty(t, resp.AuthorizationURL)
		assert.NotEmpty(t, resp.State)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid request", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)

		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/authorization-url", bytes.NewBuffer([]byte("invalid")))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.GetAuthorizationURL(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error from service", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("GetAuthorizationURL", ctx, dto.GoogleProvider, "http://localhost:3000/callback").
			Return(nil, errors.New("service error")).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		req := dto.OAuthAuthorizationURLRequest{
			Provider:    dto.GoogleProvider,
			RedirectURI: "http://localhost:3000/callback",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/authorization-url", bytes.NewBuffer(body))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.GetAuthorizationURL(ginCtx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOAuthHandler_ExchangeCodeForToken(t *testing.T) {
	logger := testutil.NewTestLogger(t)
	mockService := new(MockOAuthService)

	t.Run("should exchange code for token successfully", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("ExchangeCodeForToken", ctx, dto.GoogleProvider, "code123").
			Return(&dto.OAuthTokenResponse{
				AccessToken:  "access_token_123",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "refresh_token_123",
			}, nil).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		req := dto.OAuthTokenRequest{
			Provider: dto.GoogleProvider,
			Code:     "code123",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/token", bytes.NewBuffer(body))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.ExchangeCodeForToken(ginCtx)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.OAuthTokenResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "access_token_123", resp.AccessToken)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for empty code", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)
		req := dto.OAuthTokenRequest{
			Provider: dto.GoogleProvider,
			Code:     "",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/token", bytes.NewBuffer(body))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.ExchangeCodeForToken(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error for invalid request", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)

		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/token", bytes.NewBuffer([]byte("invalid")))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.ExchangeCodeForToken(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOAuthHandler_GetUserInfo(t *testing.T) {
	logger := testutil.NewTestLogger(t)
	mockService := new(MockOAuthService)

	t.Run("should get user info successfully", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("GetUserInfo", ctx, dto.GoogleProvider, "access_token_123").
			Return(&dto.OAuthUserInfo{
				ID:        "user_123",
				Email:     "user@example.com",
				Name:      "User Name",
				AvatarURL: "https://example.com/avatar.jpg",
				Provider:  "google",
			}, nil).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/user-info?provider=google&access_token=access_token_123", nil)

		handler.GetUserInfo(ginCtx)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.OAuthUserInfo
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "user@example.com", resp.Email)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for missing provider", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/user-info?access_token=token123", nil)

		handler.GetUserInfo(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error for missing access token", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/user-info?provider=google", nil)

		handler.GetUserInfo(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error from service", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("GetUserInfo", ctx, dto.GoogleProvider, "invalid_token").
			Return(nil, errors.New("invalid token")).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/user-info?provider=google&access_token=invalid_token", nil)

		handler.GetUserInfo(ginCtx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOAuthHandler_Authenticate(t *testing.T) {
	logger := testutil.NewTestLogger(t)
	mockService := new(MockOAuthService)

	t.Run("should authenticate successfully", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("Authenticate", ctx, dto.GoogleProvider, "code123", "state123").
			Return(&dto.OAuthLoginResponse{
				Token: "jwt_token_123",
				UserInfo: dto.OAuthUserInfo{
					ID:        "user_123",
					Email:     "user@example.com",
					Name:      "User Name",
					AvatarURL: "https://example.com/avatar.jpg",
					Provider:  "google",
				},
			}, nil).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		req := dto.OAuthAuthorizationRequest{
			Provider: dto.GoogleProvider,
			Code:     "code123",
			State:    "state123",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/authenticate", bytes.NewBuffer(body))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.Authenticate(ginCtx)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotNil(t, resp["data"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid request", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)

		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/authenticate", bytes.NewBuffer([]byte("invalid")))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.Authenticate(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error from service", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("Authenticate", ctx, dto.GoogleProvider, "invalid_code", "state123").
			Return(nil, errors.New("authentication failed")).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		req := dto.OAuthAuthorizationRequest{
			Provider: dto.GoogleProvider,
			Code:     "invalid_code",
			State:    "state123",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/authenticate", bytes.NewBuffer(body))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.Authenticate(ginCtx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOAuthHandler_RefreshToken(t *testing.T) {
	logger := testutil.NewTestLogger(t)
	mockService := new(MockOAuthService)

	t.Run("should refresh token successfully", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("RefreshToken", ctx, dto.GoogleProvider, "refresh_token_123").
			Return(&dto.OAuthTokenResponse{
				AccessToken:  "new_access_token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "new_refresh_token",
			}, nil).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		req := dto.OAuthAuthorizationRequest{
			Provider: dto.GoogleProvider,
			Code:     "refresh_token_123",
			State:    "state123",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/refresh", bytes.NewBuffer(body))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.RefreshToken(ginCtx)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.OAuthTokenResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "new_access_token", resp.AccessToken)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid request", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)

		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/refresh", bytes.NewBuffer([]byte("invalid")))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.RefreshToken(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error from service", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("RefreshToken", ctx, dto.GoogleProvider, "invalid_token").
			Return(nil, errors.New("invalid refresh token")).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		req := dto.OAuthAuthorizationRequest{
			Provider: dto.GoogleProvider,
			Code:     "invalid_token",
			State:    "state123",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("POST", "/oauth/refresh", bytes.NewBuffer(body))
		ginCtx.Request.Header.Set("Content-Type", "application/json")

		handler.RefreshToken(ginCtx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOAuthHandler_GetCurrentUser(t *testing.T) {
	logger := testutil.NewTestLogger(t)
	mockService := new(MockOAuthService)

	t.Run("should get current user successfully", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("GetCurrentUser", ctx, "jwt_token_123").
			Return(&entity.User{
				ID:       1,
				Email:    "user@example.com",
				FullName: "User Name",
			}, nil).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/me", nil)
		ginCtx.Request.Header.Set("Authorization", "Bearer jwt_token_123")

		handler.GetCurrentUser(ginCtx)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotNil(t, resp["data"])
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for missing authorization header", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/me", nil)

		handler.GetCurrentUser(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error for invalid authorization header format", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/me", nil)
		ginCtx.Request.Header.Set("Authorization", "InvalidFormat token123")

		handler.GetCurrentUser(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error for empty token", func(t *testing.T) {
		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/me", nil)
		ginCtx.Request.Header.Set("Authorization", "Bearer ")

		handler.GetCurrentUser(ginCtx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 401 for invalid token", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("GetCurrentUser", ctx, "invalid_token").
			Return(nil, errors.New("invalid or expired token")).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/me", nil)
		ginCtx.Request.Header.Set("Authorization", "Bearer invalid_token")

		handler.GetCurrentUser(ginCtx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return 404 for user not found", func(t *testing.T) {
		ctx := context.Background()
		mockService.On("GetCurrentUser", ctx, "valid_token").
			Return(nil, errors.New("user not found")).
			Once()

		handler := NewOAuthHandler(mockService, logger)
		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		ginCtx.Request = httptest.NewRequest("GET", "/oauth/me", nil)
		ginCtx.Request.Header.Set("Authorization", "Bearer valid_token")

		handler.GetCurrentUser(ginCtx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}
