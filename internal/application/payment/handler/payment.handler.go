package handler

import (
	"net/http"
	"strconv"

	"github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/payment/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	service service.PaymentService
	logger  *zap.Logger
}

func NewPaymentHandler(service service.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		service: service,
		logger:  logger,
	}
}

// CreatePayment godoc
// @Summary Create a new payment
// @Description Create a new payment with the provided information
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body dto.CreatePaymentRequest true "Payment creation request"
// @Success 201 {object} map[string]interface{} "Created payment"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(ctx *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.service.CreatePayment(ctx.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create payment", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": payment})
}

// GetPayment godoc
// @Summary Get a payment by ID
// @Description Get a single payment by its ID
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} map[string]interface{} "Payment details"
// @Failure 400 {object} map[string]interface{} "Invalid payment ID"
// @Failure 404 {object} map[string]interface{} "Payment not found"
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPayment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	payment, err := h.service.GetPaymentByID(ctx.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get payment", zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": payment})
}

// GetPayments godoc
// @Summary Get all payments
// @Description Get a list of payments with optional filtering and pagination
// @Tags payments
// @Accept json
// @Produce json
// @Param status query string false "Filter by status" Enums(pending, completed, failed, canceled)
// @Param currency query string false "Filter by currency (3-letter code)"
// @Param user_id query int false "Filter by user ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} dto.PaymentListResponse "List of payments"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /payments [get]
func (h *PaymentHandler) GetPayments(ctx *gin.Context) {
	var filter dto.PaymentFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payments, err := h.service.GetPayments(ctx.Request.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to get payments", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payments"})
		return
	}

	ctx.JSON(http.StatusOK, payments)
}

// UpdatePayment godoc
// @Summary Update a payment
// @Description Update a payment's information by ID
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param payment body dto.UpdatePaymentRequest true "Payment update request"
// @Success 200 {object} map[string]interface{} "Updated payment"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /payments/{id} [put]
func (h *PaymentHandler) UpdatePayment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	var req dto.UpdatePaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.service.UpdatePayment(ctx.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to update payment", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": payment})
}

func (h *PaymentHandler) RegisterRoutes(api *gin.RouterGroup) {
	payments := api.Group("/payments")
	{
		payments.POST("", h.CreatePayment)
		payments.GET("", h.GetPayments)
		payments.GET("/:id", h.GetPayment)
		payments.PUT("/:id", h.UpdatePayment)
	}
}
