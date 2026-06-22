package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/application/auth/usecase"
	"go.uber.org/zap"
)

type LogoutHandler struct {
	useCase usecase.LogoutUseCase
	logger  *zap.Logger
}

func NewLogoutHandler(
	useCase usecase.LogoutUseCase,
	logger *zap.Logger,
) *LogoutHandler {

	return &LogoutHandler{
		useCase: useCase,
		logger:  logger,
	}
}

func (h *LogoutHandler) Logout(c *gin.Context) {

	var req dto.LogoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"success": false,
				"message": "invalid request",
			},
		)

		return
	}

	err := h.useCase.Execute(
		c.Request.Context(),
		&req,
	)

	if err != nil {

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
			"message": "logout success",
		},
	)
}

func (h *LogoutHandler) RegisterRoutes(api *gin.RouterGroup) {
	api.POST("/logout", h.Logout)
}
