package handler

import (
	"net/http"
	"strings"

	"github.com/novriyantoAli/moodly/internal/application/oauth/dto"
	"github.com/novriyantoAli/moodly/internal/application/oauth/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OAuthHandler handles OAuth-related HTTP requests
type OAuthHandler struct {
	service service.OAuthService
	logger  *zap.Logger
}

// NewOAuthHandler creates a new OAuth handler instance
func NewOAuthHandler(service service.OAuthService, logger *zap.Logger) *OAuthHandler {
	return &OAuthHandler{
		service: service,
		logger:  logger,
	}
}

// GetAuthorizationURL godoc
// @Summary Get OAuth authorization URL
// @Description Get the authorization URL for the specified OAuth provider
// @Tags oauth
// @Accept json
// @Produce json
// @Param request body dto.OAuthAuthorizationURLRequest true "Authorization URL request"
// @Success 200 {object} dto.OAuthAuthorizationURLResponse "Authorization URL"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /oauth/authorization-url [post]
func (h *OAuthHandler) GetAuthorizationURL(ctx *gin.Context) {
	var req dto.OAuthAuthorizationURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authURL, err := h.service.GetAuthorizationURL(ctx.Request.Context(), req.Provider, req.RedirectURI)
	if err != nil {
		h.logger.Error("Failed to get authorization URL", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, authURL)
}

// ExchangeCodeForToken godoc
// @Summary Exchange authorization code for token
// @Description Exchange the authorization code from OAuth provider for access token
// @Tags oauth
// @Accept json
// @Produce json
// @Param request body dto.OAuthTokenRequest true "Token exchange request"
// @Success 200 {object} dto.OAuthTokenResponse "Access token"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /oauth/token [post]
func (h *OAuthHandler) ExchangeCodeForToken(ctx *gin.Context) {
	var req dto.OAuthTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Code == "" {
		h.logger.Error("Authorization code is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
		return
	}

	tokenResp, err := h.service.ExchangeCodeForToken(ctx.Request.Context(), req.Provider, req.Code)
	if err != nil {
		h.logger.Error("Failed to exchange code for token", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tokenResp)
}

// GetUserInfo godoc
// @Summary Get OAuth user information
// @Description Get user information from OAuth provider using access token
// @Tags oauth
// @Accept json
// @Produce json
// @Param provider query string true "OAuth provider (google, github, gitlab, microsoft)"
// @Param access_token query string true "Access token from OAuth provider"
// @Success 200 {object} dto.OAuthUserInfo "User information"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /oauth/user-info [get]
func (h *OAuthHandler) GetUserInfo(ctx *gin.Context) {
	provider := ctx.Query("provider")
	accessToken := ctx.Query("access_token")

	if provider == "" {
		h.logger.Error("Provider is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	if accessToken == "" {
		h.logger.Error("Access token is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "access token is required"})
		return
	}

	userInfo, err := h.service.GetUserInfo(ctx.Request.Context(), dto.OAuthProvider(provider), accessToken)
	if err != nil {
		h.logger.Error("Failed to get user info", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, userInfo)
}

// Authenticate godoc
// @Summary Authenticate with OAuth
// @Description Complete OAuth authentication flow - exchange code and get user info
// @Tags oauth
// @Accept json
// @Produce json
// @Param request body dto.OAuthAuthorizationRequest true "Authentication request"
// @Success 200 {object} dto.OAuthUserInfo "User information"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /oauth/authenticate [post]
func (h *OAuthHandler) Authenticate(ctx *gin.Context) {
	var req dto.OAuthAuthorizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userInfo, err := h.service.Authenticate(ctx.Request.Context(), req.Provider, req.Code, req.State)
	if err != nil {
		h.logger.Error("Authentication failed", zap.Error(err))
		// Return 401 for invalid/expired state token
		if strings.Contains(err.Error(), "state token") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": userInfo})
}

// GetCurrentUser godoc
// @Summary Get current user details
// @Description Get the current authenticated user's details using JWT token
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Current user details"
// @Failure 400 {object} map[string]interface{} "Missing or invalid token"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid token"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/me [get]
func (h *OAuthHandler) GetCurrentUser(ctx *gin.Context) {
	// Extract token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn("Missing authorization header")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	// Extract token from "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		h.logger.Warn("Invalid authorization header format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		h.logger.Warn("Empty token in authorization header")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	user, err := h.service.GetCurrentUser(ctx.Request.Context(), token)
	if err != nil {
		h.logger.Error("Failed to get current user", zap.Error(err))
		if err.Error() == "invalid or expired token" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh the access token using refresh token
// @Tags oauth
// @Accept json
// @Produce json
// @Param request body dto.OAuthAuthorizationRequest true "Refresh token request"
// @Success 200 {object} dto.OAuthTokenResponse "New access token"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /oauth/refresh [post]
func (h *OAuthHandler) RefreshToken(ctx *gin.Context) {
	var req dto.OAuthAuthorizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenResp, err := h.service.RefreshToken(ctx.Request.Context(), req.Provider, req.Code)
	if err != nil {
		h.logger.Error("Failed to refresh token", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tokenResp)
}

// RegisterRoutes registers OAuth routes
func (h *OAuthHandler) RegisterRoutes(router *gin.Engine) {
	oauthGroup := router.Group("/api/v1/oauth")
	{
		oauthGroup.POST("/authorization-url", h.GetAuthorizationURL)
		oauthGroup.POST("/token", h.ExchangeCodeForToken)
		oauthGroup.GET("/user-info", h.GetUserInfo)
		oauthGroup.POST("/authenticate", h.Authenticate)
		oauthGroup.POST("/refresh", h.RefreshToken)
		oauthGroup.GET("/me", h.GetCurrentUser)
	}
}
