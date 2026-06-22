package handler

import (
	"net/http"
	"strconv"

	"github.com/novriyantoAli/moodly/internal/application/scan/dto"
	"github.com/novriyantoAli/moodly/internal/application/scan/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ScanHandler handles HTTP requests for scan operations
type ScanHandler struct {
	service service.ScanService
	logger  *zap.Logger
}

// NewScanHandler creates a new scan handler
func NewScanHandler(service service.ScanService, logger *zap.Logger) *ScanHandler {
	return &ScanHandler{
		service: service,
		logger:  logger,
	}
}

// CreateScan creates a new scan
// @Summary Create a new scan
// @Description Submit a barcode scan with device information
// @Tags Scans
// @Accept json
// @Produce json
// @Param request body dto.CreateScanRequest true "Scan request"
// @Success 201 {object} map[string]interface{} "Scan created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scans [post]
func (h *ScanHandler) CreateScan(c *gin.Context) {
	var req dto.CreateScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create scan request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Extract user ID from context
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDUint := userID.(uint)

	// Create scan
	response, err := h.service.CreateScan(c.Request.Context(), userIDUint, &req)
	if err != nil {
		h.logger.Error("Failed to create scan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// GetScan retrieves a scan by ID
// @Summary Get a scan by ID
// @Description Retrieve scan details by ID
// @Tags Scans
// @Accept json
// @Produce json
// @Param id path uint true "Scan ID"
// @Success 200 {object} map[string]interface{} "Scan details"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "Scan not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scans/{id} [get]
func (h *ScanHandler) GetScan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid scan ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scan ID"})
		return
	}

	response, err := h.service.GetScanByID(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "scan not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
			return
		}
		h.logger.Error("Failed to get scan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetScans retrieves all scans with pagination
// @Summary Get all scans
// @Description List scans with pagination and filtering
// @Tags Scans
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param status query string false "Scan status (pending, completed, failed)"
// @Success 200 {object} dto.ScanListResponse "List of scans"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scans [get]
func (h *ScanHandler) GetScans(c *gin.Context) {
	filter := &dto.ScanFilter{
		Page:     1,
		PageSize: 10,
	}

	// Parse query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			h.logger.Warn("Invalid page parameter", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
			return
		}
		filter.Page = page
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			h.logger.Warn("Invalid page_size parameter", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page_size parameter"})
			return
		}
		filter.PageSize = pageSize
	}

	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	response, err := h.service.GetScans(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get scans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserScans retrieves scans for a specific user
// @Summary Get user scans
// @Description List scans for a specific user with pagination
// @Tags Scans
// @Accept json
// @Produce json
// @Param user_id path uint true "User ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} dto.ScanListResponse "List of user scans"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/users/{user_id}/scans [get]
func (h *ScanHandler) GetUserScans(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid user ID", zap.String("user_id", userIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	filter := &dto.ScanFilter{
		Page:     1,
		PageSize: 10,
	}

	// Parse query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			h.logger.Warn("Invalid page parameter", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
			return
		}
		filter.Page = page
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			h.logger.Warn("Invalid page_size parameter", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page_size parameter"})
			return
		}
		filter.PageSize = pageSize
	}

	response, err := h.service.GetUserScans(c.Request.Context(), uint(userID), filter)
	if err != nil {
		h.logger.Error("Failed to get user scans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateScanStatus updates the status of a scan
// @Summary Update scan status
// @Description Update the status of a scan (pending, completed, failed)
// @Tags Scans
// @Accept json
// @Produce json
// @Param id path uint true "Scan ID"
// @Param request body dto.UpdateScanRequest true "Status update request"
// @Success 200 {object} map[string]interface{} "Scan status updated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Scan not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scans/{id} [put]
func (h *ScanHandler) UpdateScanStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid scan ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scan ID"})
		return
	}

	var req dto.UpdateScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update scan request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := h.service.UpdateScanStatus(c.Request.Context(), uint(id), &req)
	if err != nil {
		if err.Error() == "scan not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
			return
		}
		h.logger.Error("Failed to update scan status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// DeleteScan deletes a scan
// @Summary Delete a scan
// @Description Delete a scan by ID
// @Tags Scans
// @Accept json
// @Produce json
// @Param id path uint true "Scan ID"
// @Success 200 {object} map[string]interface{} "Scan deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "Scan not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scans/{id} [delete]
func (h *ScanHandler) DeleteScan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid scan ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scan ID"})
		return
	}

	if err := h.service.DeleteScan(c.Request.Context(), uint(id)); err != nil {
		if err.Error() == "scan not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
			return
		}
		h.logger.Error("Failed to delete scan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scan deleted successfully"})
}

// RegisterRoutes registers scan routes
func (h *ScanHandler) RegisterRoutes(api *gin.RouterGroup) {
	scans := api.Group("/scans")
	{
		scans.POST("", h.CreateScan)
		scans.GET("", h.GetScans)
		scans.GET("/:id", h.GetScan)
		scans.PUT("/:id", h.UpdateScanStatus)
		scans.DELETE("/:id", h.DeleteScan)
	}

	// User scans
	api.GET("/user/:user_id/scans", h.GetUserScans)
}
