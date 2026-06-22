package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	service "github.com/novriyantoAli/moodly/internal/application/security/service"
	"go.uber.org/zap"
)

type UserPasswordHandler struct {
	service service.UserPasswordService
	logger  *zap.Logger
}

func NewUserPasswordHandler(service service.UserPasswordService, logger *zap.Logger) *UserPasswordHandler {
	return &UserPasswordHandler{
		service: service,
		logger:  logger,
	}
}

// SetPassword godoc
// @Summary Set user password
// @Description Create or update user password
// @Tags user-password
// @Accept json
// @Produce json
// @Param request body dto.SetPasswordRequest true "Set password request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/user-password [post]
func (h *UserPasswordHandler) SetPassword(c *gin.Context) {
	var req dto.SetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(
			"Failed to bind request",
			zap.Error(err),
		)

		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		return
	}

	if err := h.service.SetPassword(
		c.Request.Context(),
		&req,
	); err != nil {

		h.logger.Error(
			"Failed to set password",
			zap.Error(err),
		)

		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to set password"},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "password set successfully",
		},
	)
}

// VerifyPassword godoc
// @Summary Verify password
// @Description Verify username and password
// @Tags user-password
// @Accept json
// @Produce json
// @Param request body dto.VerifyPasswordRequest true "Verify password request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/user-password/verify [post]
func (h *UserPasswordHandler) VerifyPassword(
	c *gin.Context,
) {

	var req dto.VerifyPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		h.logger.Error(
			"Failed to bind request",
			zap.Error(err),
		)

		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		return
	}

	verified, err := h.service.VerifyPassword(
		c.Request.Context(),
		&req,
	)

	if err != nil {

		h.logger.Error(
			"Failed to verify password",
			zap.Error(err),
		)

		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to verify password"},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"verified": verified,
		},
	)
}

// ChangePassword godoc
// @Summary Change password
// @Description Change existing password
// @Tags user-password
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Change password request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/user-password/change [put]
func (h *UserPasswordHandler) ChangePassword(
	c *gin.Context,
) {

	var req dto.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)
		return
	}

	if err := h.service.ChangePassword(
		c.Request.Context(),
		&req,
	); err != nil {

		h.logger.Error(
			"Failed to change password",
			zap.Error(err),
		)

		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "password changed successfully",
		},
	)
}

// GetPasswordInfo godoc
// @Summary Get password info
// @Description Get password security information
// @Tags user-password
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} dto.UserPasswordResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/user-password [get]
func (h *UserPasswordHandler) GetPasswordInfo(
	c *gin.Context,
) {

	userIDStr := c.Query("user_id")

	userIDUint, err := strconv.ParseUint(
		userIDStr,
		10,
		32,
	)

	if err != nil || userIDUint == 0 {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "user_id is required",
			},
		)
		return
	}

	resp, err := h.service.GetPasswordInfo(
		c.Request.Context(),
		uint(userIDUint),
	)

	if err != nil {

		h.logger.Error(
			"Failed to get password info",
			zap.Error(err),
		)

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "failed to get password info",
			},
		)
		return
	}

	if resp == nil {

		c.JSON(
			http.StatusNoContent,
			gin.H{
				"error": "password not configured",
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		resp,
	)
}

func (h *UserPasswordHandler) RegisterRoutes(
	api *gin.RouterGroup,
) {

	password := api.Group("/user-password")

	{
		password.POST(
			"",
			h.SetPassword,
		)

		password.POST(
			"/verify",
			h.VerifyPassword,
		)

		password.PUT(
			"/change",
			h.ChangePassword,
		)

		password.GET(
			"",
			h.GetPasswordInfo,
		)
	}
}
