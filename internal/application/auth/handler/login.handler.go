package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/application/auth/usecase"
	"go.uber.org/zap"
)

type LoginHandler struct {
	loginUseCase usecase.LoginUseCase
	logger       *zap.Logger
}

func NewLoginHandler(
	loginUseCase usecase.LoginUseCase,
	logger *zap.Logger,
) *LoginHandler {

	return &LoginHandler{
		loginUseCase: loginUseCase,
		logger:       logger,
	}
}

func (h *LoginHandler) Login(c *gin.Context) {

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(
		&req,
	); err != nil {

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

	resp, err := h.loginUseCase.Execute(
		c.Request.Context(),
		&req,
	)

	if err != nil {

		h.logger.Error(
			"login failed",
			zap.Error(err),
			zap.String(
				"username",
				req.Username,
			),
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
			"message": "login success",
			"data":    resp,
		},
	)
}

func (h *LoginHandler) RegisterRoutes(api *gin.RouterGroup) {
	login := api.Group("/login")
	{
		login.POST("", h.Login)
	}
}
