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
		statusCode, resp := apperror.ToHTTP(err)
		c.JSON(statusCode, response.Response{
			Success: false,
			Error:   resp,
		})

		return
	}

	err := h.useCase.Execute(c.Request.Context(), &req)
	if err != nil {
		statusCode, resp := apperror.ToHTTP(err)
		c.JSON(statusCode, response.Response{
			Success: false,
			Error:   resp,
		})
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (h *LogoutHandler) RegisterRoutes(api *gin.RouterGroup) {
	api.POST("/logout", h.Logout)
}
