package handler

import (
	"net/http"

	"github.com/novriyantoAli/moodly/internal/application/bill/dto"
	"github.com/novriyantoAli/moodly/internal/application/bill/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BillHandler struct {
	service service.BillService
	logger  *zap.Logger
}

func NewBillHandler(service service.BillService, logger *zap.Logger) *BillHandler {
	return &BillHandler{
		service: service,
		logger:  logger,
	}
}

// CreateBill godoc
// @Summary Create a new bill
// @Description Create a new bill with the provided information
// @Tags bills
// @Accept json
// @Produce json
// @Param bill body dto.CreateBillRequest true "Bill creation request"
// @Success 201 {object} map[string]interface{} "Created bill"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /bills [post]
func (h *BillHandler) CreateBill(ctx *gin.Context) {
	var req dto.CreateBillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bill, err := h.service.CreateBill(ctx.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create bill", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bill"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": bill})
}

// GetBill godoc
// @Summary Get a bill by ID
// @Description Get a single bill by its ID
// @Tags bills
// @Accept json
// @Produce json
// @Param id path string true "Bill ID (UUID)"
// @Success 200 {object} map[string]interface{} "Bill details"
// @Failure 400 {object} map[string]interface{} "Invalid bill ID"
// @Failure 404 {object} map[string]interface{} "Bill not found"
// @Router /bills/{id} [get]
func (h *BillHandler) GetBill(ctx *gin.Context) {
	billID := ctx.Param("id")

	bill, err := h.service.GetBillByID(ctx.Request.Context(), billID)
	if err != nil {
		h.logger.Error("Failed to get bill", zap.String("id", billID), zap.Error(err))
		if err.Error() == "invalid bill id format" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Bill not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": bill})
}

// GetBills godoc
// @Summary Get all bills
// @Description Get a list of bills with optional filtering and pagination
// @Tags bills
// @Accept json
// @Produce json
// @Param subscribe_id query int false "Filter by subscribe ID"
// @Param status query string false "Filter by status (paid, unpaid, pending, overdue)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} dto.BillListResponse "List of bills"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /bills [get]
func (h *BillHandler) GetBills(ctx *gin.Context) {
	var filter dto.BillFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bills, err := h.service.GetBills(ctx.Request.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to get bills", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bills"})
		return
	}

	ctx.JSON(http.StatusOK, bills)
}

// UpdateBillStatus godoc
// @Summary Update bill status
// @Description Update the status of a bill by ID
// @Tags bills
// @Accept json
// @Produce json
// @Param id path string true "Bill ID (UUID)"
// @Param request body map[string]string true "Status update request (status: paid, unpaid, pending, overdue)"
// @Success 200 {object} map[string]interface{} "Status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Bill not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /bills/{id}/status [put]
func (h *BillHandler) UpdateBillStatus(ctx *gin.Context) {
	billID := ctx.Param("id")

	var req struct {
		Status string `json:"status" binding:"required,oneof=paid unpaid pending overdue"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateBillStatus(ctx.Request.Context(), billID, req.Status)
	if err != nil {
		h.logger.Error("Failed to update bill status", zap.String("id", billID), zap.Error(err))
		if err.Error() == "invalid bill id format" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "bill not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invalid status. allowed values: paid, unpaid, pending, overdue" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bill status"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Bill status updated successfully"})
}

func (h *BillHandler) GetSumAmountMonthYear(ctx *gin.Context) {
	var filter dto.SumAmountBillFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount, err := h.service.SumAmountByMonthYear(ctx.Request.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to sum amount by month and year", zap.Uint("month", filter.Month), zap.Uint("year", filter.Year), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sum amount by month and year"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": amount})
}

func (h *BillHandler) GetCountBillsByMonthYear(ctx *gin.Context) {
	var filter dto.CountBillFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.CountBillsByMonthYear(ctx.Request.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to count bills by month and year", zap.Uint("month", filter.Month), zap.Uint("year", filter.Year), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count bills by month and year"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *BillHandler) QuickCountUnpaidBills(ctx *gin.Context) {
	result, err := h.service.QuickCountUnpaidBills(ctx.Request.Context())
	if err != nil {
		h.logger.Error("Failed to count unpaid bills", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count unpaid bills"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *BillHandler) QuickCountOverdueBills(ctx *gin.Context) {
	result, err := h.service.QuickCountOverdueBills(ctx.Request.Context())
	if err != nil {
		h.logger.Error("Failed to count overdue bills", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count overdue bills"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *BillHandler) RegisterRoutes(api *gin.RouterGroup) {
	bills := api.Group("/bills")
	{
		bills.POST("", h.CreateBill)
		bills.GET("", h.GetBills)
		bills.GET("/:id", h.GetBill)
		bills.PUT("/:id/status", h.UpdateBillStatus)
		bills.GET("/sum/amount", h.GetSumAmountMonthYear)
		bills.GET("/count", h.GetCountBillsByMonthYear)
		bills.GET("/quick-count/unpaid", h.QuickCountUnpaidBills)
		bills.GET("/quick-count/overdue", h.QuickCountOverdueBills)
	}
}
