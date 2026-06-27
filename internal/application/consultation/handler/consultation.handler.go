package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/consultation/dto"
	"github.com/novriyantoAli/moodly/internal/application/consultation/usecase"
	securityService "github.com/novriyantoAli/moodly/internal/application/authorization/service"
	"github.com/novriyantoAli/moodly/internal/middleware"
	"github.com/novriyantoAli/moodly/internal/security"
	"go.uber.org/zap"
)

type ConsultationHandler struct {
	usecase usecase.ConsultationUsecase
	authSvc securityService.AuthorizationService
	logger  *zap.Logger
}

func NewConsultationHandler(u usecase.ConsultationUsecase, authSvc securityService.AuthorizationService, l *zap.Logger) *ConsultationHandler {
	return &ConsultationHandler{usecase: u, authSvc: authSvc, logger: l}
}

func (h *ConsultationHandler) RegisterRoutes(api *gin.RouterGroup) {
	consultations := api.Group("/consultations")
	{
		consultations.POST(
			"",
			middleware.RequireRoles([]string{"atlit"}, h.logger),
			middleware.RequirePermission(
				h.authSvc,
				"consultation.create",
				h.logger,
			),
			h.CreateConsultation,
		)

		consultations.GET(
			"",
			middleware.RequireRoles([]string{"psikolog","atlit"}, h.logger), 
			h.GetConsultations,
		)
		consultations.GET(
			"/:id",
			middleware.RequireRoles([]string{"psikolog","atlit"}, h.logger),
			h.GetConsultationByID,
		)
		consultations.PATCH(
			"/:id/approve",
			middleware.RequireRoles([]string{"psikolog"}, h.logger),
			middleware.RequirePermission(
				h.authSvc,
				"consultation.approve",
				h.logger,
			),
			h.ApproveConsultation,
		)
		consultations.POST(
			"/:id/messages",
			middleware.RequireRoles([]string{"psikolog","atlit"}, h.logger), 
			h.SendMessage,
		)
		consultations.GET(
			"/:id/messages",
			middleware.RequireRoles([]string{"psikolog","atlit"}, h.logger),  
			h.GetMessages,
		)
		consultations.POST("/:id/read", h.MarkMessageRead)
		consultations.PATCH(
			"/:id/close",
			middleware.RequireRoles([]string{"psikolog","atlit"}, h.logger),  
			h.CloseConsultation,
		)
	}
}

// helper to get user ID from context, assuming it's stored by JWT middleware
func getUserID(c *gin.Context) (uint, bool) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if ok {
		return principal.UserID, true
	}
	
	userIDStr, exists := c.Get("user_id") // fallback
	if !exists {
		return 0, false
	}

	switch v := userIDStr.(type) {
	case string:
		id, _ := strconv.ParseUint(v, 10, 32)
		return uint(id), true
	case float64:
		return uint(v), true
	case uint:
		return v, true
	case int:
		return uint(v), true
	}
	return 0, false
}

// CreateConsultation godoc
// @Summary Create a new consultation
// @Description Create a new consultation with a psychologist
// @Tags consultations
// @Accept json
// @Produce json
// @Param request body dto.CreateConsultationRequest true "Create consultation request"
// @Success 201 {object} dto.CreateConsultationResponse "Created consultation"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /consultations [post]
func (h *ConsultationHandler) CreateConsultation(c *gin.Context) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := principal.UserID

	var req dto.CreateConsultationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.usecase.CreateConsultation(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("failed to create consultation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetConsultations godoc
// @Summary Get all consultations
// @Description Get all consultations for the authenticated user
// @Tags consultations
// @Accept json
// @Produce json
// @Success 200 {array} dto.ConsultationResponse "List of consultations"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /consultations [get]
func (h *ConsultationHandler) GetConsultations(c *gin.Context) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	res, err := h.usecase.GetConsultations(c.Request.Context(), principal.UserID)
	if err != nil {
		h.logger.Error("failed to get consultations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if res == nil {
		res = []dto.ConsultationResponse{}
	}

	c.JSON(http.StatusOK, res)
}

// GetConsultationByID godoc
// @Summary Get consultation by ID
// @Description Get details of a specific consultation by its ID
// @Tags consultations
// @Accept json
// @Produce json
// @Param id path string true "Consultation ID (UUID)"
// @Success 200 {object} dto.ConsultationResponse "Consultation details"
// @Failure 400 {object} map[string]interface{} "Invalid ID format"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Consultation not found"
// @Router /consultations/{id} [get]
func (h *ConsultationHandler) GetConsultationByID(c *gin.Context) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := principal.UserID

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	res, err := h.usecase.GetConsultationByID(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.Error("failed to get consultation", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// SendMessage godoc
// @Summary Send a message
// @Description Send a message in a specific consultation
// @Tags consultations
// @Accept json
// @Produce json
// @Param id path string true "Consultation ID (UUID)"
// @Param request body dto.SendMessageRequest true "Send message request"
// @Success 201 {object} dto.MessageResponse "Created message"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /consultations/{id}/messages [post]
func (h *ConsultationHandler) SendMessage(c *gin.Context) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := principal.UserID

	idParam := c.Param("id")
	conversationID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id format"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.usecase.SendMessage(c.Request.Context(), conversationID, userID, &req)
	if err != nil {
		h.logger.Error("failed to send message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetMessages godoc
// @Summary Get messages
// @Description Get paginated messages for a specific consultation
// @Tags consultations
// @Accept json
// @Produce json
// @Param id path string true "Consultation ID (UUID)"
// @Param cursor query string false "Pagination cursor (Message ID UUID)"
// @Param limit query int false "Pagination limit" default(20)
// @Success 200 {array} dto.MessageResponse "List of messages"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /consultations/{id}/messages [get]
func (h *ConsultationHandler) GetMessages(c *gin.Context) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := principal.UserID

	idParam := c.Param("id")
	conversationID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id format"})
		return
	}

	cursorParam := c.Query("cursor")
	var cursor uuid.UUID
	if cursorParam != "" {
		cursor, err = uuid.Parse(cursorParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor format"})
			return
		}
	}

	limitParam := c.Query("limit")
	limit := 20 // default limit
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	res, err := h.usecase.GetMessages(c.Request.Context(), conversationID, userID, cursor, limit)
	if err != nil {
		h.logger.Error("failed to get messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if res == nil {
		res = []dto.MessageResponse{}
	}

	c.JSON(http.StatusOK, res)
}

// MarkMessageRead godoc
// @Summary Mark a message as read
// @Description Mark a specific message as read by the user
// @Tags consultations
// @Accept json
// @Produce json
// @Param id path string true "Consultation ID (UUID)"
// @Param request body dto.MarkMessageReadRequest true "Mark read request"
// @Success 200 {object} dto.MessageResponse "Updated message details"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /consultations/{id}/read [post]
func (h *ConsultationHandler) MarkMessageRead(c *gin.Context) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := principal.UserID

	idParam := c.Param("id")
	conversationID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id format"})
		return
	}

	var req dto.MarkMessageReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.usecase.MarkMessageRead(c.Request.Context(), conversationID, userID, &req)
	if err != nil {
		h.logger.Error("failed to mark message as read", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// CloseConsultation godoc
// @Summary Close a consultation
// @Description Close an active consultation session
// @Tags consultations
// @Accept json
// @Produce json
// @Param id path string true "Consultation ID (UUID)"
// @Param request body dto.CloseConsultationRequest true "Close consultation request"
// @Success 200 {object} dto.CloseConsultationResponse "Closed consultation details"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /consultations/{id} [patch]
func (h *ConsultationHandler) CloseConsultation(c *gin.Context) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := principal.UserID

	idParam := c.Param("id")
	conversationID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id format"})
		return
	}

	res, err := h.usecase.CloseConsultation(c.Request.Context(), conversationID, userID)
	if err != nil {
		h.logger.Error("failed to close consultation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ConsultationHandler) ApproveConsultation(c *gin.Context) {
	principal, ok := security.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}	

	idParam := c.Param("id")
	conversationID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id format"})
		return
	}

	err = h.usecase.ApproveConsultation(c.Request.Context(), conversationID, principal.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}
