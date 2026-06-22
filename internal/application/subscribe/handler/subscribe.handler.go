package handler

import (
	"net/http"
	"strconv"

	"github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	"github.com/novriyantoAli/moodly/internal/application/subscribe/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SubscribeHandler struct {
	service service.SubscribeService
	logger  *zap.Logger
}

func NewSubscribeHandler(service service.SubscribeService, logger *zap.Logger) *SubscribeHandler {
	return &SubscribeHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscriber godoc
// @Summary Create a new subscriber
// @Description Create a new subscriber with the provided information
// @Tags subscribers
// @Accept json
// @Produce json
// @Param subscriber body dto.CreateSubscriberRequest true "Subscriber creation request"
// @Success 201 {object} map[string]interface{} "Created subscriber"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Username already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscribers [post]
func (h *SubscribeHandler) CreateSubscriber(ctx *gin.Context) {
	var req dto.CreateSubscriberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscriber, err := h.service.CreateSubscriber(ctx.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create subscriber", zap.Error(err))
		if err.Error() == "username already exists" {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invalid plan, must be 'pppoe' or 'hotspot'" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscriber"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": subscriber})
}

// GetSubscriber godoc
// @Summary Get a subscriber by ID
// @Description Get a single subscriber by their ID
// @Tags subscribers
// @Accept json
// @Produce json
// @Param id path int true "Subscriber ID"
// @Success 200 {object} map[string]interface{} "Subscriber details"
// @Failure 400 {object} map[string]interface{} "Invalid subscriber ID"
// @Failure 404 {object} map[string]interface{} "Subscriber not found"
// @Router /subscribers/{id} [get]
func (h *SubscribeHandler) GetSubscriber(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	subscriber, err := h.service.GetSubscriberByID(ctx.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get subscriber", zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": subscriber})
}

// GetSubscribers godoc
// @Summary Get all subscribers
// @Description Get a list of subscribers with optional filtering and pagination
// @Tags subscribers
// @Accept json
// @Produce json
// @Param username query string false "Filter by username"
// @Param call_name query string false "Filter by call name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} dto.SubscriberListResponse "List of subscribers"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscribers [get]
func (h *SubscribeHandler) GetSubscribers(ctx *gin.Context) {
	var filter dto.SubscribeFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscribers, err := h.service.GetSubscribers(ctx.Request.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to get subscribers", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscribers"})
		return
	}

	ctx.JSON(http.StatusOK, subscribers)
}

// UpdateSubscriber godoc
// @Summary Update a subscriber
// @Description Update a subscriber's information by ID
// @Tags subscribers
// @Accept json
// @Produce json
// @Param id path int true "Subscriber ID"
// @Param subscriber body dto.UpdateSubscriberRequest true "Subscriber update request"
// @Success 200 {object} map[string]interface{} "Updated subscriber"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Subscriber not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscribers/{id} [put]
func (h *SubscribeHandler) UpdateSubscriber(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	var req dto.UpdateSubscriberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscriber, err := h.service.UpdateSubscriber(ctx.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to update subscriber", zap.Error(err))
		if err.Error() == "subscriber not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invalid plan, must be 'pppoe' or 'hotspot'" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscriber"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": subscriber})
}

// DeleteSubscriber godoc
// @Summary Delete a subscriber
// @Description Delete a subscriber by ID
// @Tags subscribers
// @Accept json
// @Produce json
// @Param id path int true "Subscriber ID"
// @Success 200 {object} map[string]interface{} "Subscriber deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid subscriber ID"
// @Failure 404 {object} map[string]interface{} "Subscriber not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /subscribers/{id} [delete]
func (h *SubscribeHandler) DeleteSubscriber(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	err = h.service.DeleteSubscriber(ctx.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete subscriber", zap.Error(err))
		if err.Error() == "subscriber not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscriber"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Subscriber deleted successfully"})
}

func (h *SubscribeHandler) GetCount(ctx *gin.Context) {
	var filter dto.CountFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count, err := h.service.CountFilter(ctx.Request.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to count subscribers", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count subscribers"})
		return
	}

	ctx.JSON(http.StatusOK, count)
}

func (h *SubscribeHandler) RegisterRoutes(api *gin.RouterGroup) {
	subscribers := api.Group("/subscribers")
	{
		subscribers.POST("", h.CreateSubscriber)
		subscribers.GET("", h.GetSubscribers)
		subscribers.GET("/:id", h.GetSubscriber)
		subscribers.GET("/count", h.GetCount)
		subscribers.PUT("/:id", h.UpdateSubscriber)
		subscribers.DELETE("/:id", h.DeleteSubscriber)
	}
}
