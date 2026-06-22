package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupSubscribeHandler() (*SubscribeHandler, *testutil.MockSubscribeService) {
	gin.SetMode(gin.TestMode)
	mockService := &testutil.MockSubscribeService{}
	logger := testutil.NewSilentLogger()
	handler := NewSubscribeHandler(mockService, logger)
	return handler, mockService
}

func TestSubscribeHandler_CreateSubscriber(t *testing.T) {
	t.Run("should create subscriber successfully", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		req := testutil.CreateSubscriberRequestFixture()
		response := &dto.SubscriberResponse{
			ID:        1,
			Username:  req.Username,
			CallName:  req.CallName,
			Plan:      req.Plan,
			Price:     req.Price,
			StartDate: req.StartDate,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockService.On("CreateSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.CreateSubscriberRequest) bool {
			return r.Username == req.Username && r.CallName == req.CallName
		})).Return(response, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/subscribers", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		handler.CreateSubscriber(ctx)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, req.Username, data["username"])
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		handler, _ := setupSubscribeHandler()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/subscribers", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		handler.CreateSubscriber(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return conflict when username already exists", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		req := testutil.CreateSubscriberRequestFixture()
		mockService.On("CreateSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.CreateSubscriberRequest) bool {
			return r.Username == req.Username && r.CallName == req.CallName
		})).Return(nil, errors.New("username already exists"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/subscribers", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		handler.CreateSubscriber(ctx)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid plan", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		req := testutil.CreateSubscriberRequestFixture()
		mockService.On("CreateSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.CreateSubscriberRequest) bool {
			return r.Username == req.Username
		})).Return(nil, errors.New("invalid plan, must be 'pppoe' or 'hotspot'"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/subscribers", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		handler.CreateSubscriber(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error for other errors", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		req := testutil.CreateSubscriberRequestFixture()
		mockService.On("CreateSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.CreateSubscriberRequest) bool {
			return r.Username == req.Username && r.CallName == req.CallName
		})).Return(nil, errors.New("database error"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/subscribers", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		handler.CreateSubscriber(ctx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestSubscribeHandler_GetSubscriber(t *testing.T) {
	t.Run("should get subscriber successfully", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		subscriberID := uint(1)
		response := &dto.SubscriberResponse{
			ID:        subscriberID,
			Username:  "testuser",
			CallName:  "Test User",
			Plan:      "pppoe",
			Price:     50000.0,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockService.On("GetSubscriberByID", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), subscriberID).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/subscribers/1", nil)
		ctx.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetSubscriber(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		handler, _ := setupSubscribeHandler()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/subscribers/invalid", nil)
		ctx.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetSubscriber(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return not found when subscriber not found", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		subscriberID := uint(999)
		mockService.On("GetSubscriberByID", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), subscriberID).Return(nil, errors.New("subscriber not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/subscribers/999", nil)
		ctx.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.GetSubscriber(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestSubscribeHandler_GetSubscribers(t *testing.T) {
	t.Run("should get subscribers successfully", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		response := &dto.SubscriberListResponse{
			Data: []dto.SubscriberResponse{
				{ID: 1, Username: "user1", CallName: "User 1", Plan: "pppoe", Price: 50000.0},
				{ID: 2, Username: "user2", CallName: "User 2", Plan: "hotspot", Price: 30000.0},
			},
			TotalCount: 2,
			Page:       1,
			PageSize:   10,
		}
		mockService.On("GetSubscribers", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SubscribeFilter) bool {
			return true
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/subscribers?page=1&page_size=10", nil)

		handler.GetSubscribers(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		var result dto.SubscriberListResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("should return bad request for invalid query parameters", func(t *testing.T) {
		handler, _ := setupSubscribeHandler()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/subscribers?page=invalid", nil)

		handler.GetSubscribers(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		mockService.On("GetSubscribers", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SubscribeFilter) bool {
			return true
		})).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/subscribers", nil)

		handler.GetSubscribers(ctx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestSubscribeHandler_UpdateSubscriber(t *testing.T) {
	t.Run("should update subscriber successfully", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		subscriberID := uint(1)
		req := testutil.CreateUpdateSubscriberRequestFixture()
		response := &dto.SubscriberResponse{
			ID:        subscriberID,
			Username:  "testuser",
			CallName:  req.CallName,
			Plan:      req.Plan,
			Price:     req.Price,
			IsActive:  req.IsActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockService.On("UpdateSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), subscriberID, mock.MatchedBy(func(r *dto.UpdateSubscriberRequest) bool {
			return r.CallName == req.CallName && r.Plan == req.Plan
		})).Return(response, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/subscribers/1", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.UpdateSubscriber(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		handler, _ := setupSubscribeHandler()
		req := testutil.CreateUpdateSubscriberRequestFixture()
		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/subscribers/invalid", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.UpdateSubscriber(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return not found when subscriber not found", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		subscriberID := uint(999)
		req := testutil.CreateUpdateSubscriberRequestFixture()
		mockService.On("UpdateSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), subscriberID, mock.MatchedBy(func(r *dto.UpdateSubscriberRequest) bool {
			return true
		})).Return(nil, errors.New("subscriber not found"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/subscribers/999", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.UpdateSubscriber(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request when plan is invalid", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		subscriberID := uint(1)
		req := testutil.CreateUpdateSubscriberRequestFixture()
		mockService.On("UpdateSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), subscriberID, mock.MatchedBy(func(r *dto.UpdateSubscriberRequest) bool {
			return true
		})).Return(nil, errors.New("invalid plan, must be 'pppoe' or 'hotspot'"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/subscribers/1", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.UpdateSubscriber(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestSubscribeHandler_DeleteSubscriber(t *testing.T) {
	t.Run("should delete subscriber successfully", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		subscriberID := uint(1)
		mockService.On("DeleteSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), subscriberID).Return(nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/subscribers/1", nil)
		ctx.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.DeleteSubscriber(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "message")
		assert.Equal(t, "Subscriber deleted successfully", result["message"])
	})

	t.Run("should return not found when subscriber not found", func(t *testing.T) {
		handler, mockService := setupSubscribeHandler()
		subscriberID := uint(999)
		mockService.On("DeleteSubscriber", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), subscriberID).Return(errors.New("subscriber not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/subscribers/999", nil)
		ctx.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.DeleteSubscriber(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		handler, _ := setupSubscribeHandler()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/subscribers/invalid", nil)
		ctx.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.DeleteSubscriber(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSubscribeHandler_RegisterRoutes(t *testing.T) {
	t.Run("should register all routes correctly", func(t *testing.T) {
		handler, _ := setupSubscribeHandler()
		router := gin.New()
		api := router.Group("/api/v1")

		handler.RegisterRoutes(api)

		routes := router.Routes()
		expectedRoutes := []string{
			"POST /api/v1/subscribers",
			"GET /api/v1/subscribers",
			"GET /api/v1/subscribers/:id",
			"PUT /api/v1/subscribers/:id",
			"DELETE /api/v1/subscribers/:id",
		}

		for _, expectedRoute := range expectedRoutes {
			found := false
			for _, route := range routes {
				if route.Method+" "+route.Path == expectedRoute {
					found = true
					break
				}
			}
			assert.True(t, found, "Route %s not found", expectedRoute)
		}
	})
}
