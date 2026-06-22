package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserPINHandler() (*UserPINHandler, *testutil.MockUserPINService) {
	gin.SetMode(gin.TestMode)
	mockService := &testutil.MockUserPINService{}
	logger := testutil.NewSilentLogger()
	handler := NewUserPINHandler(mockService, logger)
	return handler, mockService
}

func TestUserPINHandler_SetPIN(t *testing.T) {
	t.Run("should set PIN successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserPINHandler()

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "1234",
		}

		mockService.On("SetPIN", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.SetPINRequest) bool {
			return r.UserID == req.UserID && r.PIN == req.PIN
		})).Return(nil)

		// Prepare request
		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.SetPIN(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "message")
		assert.Equal(t, "PIN set successfully", result["message"])
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		// Setup
		handler, _ := setupUserPINHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.SetPIN(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for missing required fields", func(t *testing.T) {
		// Setup
		handler, _ := setupUserPINHandler()

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "", // Missing PIN
		}

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.SetPIN(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserPINHandler()

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "1234",
		}

		mockService.On("SetPIN", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.SetPINRequest) bool {
			return r.UserID == req.UserID && r.PIN == req.PIN
		})).Return(errors.New("database error"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.SetPIN(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "error")
		assert.Equal(t, "Failed to set PIN", result["error"])
	})

	t.Run("should return bad request for invalid PIN length", func(t *testing.T) {
		// Setup
		handler, _ := setupUserPINHandler()

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "123", // Too short (min 4)
		}

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.SetPIN(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for non-numeric PIN", func(t *testing.T) {
		// Setup
		handler, _ := setupUserPINHandler()

		req := &dto.SetPINRequest{
			UserID: 1,
			PIN:    "abcd", // Non-numeric
		}

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.SetPIN(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserPINHandler_VerifyPIN(t *testing.T) {
	t.Run("should verify PIN successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserPINHandler()

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "1234",
		}

		mockService.On("VerifyPIN", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.VerifyPINRequest) bool {
			return r.UserID == req.UserID && r.PIN == req.PIN
		})).Return(true, nil)

		// Prepare request
		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin/verify", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.VerifyPIN(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "verified")
		assert.Equal(t, true, result["verified"])
	})

	t.Run("should return false when PIN verification fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserPINHandler()

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "5678", // Wrong PIN
		}

		mockService.On("VerifyPIN", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.VerifyPINRequest) bool {
			return r.UserID == req.UserID && r.PIN == req.PIN
		})).Return(false, nil)

		// Prepare request
		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin/verify", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.VerifyPIN(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "verified")
		assert.Equal(t, false, result["verified"])
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		// Setup
		handler, _ := setupUserPINHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin/verify", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.VerifyPIN(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for missing required fields", func(t *testing.T) {
		// Setup
		handler, _ := setupUserPINHandler()

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "", // Missing PIN
		}

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin/verify", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.VerifyPIN(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserPINHandler()

		req := &dto.VerifyPINRequest{
			UserID: 1,
			PIN:    "1234",
		}

		mockService.On("VerifyPIN", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.VerifyPINRequest) bool {
			return r.UserID == req.UserID && r.PIN == req.PIN
		})).Return(false, errors.New("database error"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/user-security/pin/verify", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.VerifyPIN(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "error")
		assert.Equal(t, "Failed to verify PIN", result["error"])
	})
}

func TestUserPINHandler_GetSecurity(t *testing.T) {
	t.Run("should return bad request when user_id is not provided", func(t *testing.T) {
		// Setup
		handler, _ := setupUserPINHandler()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/user-security", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// When
		handler.GetSecurity(c)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "error")
		assert.Equal(t, "user_id is required", result["error"])
	})

	t.Run("should return bad request when user_id is invalid", func(t *testing.T) {
		// Setup
		handler, _ := setupUserPINHandler()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/user-security?user_id=invalid", nil)
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// When
		handler.GetSecurity(c)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
