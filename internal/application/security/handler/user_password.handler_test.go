package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	dto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	testutil "github.com/novriyantoAli/moodly/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserPasswordHandler() (
	*UserPasswordHandler,
	*testutil.MockUserPasswordService,
) {
	gin.SetMode(gin.TestMode)

	mockService := &testutil.MockUserPasswordService{}
	logger := testutil.NewSilentLogger()

	handler := NewUserPasswordHandler(
		mockService,
		logger,
	)

	return handler, mockService
}

func TestUserPasswordHandler_SetPassword(
	t *testing.T,
) {

	t.Run("should set password successfully", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		req := &dto.SetPasswordRequest{
			UserID:   1,
			Username: "john",
			Password: "Password123",
		}

		mockService.On(
			"SetPassword",
			mock.Anything,
			mock.MatchedBy(func(
				r *dto.SetPasswordRequest,
			) bool {
				return r.UserID == req.UserID
			}),
		).Return(nil)

		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/user-password",
			bytes.NewBuffer(body),
		)

		c.Request.Header.Set(
			"Content-Type",
			"application/json",
		)

		handler.SetPassword(c)

		assert.Equal(
			t,
			http.StatusOK,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid json", func(t *testing.T) {

		handler, _ := setupUserPasswordHandler()

		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/user-password",
			bytes.NewBuffer([]byte("invalid")),
		)

		handler.SetPassword(c)

		assert.Equal(
			t,
			http.StatusBadRequest,
			w.Code,
		)
	})

	t.Run("should return internal server error", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		req := &dto.SetPasswordRequest{
			UserID:   1,
			Username: "john",
			Password: "Password123",
		}

		mockService.On(
			"SetPassword",
			mock.Anything,
			mock.Anything,
		).Return(errors.New("database error"))

		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/user-password",
			bytes.NewBuffer(body),
		)

		handler.SetPassword(c)

		assert.Equal(
			t,
			http.StatusInternalServerError,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})
}

func TestUserPasswordHandler_VerifyPassword(
	t *testing.T,
) {

	t.Run("should verify password successfully", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		req := &dto.VerifyPasswordRequest{
			Username: "john",
			Password: "Password123",
		}

		mockService.On(
			"VerifyPassword",
			mock.Anything,
			mock.Anything,
		).Return(true, nil)

		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/user-password/verify",
			bytes.NewBuffer(body),
		)

		handler.VerifyPassword(c)

		assert.Equal(
			t,
			http.StatusOK,
			w.Code,
		)

		mockService.AssertExpectations(t)

		var result map[string]interface{}

		_ = json.Unmarshal(
			w.Body.Bytes(),
			&result,
		)

		assert.Equal(
			t,
			true,
			result["verified"],
		)
	})

	t.Run("should return false when password invalid", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		req := &dto.VerifyPasswordRequest{
			Username: "john",
			Password: "wrong-password",
		}

		mockService.On(
			"VerifyPassword",
			mock.Anything,
			mock.Anything,
		).Return(false, nil)

		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/user-password/verify",
			bytes.NewBuffer(body),
		)

		handler.VerifyPassword(c)

		assert.Equal(
			t,
			http.StatusOK,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		req := &dto.VerifyPasswordRequest{
			Username: "john",
			Password: "Password123",
		}

		mockService.On(
			"VerifyPassword",
			mock.Anything,
			mock.Anything,
		).Return(false, errors.New("db error"))

		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/user-password/verify",
			bytes.NewBuffer(body),
		)

		handler.VerifyPassword(c)

		assert.Equal(
			t,
			http.StatusInternalServerError,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})
}

func TestUserPasswordHandler_ChangePassword(
	t *testing.T,
) {

	t.Run("should change password successfully", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		req := &dto.ChangePasswordRequest{
			UserID:          1,
			CurrentPassword: "Password123",
			NewPassword:     "Password456",
			ConfirmPassword: "Password456",
		}

		mockService.On(
			"ChangePassword",
			mock.Anything,
			mock.Anything,
		).Return(nil)

		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(
			http.MethodPut,
			"/user-password/change",
			bytes.NewBuffer(body),
		)

		handler.ChangePassword(c)

		assert.Equal(
			t,
			http.StatusOK,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		req := &dto.ChangePasswordRequest{
			UserID:          1,
			CurrentPassword: "Password123",
			NewPassword:     "Password456",
			ConfirmPassword: "Password456",
		}

		mockService.On(
			"ChangePassword",
			mock.Anything,
			mock.Anything,
		).Return(errors.New("change failed"))

		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(
			http.MethodPut,
			"/user-password/change",
			bytes.NewBuffer(body),
		)

		handler.ChangePassword(c)

		assert.Equal(
			t,
			http.StatusInternalServerError,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})
}

func TestUserPasswordHandler_GetPasswordInfo(
	t *testing.T,
) {

	t.Run("should get password info successfully", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		resp := &dto.UserPasswordResponse{
			UserID:        1,
			Username:      "john",
			FailedAttempt: 0,
			IsLocked:      false,
		}

		mockService.On(
			"GetPasswordInfo",
			mock.Anything,
			uint(1),
		).Return(resp, nil)

		w := httptest.NewRecorder()

		req := httptest.NewRequest(
			http.MethodGet,
			"/user-password?user_id=1",
			nil,
		)

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetPasswordInfo(c)

		assert.Equal(
			t,
			http.StatusOK,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request when user_id missing", func(t *testing.T) {

		handler, _ := setupUserPasswordHandler()

		w := httptest.NewRecorder()

		req := httptest.NewRequest(
			http.MethodGet,
			"/user-password",
			nil,
		)

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetPasswordInfo(c)

		assert.Equal(
			t,
			http.StatusBadRequest,
			w.Code,
		)
	})

	t.Run("should return no content when password not found", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		mockService.On(
			"GetPasswordInfo",
			mock.Anything,
			uint(1),
		).Return(nil, nil)

		w := httptest.NewRecorder()

		req := httptest.NewRequest(
			http.MethodGet,
			"/user-password?user_id=1",
			nil,
		)

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetPasswordInfo(c)

		assert.Equal(
			t,
			http.StatusNoContent,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error", func(t *testing.T) {

		handler, mockService := setupUserPasswordHandler()

		mockService.On(
			"GetPasswordInfo",
			mock.Anything,
			uint(1),
		).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()

		req := httptest.NewRequest(
			http.MethodGet,
			"/user-password?user_id=1",
			nil,
		)

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetPasswordInfo(c)

		assert.Equal(
			t,
			http.StatusInternalServerError,
			w.Code,
		)

		mockService.AssertExpectations(t)
	})
}
