package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/moodly/internal/application/auth/dto"
	"github.com/novriyantoAli/moodly/internal/application/auth/usecase"
	"github.com/novriyantoAli/moodly/internal/shared/apperror"
	"github.com/novriyantoAli/moodly/internal/shared/response"
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

	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := apperror.ToHTTP(err)
		c.JSON(status, response.Response{
			Success: false,
			Error:   resp,
		})

		return
	}

	req.IPAddress = c.ClientIP()

	req.UserAgent = c.Request.UserAgent()

	resp, err := h.loginUseCase.Execute(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error(
			"login failed",
			zap.Error(err),
			zap.String(
				"username",
				req.Username,
			),
		)

		status, resp := apperror.ToHTTP(err)
		c.JSON(status, response.Response{
			Success: false,
			Error:   resp,
		})
		return
	}

	c.JSON(http.StatusOK, response.Success(resp))
}

func (h *LoginHandler) RegisterRoutes(api *gin.RouterGroup) {
	login := api.Group("/login")
	{
		login.POST("", h.Login)
	}
}
