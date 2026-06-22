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

	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/bill/dto"
	"github.com/novriyantoAli/moodly/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupBillHandler() (*BillHandler, *testutil.MockBillService) {
	gin.SetMode(gin.TestMode)
	mockService := &testutil.MockBillService{}
	logger := testutil.NewSilentLogger()
	handler := NewBillHandler(mockService, logger)
	return handler, mockService
}

func TestBillHandler_CreateBill(t *testing.T) {
	t.Run("should create bill successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		req := testutil.CreateBillRequestFixture()
		response := &dto.BillResponse{
			ID:          uuid.New(),
			SubscribeID: req.SubscribeID,
			BillMonth:   req.BillMonth,
			BillYear:    req.BillYear,
			Amount:      req.Amount,
			DueDate:     req.DueDate,
			Status:      "unpaid",
			CreatedAt:   time.Now(),
		}

		mockService.On("CreateBill", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.CreateBillRequest) bool {
			return r.SubscribeID == req.SubscribeID && r.Amount == req.Amount
		})).Return(response, nil)

		// Prepare request
		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/bills", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateBill(ctx)

		// Then
		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(req.SubscribeID), data["subscribe_id"])
		assert.Equal(t, float64(req.Amount), data["amount"])
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		// Setup
		handler, _ := setupBillHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/bills", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateBill(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		req := testutil.CreateBillRequestFixture()
		mockService.On("CreateBill", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(r *dto.CreateBillRequest) bool {
			return r.SubscribeID == req.SubscribeID
		})).Return(nil, errors.New("database error"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/bills", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateBill(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestBillHandler_GetBill(t *testing.T) {
	t.Run("should get bill successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		billID := uuid.New()
		response := &dto.BillResponse{
			ID:          billID,
			SubscribeID: 1,
			BillMonth:   3,
			BillYear:    2026,
			Amount:      150000,
			DueDate:     time.Now().Add(30 * 24 * time.Hour),
			Status:      "unpaid",
			CreatedAt:   time.Now(),
		}

		mockService.On("GetBillByID", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), billID.String()).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/"+billID.String(), nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: billID.String()},
		}

		// When
		handler.GetBill(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["subscribe_id"])
	})

	t.Run("should return bad request for invalid bill ID format", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		invalidID := "invalid-uuid"
		mockService.On("GetBillByID", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), invalidID).Return(nil, errors.New("invalid bill id format"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/"+invalidID, nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: invalidID},
		}

		// When
		handler.GetBill(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when bill not found", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		billID := uuid.New()
		mockService.On("GetBillByID", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), billID.String()).Return(nil, errors.New("bill not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/"+billID.String(), nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: billID.String()},
		}

		// When
		handler.GetBill(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestBillHandler_GetBills(t *testing.T) {
	t.Run("should get bills successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillListResponse{
			Data: []dto.BillResponse{
				{ID: uuid.New(), SubscribeID: 1, BillMonth: 3, Amount: 150000, Status: "unpaid"},
				{ID: uuid.New(), SubscribeID: 2, BillMonth: 3, Amount: 200000, Status: "paid"},
			},
			TotalCount: 2,
			Page:       1,
			PageSize:   10,
		}

		mockService.On("GetBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.BillFilter) bool {
			return true
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills?page=1&page_size=10", nil)

		// When
		handler.GetBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillListResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("should filter bills by subscribe_id", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillListResponse{
			Data: []dto.BillResponse{
				{ID: uuid.New(), SubscribeID: 1, BillMonth: 3, Amount: 150000, Status: "unpaid"},
			},
			TotalCount: 1,
			Page:       1,
			PageSize:   10,
		}

		mockService.On("GetBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.BillFilter) bool {
			return f.SubscribeID == 1
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills?subscribe_id=1", nil)

		// When
		handler.GetBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillListResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Len(t, result.Data, 1)
	})

	t.Run("should filter bills by status", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillListResponse{
			Data: []dto.BillResponse{
				{ID: uuid.New(), SubscribeID: 1, BillMonth: 3, Amount: 150000, Status: "paid"},
			},
			TotalCount: 1,
			Page:       1,
			PageSize:   10,
		}

		mockService.On("GetBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.BillFilter) bool {
			return f.Status == "paid"
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills?status=paid", nil)

		// When
		handler.GetBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillListResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, "paid", result.Data[0].Status)
	})

	t.Run("should return bad request for invalid query parameters", func(t *testing.T) {
		// Setup
		handler, _ := setupBillHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills?page=invalid", nil)

		// When
		handler.GetBills(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("GetBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.BillFilter) bool {
			return true
		})).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills", nil)

		// When
		handler.GetBills(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestBillHandler_UpdateBillStatus(t *testing.T) {
	t.Run("should update bill status successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		billID := uuid.New()
		mockService.On("UpdateBillStatus", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), billID.String(), "paid").Return(nil)

		reqBody, _ := json.Marshal(map[string]string{"status": "paid"})
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/bills/"+billID.String()+"/status", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: billID.String()},
		}

		// When
		handler.UpdateBillStatus(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "message")
		assert.Equal(t, "Bill status updated successfully", result["message"])
	})

	t.Run("should return bad request for missing status field", func(t *testing.T) {
		// Setup
		handler, _ := setupBillHandler()

		billID := uuid.New()
		reqBody, _ := json.Marshal(map[string]string{})
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/bills/"+billID.String()+"/status", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: billID.String()},
		}

		// When
		handler.UpdateBillStatus(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for invalid status value", func(t *testing.T) {
		// Setup
		handler, _ := setupBillHandler()

		billID := uuid.New()
		reqBody, _ := json.Marshal(map[string]string{"status": "invalid-status"})
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/bills/"+billID.String()+"/status", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: billID.String()},
		}

		// When
		handler.UpdateBillStatus(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return bad request for invalid bill ID format", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		invalidID := "invalid-uuid"
		mockService.On("UpdateBillStatus", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), invalidID, "paid").Return(errors.New("invalid bill id format"))

		reqBody, _ := json.Marshal(map[string]string{"status": "paid"})
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/bills/"+invalidID+"/status", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: invalidID},
		}

		// When
		handler.UpdateBillStatus(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when bill not found", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		billID := uuid.New()
		mockService.On("UpdateBillStatus", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), billID.String(), "paid").Return(errors.New("bill not found"))

		reqBody, _ := json.Marshal(map[string]string{"status": "paid"})
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/bills/"+billID.String()+"/status", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: billID.String()},
		}

		// When
		handler.UpdateBillStatus(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should accept all valid status values", func(t *testing.T) {
		handler, mockService := setupBillHandler()

		billID := uuid.New()
		validStatuses := []string{"paid", "unpaid", "pending", "overdue"}

		for _, status := range validStatuses {
			mockService.On("UpdateBillStatus", mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}), billID.String(), status).Return(nil)

			reqBody, _ := json.Marshal(map[string]string{"status": status})
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("PUT", "/bills/"+billID.String()+"/status", bytes.NewBuffer(reqBody))
			ctx.Request.Header.Set("Content-Type", "application/json")
			ctx.Params = gin.Params{
				{Key: "id", Value: billID.String()},
			}

			// When
			handler.UpdateBillStatus(ctx)

			// Then
			assert.Equal(t, http.StatusOK, w.Code, "should accept status: %s", status)
		}
	})
}

func TestBillHandler_RegisterRoutes(t *testing.T) {
	t.Run("should register all routes correctly", func(t *testing.T) {
		// Setup
		handler, _ := setupBillHandler()
		router := gin.New()
		api := router.Group("/api/v1")

		// When
		handler.RegisterRoutes(api)

		// Then
		routes := router.Routes()
		expectedRoutes := []string{
			"POST /api/v1/bills",
			"GET /api/v1/bills",
			"GET /api/v1/bills/:id",
			"PUT /api/v1/bills/:id/status",
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

func TestBillHandler_QuickCountUnpaidBills(t *testing.T) {
	t.Run("should get unpaid bills count and amount successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillQuickCountUnpaidResponse{
			Count:  5,
			Amount: 750000,
		}

		mockService.On("QuickCountUnpaidBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/unpaid", nil)

		// When
		handler.QuickCountUnpaidBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillQuickCountUnpaidResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, int64(5), result.Count)
		assert.Equal(t, int64(750000), result.Amount)
	})

	t.Run("should return zero values when no unpaid bills", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillQuickCountUnpaidResponse{
			Count:  0,
			Amount: 0,
		}

		mockService.On("QuickCountUnpaidBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/unpaid", nil)

		// When
		handler.QuickCountUnpaidBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillQuickCountUnpaidResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, int64(0), result.Count)
		assert.Equal(t, int64(0), result.Amount)
	})

	t.Run("should handle high count and amount values", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillQuickCountUnpaidResponse{
			Count:  1000,
			Amount: 150000000,
		}

		mockService.On("QuickCountUnpaidBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/unpaid", nil)

		// When
		handler.QuickCountUnpaidBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillQuickCountUnpaidResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, int64(1000), result.Count)
		assert.Equal(t, int64(150000000), result.Amount)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("QuickCountUnpaidBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/unpaid", nil)

		// When
		handler.QuickCountUnpaidBills(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "error")
		assert.Equal(t, "Failed to count unpaid bills", result["error"])
	})
}

func TestBillHandler_QuickCountOverdueBills(t *testing.T) {
	t.Run("should get overdue bills count and amount successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillQuickCountOverdueResponse{
			Count:  3,
			Amount: 450000,
		}

		mockService.On("QuickCountOverdueBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/overdue", nil)

		// When
		handler.QuickCountOverdueBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillQuickCountOverdueResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, int64(3), result.Count)
		assert.Equal(t, int64(450000), result.Amount)
	})

	t.Run("should return zero values when no overdue bills", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillQuickCountOverdueResponse{
			Count:  0,
			Amount: 0,
		}

		mockService.On("QuickCountOverdueBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/overdue", nil)

		// When
		handler.QuickCountOverdueBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillQuickCountOverdueResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, int64(0), result.Count)
		assert.Equal(t, int64(0), result.Amount)
	})

	t.Run("should handle high count and amount values", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		response := &dto.BillQuickCountOverdueResponse{
			Count:  500,
			Amount: 75000000,
		}

		mockService.On("QuickCountOverdueBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/overdue", nil)

		// When
		handler.QuickCountOverdueBills(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.BillQuickCountOverdueResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, int64(500), result.Count)
		assert.Equal(t, int64(75000000), result.Amount)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("QuickCountOverdueBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/overdue", nil)

		// When
		handler.QuickCountOverdueBills(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "error")
		assert.Equal(t, "Failed to count overdue bills", result["error"])
	})

	t.Run("should handle service timeout", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("QuickCountOverdueBills", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		})).Return(nil, errors.New("context deadline exceeded"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/quick-count/overdue", nil)

		// When
		handler.QuickCountOverdueBills(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestBillHandler_GetSumAmountMonthYear(t *testing.T) {
	t.Run("should get sum amount successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("SumAmountByMonthYear", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SumAmountBillFilter) bool {
			return f.Month == 3 && f.Year == 2026 && f.Status == "paid"
		})).Return(int64(500000), nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/sum/amount/month/year?month=3&year=2026&status=paid", nil)

		// When
		handler.GetSumAmountMonthYear(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		assert.Equal(t, float64(500000), result["data"])
	})

	t.Run("should return zero amount when no bills found", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("SumAmountByMonthYear", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SumAmountBillFilter) bool {
			return f.Month == 3 && f.Year == 2026 && f.Status == "paid"
		})).Return(int64(0), nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/sum/amount/month/year?month=3&year=2026&status=paid", nil)

		// When
		handler.GetSumAmountMonthYear(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, float64(0), result["data"])
	})

	t.Run("should handle high amount values", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("SumAmountByMonthYear", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SumAmountBillFilter) bool {
			return f.Month == 3 && f.Year == 2026 && f.Status == "all"
		})).Return(int64(150000000), nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/sum/amount/month/year?month=3&year=2026&status=all", nil)

		// When
		handler.GetSumAmountMonthYear(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, float64(150000000), result["data"])
	})

	t.Run("should return bad request for invalid query parameters", func(t *testing.T) {
		// Setup
		handler, _ := setupBillHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/sum/amount/month/year?month=invalid&year=2026", nil)

		// When
		handler.GetSumAmountMonthYear(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should use default values for missing parameters", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("SumAmountByMonthYear", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SumAmountBillFilter) bool {
			return true // Service should set defaults
		})).Return(int64(250000), nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/sum/amount/month/year", nil)

		// When
		handler.GetSumAmountMonthYear(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, float64(250000), result["data"])
	})

	t.Run("should filter by status unpaid", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("SumAmountByMonthYear", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SumAmountBillFilter) bool {
			return f.Status == "unpaid"
		})).Return(int64(300000), nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/sum/amount/month/year?month=3&year=2026&status=unpaid", nil)

		// When
		handler.GetSumAmountMonthYear(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should filter by status overdue", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("SumAmountByMonthYear", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SumAmountBillFilter) bool {
			return f.Status == "overdue"
		})).Return(int64(100000), nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/sum/amount/month/year?month=3&year=2026&status=overdue", nil)

		// When
		handler.GetSumAmountMonthYear(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupBillHandler()

		mockService.On("SumAmountByMonthYear", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), mock.MatchedBy(func(f *dto.SumAmountBillFilter) bool {
			return true
		})).Return(int64(0), errors.New("database error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/bills/sum/amount/month/year?month=3&year=2026&status=paid", nil)

		// When
		handler.GetSumAmountMonthYear(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "error")
		assert.Equal(t, "Failed to sum amount by month and year", result["error"])
	})
}
