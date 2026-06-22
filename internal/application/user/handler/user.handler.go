package handler

import (
	"net/http"
	"strconv"

	"github.com/novriyantoAli/moodly/internal/application/user/dto"
	"github.com/novriyantoAli/moodly/internal/application/user/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service service.UserService
	logger  *zap.Logger
}

func NewUserHandler(service service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User creation request"
// @Success 201 {object} map[string]interface{} "Created user"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Email already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users [post]
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.CreateUser(ctx.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		if err.Error() == "email already exists" {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": user})
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Get a single user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User details"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.service.GetUserByID(ctx.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

// GetUsers godoc
// @Summary Get all users
// @Description Get a list of users with optional filtering and pagination
// @Tags users
// @Accept json
// @Produce json
// @Param name query string false "Filter by name"
// @Param email query string false "Filter by email"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} dto.UserListResponse "List of users"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetUsers(ctx *gin.Context) {
	var filter dto.UserFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := h.service.GetUsers(ctx.Request.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to get users", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update a user's information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User update request"
// @Success 200 {object} map[string]interface{} "Updated user"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 409 {object} map[string]interface{} "Email already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.UpdateUser(ctx.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "email already exists" {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/{id} [delete]
// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.service.DeleteUser(ctx.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Login godoc
// @Summary User login
// @Description Authenticate a user with email and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body dto.LoginUserRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with user details"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid email or password"
// @Failure 403 {object} map[string]interface{} "User account is inactive"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *UserHandler) Login(ctx *gin.Context) {
	var req dto.LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Login(ctx.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to login user", zap.String("email", req.Email), zap.Error(err))
		if err.Error() == "invalid email or password" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user account is inactive" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) RegisterRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		users.POST("", h.CreateUser)
		users.GET("", h.GetUsers)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}

	auth := api.Group("/auth")
	{
		auth.POST("/login", h.Login)
	}
}
