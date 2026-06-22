package oauth

import (
	"github.com/novriyantoAli/moodly/internal/application/oauth/handler"
	"github.com/novriyantoAli/moodly/internal/application/oauth/service"
	"github.com/novriyantoAli/moodly/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides OAuth-related dependencies
var Module = fx.Options(
	fx.Provide(
		NewOAuthConfig,
		service.NewOAuthService,
		handler.NewOAuthHandler,
	),
)

// NewOAuthConfig creates a new OAuth configuration from config.Config
func NewOAuthConfig(cfg *config.Config, logger *zap.Logger) service.OAuthConfig {
	logger.Info("Initializing OAuth configuration from config")

	return service.OAuthConfig{
		Google: service.GoogleOAuthConfig{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			RedirectURI:  cfg.OAuth.Google.RedirectURL,
			Scopes:       []string{"openid", "profile", "email"},
		},
		Github: service.GithubOAuthConfig{
			ClientID:     "your-github-client-id",
			ClientSecret: "your-github-client-secret",
			RedirectURI:  "http://localhost:8080/api/v1/oauth/callback/github",
			Scopes:       []string{"user:email"},
		},
		Gitlab: service.GitlabOAuthConfig{
			ClientID:     "your-gitlab-client-id",
			ClientSecret: "your-gitlab-client-secret",
			RedirectURI:  "http://localhost:8080/api/v1/oauth/callback/gitlab",
			Scopes:       []string{"openid", "profile", "email"},
		},
		Microsoft: service.MicrosoftOAuthConfig{
			ClientID:     "your-microsoft-client-id",
			ClientSecret: "your-microsoft-client-secret",
			RedirectURI:  "http://localhost:8080/api/v1/oauth/callback/microsoft",
			Scopes:       []string{"openid", "profile", "email"},
			TenantID:     "common",
		},
	}
}
