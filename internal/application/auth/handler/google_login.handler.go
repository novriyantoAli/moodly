package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/application/auth/usecase"
	"go.uber.org/zap"
)

type GoogleLoginHandler struct {
	googleLoginUseCase usecase.GoogleLoginUseCase
	logger             *zap.Logger
}

func NewGoogleLoginHandler(
	googleLoginUseCase usecase.GoogleLoginUseCase,
	logger *zap.Logger,
) *GoogleLoginHandler {

	return &GoogleLoginHandler{
		googleLoginUseCase: googleLoginUseCase,
		logger:             logger,
	}
}

func (h *GoogleLoginHandler) GoogleLogin(c *gin.Context) {

	var req dto.GoogleLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"success": false,
				"message": "invalid request",
				"error":   err.Error(),
			},
		)

		return
	}

	req.IPAddress = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	resp, err := h.googleLoginUseCase.Execute(
		c.Request.Context(),
		&req,
	)

	if err != nil {

		h.logger.Error(
			"google login failed",
			zap.Error(err),
		)

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"success": false,
				"message": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"message": "google login success",
			"data":    resp,
		},
	)
}

func (h *GoogleLoginHandler) RegisterRoutes(api *gin.RouterGroup) {
	google := api.Group("/google")
	{
		google.POST("/login", h.GoogleLogin)
	}
}
