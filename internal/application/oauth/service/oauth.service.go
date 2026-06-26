package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/common/contract"
	"github.com/novriyantoAli/moodly/internal/application/oauth/dto"
	authService "github.com/novriyantoAli/moodly/internal/application/authorization/service"
	"github.com/novriyantoAli/moodly/internal/application/user/entity"
	userRepo "github.com/novriyantoAli/moodly/internal/application/user/repository"
	userService "github.com/novriyantoAli/moodly/internal/application/user/service"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
	"gorm.io/gorm"
)

// OAuthConfig represents OAuth provider configuration
type OAuthConfig struct {
	Google    GoogleOAuthConfig
	Github    GithubOAuthConfig
	Gitlab    GitlabOAuthConfig
	Microsoft MicrosoftOAuthConfig
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

type GithubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

type GitlabOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

type MicrosoftOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	TenantID     string
}

// OAuthService defines the interface for OAuth operations
type OAuthService interface {
	// GetAuthorizationURL generates the authorization URL for a given provider
	GetAuthorizationURL(ctx context.Context, provider dto.OAuthProvider, redirectURI string) (*dto.OAuthAuthorizationURLResponse, error)

	// ExchangeCodeForToken exchanges authorization code for tokens
	ExchangeCodeForToken(ctx context.Context, provider dto.OAuthProvider, code string) (*dto.OAuthTokenResponse, error)

	// GetUserInfo retrieves user information from OAuth provider using access token
	GetUserInfo(ctx context.Context, provider dto.OAuthProvider, accessToken string) (*dto.OAuthUserInfo, error)

	// GetCurrentUser retrieves the current user based on the provided token
	GetCurrentUser(ctx context.Context, token string) (*entity.User, error)

	// Authenticate handles the complete OAuth authentication flow and returns login response with JWT token
	// State parameter is verified to prevent CSRF attacks
	Authenticate(ctx context.Context, provider dto.OAuthProvider, code string, state string) (*dto.OAuthLoginResponse, error)

	// RefreshToken refreshes the access token using refresh token
	RefreshToken(ctx context.Context, provider dto.OAuthProvider, refreshToken string) (*dto.OAuthTokenResponse, error)
}

type oauthService struct {
	config        OAuthConfig
	logger        *zap.Logger
	oauth2Configs map[dto.OAuthProvider]*oauth2.Config
	userRepo      userRepo.UserRepository
	authSvc       authService.AuthorizationService
	jwtManager    contract.TokenService
	userService   userService.UserService
}

// NewOAuthService creates a new OAuth service instance
func NewOAuthService(config OAuthConfig, logger *zap.Logger, userRepo userRepo.UserRepository, authSvc authService.AuthorizationService, userService userService.UserService, jwtManager contract.TokenService) OAuthService {
	service := &oauthService{
		config:        config,
		logger:        logger,
		userRepo:      userRepo,
		authSvc:       authSvc,
		userService:   userService,
		jwtManager:    jwtManager,
		oauth2Configs: make(map[dto.OAuthProvider]*oauth2.Config),
	}

	// Initialize oauth2 configs for each provider
	service.oauth2Configs[dto.GoogleProvider] = service.buildGoogleConfig()
	service.oauth2Configs[dto.GithubProvider] = service.buildGithubConfig()
	service.oauth2Configs[dto.GitlabProvider] = service.buildGitlabConfig()
	service.oauth2Configs[dto.MicrosoftProvider] = service.buildMicrosoftConfig()

	return service
}

// GetAuthorizationURL generates the OAuth authorization URL
func (s *oauthService) GetAuthorizationURL(ctx context.Context, provider dto.OAuthProvider, redirectURI string) (*dto.OAuthAuthorizationURLResponse, error) {
	// Generate state as a JWT token for security verification
	state, err := s.jwtManager.GenerateToken(0, "", "oauth_state", nil)
	if err != nil {
		s.logger.Error("Failed to generate state token", zap.Error(err))
		return nil, err
	}

	// Get the oauth2 config for this provider
	oauth2Config, exists := s.oauth2Configs[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Update redirect URI if provided (allows dynamic redirect URIs)
	if redirectURI != "" {
		oauth2Config.RedirectURL = redirectURI
	}

	// Generate authorization URL
	authURL := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	s.logger.Info("Generated authorization URL", zap.String("provider", string(provider)))

	return &dto.OAuthAuthorizationURLResponse{
		AuthorizationURL: authURL,
		State:            state,
	}, nil
}

// ExchangeCodeForToken exchanges the authorization code for tokens
func (s *oauthService) ExchangeCodeForToken(ctx context.Context, provider dto.OAuthProvider, code string) (*dto.OAuthTokenResponse, error) {
	if code == "" {
		return nil, errors.New("authorization code is required")
	}

	s.logger.Info("Exchanging code for token", zap.String("provider", string(provider)))

	// Get the oauth2 config for this provider
	oauth2Config, exists := s.oauth2Configs[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Exchange the code for a token
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		s.logger.Error("Failed to exchange code for token", zap.String("provider", string(provider)), zap.Error(err))
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Extract refresh token (may be empty for some providers)
	refreshToken := ""
	if token.RefreshToken != "" {
		refreshToken = token.RefreshToken
	}

	// Calculate expiration
	expiresIn := int64(3600) // Default 1 hour
	if !token.Expiry.IsZero() {
		expiresIn = int64(token.Expiry.Sub(time.Now()).Seconds())
	}

	s.logger.Info("Successfully exchanged code for token", zap.String("provider", string(provider)))

	return &dto.OAuthTokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		IDToken:      token.Extra("id_token").(string), // For OIDC providers
	}, nil
}

// GetUserInfo retrieves user information from the OAuth provider
func (s *oauthService) GetUserInfo(ctx context.Context, provider dto.OAuthProvider, accessToken string) (*dto.OAuthUserInfo, error) {
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	s.logger.Info("Fetching user info", zap.String("provider", string(provider)))

	var userInfo *dto.OAuthUserInfo
	var err error

	switch provider {
	case dto.GoogleProvider:
		userInfo, err = s.getGoogleUserInfo(ctx, accessToken)
	case dto.GithubProvider:
		userInfo, err = s.getGithubUserInfo(ctx, accessToken)
	case dto.GitlabProvider:
		userInfo, err = s.getGitlabUserInfo(ctx, accessToken)
	case dto.MicrosoftProvider:
		userInfo, err = s.getMicrosoftUserInfo(ctx, accessToken)
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	if err != nil {
		s.logger.Error("Failed to fetch user info", zap.String("provider", string(provider)), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	userInfo.Provider = string(provider)
	s.logger.Info("Successfully fetched user info", zap.String("provider", string(provider)), zap.String("user_id", userInfo.ID))

	return userInfo, nil
}

func (s *oauthService) GetCurrentUser(ctx context.Context, token string) (*entity.User, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		s.logger.Warn("Invalid token provided", zap.Error(err))
		return nil, errors.New("invalid or expired token")
	}

	// Retrieve user by ID from claims
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("User not found for token", zap.Uint("user_id", claims.UserID))
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to retrieve user", zap.Error(err))
		return nil, err
	}

	return user, nil
}

// Authenticate performs the complete OAuth authentication flow and creates/updates user
func (s *oauthService) Authenticate(ctx context.Context, provider dto.OAuthProvider, code string, state string) (*dto.OAuthLoginResponse, error) {
	// Check authorization code first
	if code == "" {
		return nil, errors.New("authorization code is required")
	}

	// Verify state token to prevent CSRF attacks
	if state == "" {
		return nil, errors.New("state parameter is required")
	}

	// Verify state token
	claims, err := s.jwtManager.ValidateToken(state)
	if err != nil {
		s.logger.Error("Invalid or expired state token", zap.Error(err))
		return nil, fmt.Errorf("invalid or expired state token: %w", err)
	}

	// Verify state token was generated for oauth flow (Level = "oauth_state")
	if claims.Level != "oauth_state" {
		return nil, errors.New("invalid state token: not an OAuth state token")
	}

	tokenResp, err := s.ExchangeCodeForToken(ctx, provider, code)
	if err != nil {
		s.logger.Error("Failed to exchange code for token", zap.Error(err))
		return nil, err
	}

	userInfo, err := s.GetUserInfo(ctx, provider, tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Failed to get user info", zap.Error(err))
		return nil, err
	}

	// Create or update user in database
	user, err := s.createOrUpdateUser(ctx, provider, userInfo)
	if err != nil {
		s.logger.Error("Failed to create or update user", zap.Error(err))
		return nil, err
	}

	// Ambil roles dan permissions
	rolesEntities, err := s.authSvc.GetRolesByUserID(ctx, user.ID)
	if err != nil {
		s.logger.Warn("Failed to fetch roles", zap.Error(err))
	}
	var roleNames []string
	for _, r := range rolesEntities {
		roleNames = append(roleNames, r.Name)
	}

	permissions, err := s.authSvc.GetPermissionsByRoles(ctx, roleNames)
	if err != nil {
		s.logger.Warn("Failed to fetch permissions", zap.Error(err))
	}
	if permissions == nil {
		permissions = []string{}
	}
	if roleNames == nil {
		roleNames = []string{}
	}

	// Generate JWT token
	accessToken, err := s.jwtManager.GenerateToken(
		user.ID,
		user.Email,
		user.Level,
		roleNames,
	)
	if err != nil {
		s.logger.Error("Failed to generate JWT token", zap.Error(err))
		return nil, err
	}

	// Buat refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(
		user.ID,
		user.Email,
		user.Level,
		roleNames,
	)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	s.logger.Info("OAuth authentication successful",
		zap.String("provider", string(provider)),
		zap.String("user_id", userInfo.ID),
		zap.Uint("db_user_id", user.ID))

	return &dto.OAuthLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiredAt:    time.Now().Add(24 * time.Hour).Unix(), // Replace with actual expiration
		UserID:       user.ID,
		IsNewUser:    false, // Optional enhancement: detect if user was just created
		Roles:        roleNames,
		Permissions:  permissions,
	}, nil
}

// RefreshToken refreshes the access token using the refresh token
func (s *oauthService) RefreshToken(ctx context.Context, provider dto.OAuthProvider, refreshToken string) (*dto.OAuthTokenResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	s.logger.Info("Refreshing token", zap.String("provider", string(provider)))

	// Get the oauth2 config for this provider
	oauth2Config, exists := s.oauth2Configs[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Create token source with the refresh token
	tokenSource := oauth2Config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	// Get a new token
	newToken, err := tokenSource.Token()
	if err != nil {
		s.logger.Error("Failed to refresh token", zap.String("provider", string(provider)), zap.Error(err))
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Calculate expiration
	expiresIn := int64(3600)
	if !newToken.Expiry.IsZero() {
		expiresIn = int64(newToken.Expiry.Sub(time.Now()).Seconds())
	}

	s.logger.Info("Successfully refreshed token", zap.String("provider", string(provider)))

	return &dto.OAuthTokenResponse{
		AccessToken:  newToken.AccessToken,
		TokenType:    newToken.TokenType,
		ExpiresIn:    expiresIn,
		RefreshToken: newToken.RefreshToken,
	}, nil
}

// getGoogleUserInfo fetches user info from Google
func (s *oauthService) getGoogleUserInfo(ctx context.Context, accessToken string) (*dto.OAuthUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google API returned status %d: %s", resp.StatusCode, string(body))
	}

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, err
	}

	return &dto.OAuthUserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
	}, nil
}

// getGithubUserInfo fetches user info from GitHub
func (s *oauthService) getGithubUserInfo(ctx context.Context, accessToken string) (*dto.OAuthUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned status %d: %s", resp.StatusCode, string(body))
	}

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// GitHub might not return email in the main user endpoint, fetch it separately if needed
	email := githubUser.Email
	if email == "" {
		email, _ = s.getGithubUserEmail(ctx, accessToken)
	}

	return &dto.OAuthUserInfo{
		ID:        fmt.Sprintf("%d", githubUser.ID),
		Email:     email,
		Name:      githubUser.Name,
		AvatarURL: githubUser.AvatarURL,
	}, nil
}

// getGithubUserEmail fetches the primary email from GitHub
func (s *oauthService) getGithubUserEmail(ctx context.Context, accessToken string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", nil
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", nil
}

// getGitlabUserInfo fetches user info from GitLab
func (s *oauthService) getGitlabUserInfo(ctx context.Context, accessToken string) (*dto.OAuthUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab API returned status %d: %s", resp.StatusCode, string(body))
	}

	var gitlabUser struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gitlabUser); err != nil {
		return nil, err
	}

	return &dto.OAuthUserInfo{
		ID:        fmt.Sprintf("%d", gitlabUser.ID),
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		AvatarURL: gitlabUser.AvatarURL,
	}, nil
}

// getMicrosoftUserInfo fetches user info from Microsoft
func (s *oauthService) getMicrosoftUserInfo(ctx context.Context, accessToken string) (*dto.OAuthUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("microsoft API returned status %d: %s", resp.StatusCode, string(body))
	}

	var microsoftUser struct {
		ID          string `json:"id"`
		Mail        string `json:"mail"`
		DisplayName string `json:"displayName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&microsoftUser); err != nil {
		return nil, err
	}

	return &dto.OAuthUserInfo{
		ID:        microsoftUser.ID,
		Email:     microsoftUser.Mail,
		Name:      microsoftUser.DisplayName,
		AvatarURL: "", // Microsoft Graph doesn't return avatar in /me endpoint by default
	}, nil
}

// createOrUpdateUser creates a new user or updates existing user
func (s *oauthService) createOrUpdateUser(ctx context.Context, provider dto.OAuthProvider, userInfo *dto.OAuthUserInfo) (*entity.User, error) {
	// Try to find existing user by email
	existingUser, err := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err == nil && existingUser != nil {
		// User exists, update it
		existingUser.FullName = userInfo.Name
		existingUser.IsActive = true
		err := s.userRepo.Update(ctx, existingUser)
		if err != nil {
			s.logger.Error("Failed to update user", zap.String("email", userInfo.Email), zap.Error(err))
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		s.logger.Info("User updated via OAuth", zap.String("email", userInfo.Email), zap.String("provider", string(provider)))
		return existingUser, nil
	}

	// User doesn't exist, create a new one
	newUser := &entity.User{
		Email:    userInfo.Email,
		FullName: userInfo.Name,
		Level:    "user",
		IsActive: true,
	}

	err = s.userRepo.Create(ctx, newUser)
	if err != nil {
		s.logger.Error("Failed to create user", zap.String("email", userInfo.Email), zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created via OAuth", zap.String("email", userInfo.Email), zap.String("provider", string(provider)))
	return newUser, nil
}

// Helper methods to build oauth2 configs for each provider

func (s *oauthService) buildGoogleConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.config.Google.ClientID,
		ClientSecret: s.config.Google.ClientSecret,
		RedirectURL:  s.config.Google.RedirectURI,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
}

func (s *oauthService) buildGithubConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.config.Github.ClientID,
		ClientSecret: s.config.Github.ClientSecret,
		RedirectURL:  s.config.Github.RedirectURI,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

func (s *oauthService) buildGitlabConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.config.Gitlab.ClientID,
		ClientSecret: s.config.Gitlab.ClientSecret,
		RedirectURL:  s.config.Gitlab.RedirectURI,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     gitlab.Endpoint,
	}
}

func (s *oauthService) buildMicrosoftConfig() *oauth2.Config {
	// Microsoft uses a different endpoint based on tenant ID
	microsoftEndpoint := microsoft.AzureADEndpoint(s.config.Microsoft.TenantID)

	return &oauth2.Config{
		ClientID:     s.config.Microsoft.ClientID,
		ClientSecret: s.config.Microsoft.ClientSecret,
		RedirectURL:  s.config.Microsoft.RedirectURI,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     microsoftEndpoint,
	}
}

func (s *oauthService) buildGoogleAuthURL(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		s.config.Google.ClientID,
		redirectURI,
		"openid profile email",
		state,
	)
}

func (s *oauthService) buildGithubAuthURL(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		s.config.Github.ClientID,
		redirectURI,
		"user:email",
		state,
	)
}

func (s *oauthService) buildGitlabAuthURL(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://gitlab.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		s.config.Gitlab.ClientID,
		redirectURI,
		"openid profile email",
		state,
	)
}

func (s *oauthService) buildMicrosoftAuthURL(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		s.config.Microsoft.TenantID,
		s.config.Microsoft.ClientID,
		redirectURI,
		"openid profile email",
		state,
	)
}

func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
