package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/handler"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/repository"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/service"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSubscriberIntegration(t *testing.T) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)

	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)

	subscriberRepo := repository.NewSubscribeRepository(db, logger)
	subscriberService := service.NewSubscribeService(subscriberRepo, logger)
	subscriberHandler := handler.NewSubscribeHandler(subscriberService, logger)

	router := gin.New()
	api := router.Group("/api/v1")
	subscriberHandler.RegisterRoutes(api)

	cleanup := func() {
		testutil.CleanDB(db)
	}

	return router, cleanup
}

func TestSubscriberIntegration_CreateAndGetSubscriber(t *testing.T) {
	router, cleanup := setupSubscriberIntegration(t)
	defer cleanup()

	createReq := &dto.CreateSubscriberRequest{
		Username:  "testuser",
		CallName:  "Test User",
		Password:  "password123",
		Plan:      "pppoe",
		Price:     50000.0,
		StartDate: time.Now(),
	}

	reqBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	subscriberID := int(data["id"].(float64))
	assert.Equal(t, createReq.Username, data["username"])
	assert.Equal(t, createReq.CallName, data["call_name"])
	assert.Equal(t, createReq.Plan, data["plan"])

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/subscribers/%d", subscriberID), nil)

	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var getResp map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &getResp)
	require.NoError(t, err)

	subscriberData := getResp["data"].(map[string]interface{})
	assert.Equal(t, float64(subscriberID), subscriberData["id"])
	assert.Equal(t, createReq.Username, subscriberData["username"])
	assert.Equal(t, createReq.CallName, subscriberData["call_name"])
}

func TestSubscriberIntegration_CreateDuplicateUsername(t *testing.T) {
	router, cleanup := setupSubscriberIntegration(t)
	defer cleanup()

	createReq := &dto.CreateSubscriberRequest{
		Username:  "duplicate_user",
		CallName:  "Duplicate User",
		Password:  "password123",
		Plan:      "pppoe",
		Price:     50000.0,
		StartDate: time.Now(),
	}

	reqBody, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(reqBody))
	req1.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(reqBody))
	req2.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)

	var errorResp map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "username already exists")
}

func TestSubscriberIntegration_GetSubscribers(t *testing.T) {
	router, cleanup := setupSubscriberIntegration(t)
	defer cleanup()

	subscribers := []dto.CreateSubscriberRequest{
		{
			Username:  "user1",
			CallName:  "User 1",
			Password:  "pass123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
		},
		{
			Username:  "user2",
			CallName:  "User 2",
			Password:  "pass123",
			Plan:      "hotspot",
			Price:     30000.0,
			StartDate: time.Now(),
		},
		{
			Username:  "user3",
			CallName:  "User 3",
			Password:  "pass123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
		},
	}

	for _, subscriber := range subscribers {
		reqBody, _ := json.Marshal(subscriber)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/subscribers?page=1&page_size=10", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SubscriberListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Data, 3)
	assert.Equal(t, int64(3), response.TotalCount)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
}

func TestSubscriberIntegration_UpdateSubscriber(t *testing.T) {
	router, cleanup := setupSubscriberIntegration(t)
	defer cleanup()

	createReq := &dto.CreateSubscriberRequest{
		Username:  "update_user",
		CallName:  "Original Name",
		Password:  "password123",
		Plan:      "pppoe",
		Price:     50000.0,
		StartDate: time.Now(),
	}

	reqBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	subscriberID := int(data["id"].(float64))

	updateReq := &dto.UpdateSubscriberRequest{
		CallName: "Updated Name",
		Plan:     "hotspot",
		Price:    30000.0,
		IsActive: false,
	}

	updateBody, _ := json.Marshal(updateReq)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/subscribers/%d", subscriberID), bytes.NewBuffer(updateBody))
	req2.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var updateResp map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &updateResp)
	require.NoError(t, err)

	updatedData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, updateReq.CallName, updatedData["call_name"])
	assert.Equal(t, updateReq.Plan, updatedData["plan"])
	assert.Equal(t, updateReq.Price, updatedData["price"])
	assert.Equal(t, updateReq.IsActive, updatedData["is_active"])
}

func TestSubscriberIntegration_DeleteSubscriber(t *testing.T) {
	router, cleanup := setupSubscriberIntegration(t)
	defer cleanup()

	createReq := &dto.CreateSubscriberRequest{
		Username:  "delete_user",
		CallName:  "To Be Deleted",
		Password:  "password123",
		Plan:      "pppoe",
		Price:     50000.0,
		StartDate: time.Now(),
	}

	reqBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	subscriberID := int(data["id"].(float64))

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/subscribers/%d", subscriberID), nil)

	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var deleteResp map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &deleteResp)
	require.NoError(t, err)
	assert.Equal(t, "Subscriber deleted successfully", deleteResp["message"])

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/subscribers/%d", subscriberID), nil)

	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusNotFound, w3.Code)
}

func TestSubscriberIntegration_InvalidPlanValidation(t *testing.T) {
	router, cleanup := setupSubscriberIntegration(t)
	defer cleanup()

	createReq := &dto.CreateSubscriberRequest{
		Username:  "invalid_plan_user",
		CallName:  "Invalid Plan User",
		Password:  "password123",
		Plan:      "invalid_plan",
		Price:     50000.0,
		StartDate: time.Now(),
	}

	reqBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.NotEmpty(t, errorResp["error"])
}

func TestSubscriberIntegration_Pagination(t *testing.T) {
	router, cleanup := setupSubscriberIntegration(t)
	defer cleanup()

	for i := 1; i <= 15; i++ {
		createReq := &dto.CreateSubscriberRequest{
			Username:  fmt.Sprintf("user%d", i),
			CallName:  fmt.Sprintf("User %d", i),
			Password:  "password123",
			Plan:      "pppoe",
			Price:     50000.0,
			StartDate: time.Now(),
		}

		reqBody, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/subscribers", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/api/v1/subscribers?page=1&page_size=10", nil)

	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	var response1 dto.SubscriberListResponse
	err := json.Unmarshal(w1.Body.Bytes(), &response1)
	require.NoError(t, err)

	assert.Len(t, response1.Data, 10)
	assert.Equal(t, int64(15), response1.TotalCount)
	assert.Equal(t, 1, response1.Page)
	assert.Equal(t, 10, response1.PageSize)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/api/v1/subscribers?page=2&page_size=10", nil)

	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 dto.SubscriberListResponse
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	require.NoError(t, err)

	assert.Len(t, response2.Data, 5)
	assert.Equal(t, int64(15), response2.TotalCount)
	assert.Equal(t, 2, response2.Page)
	assert.Equal(t, 10, response2.PageSize)
}
