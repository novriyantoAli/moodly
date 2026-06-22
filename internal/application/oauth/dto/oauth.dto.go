package dto

// OAuthProvider represents supported OAuth providers
type OAuthProvider string

const (
	GoogleProvider    OAuthProvider = "google"
	GithubProvider    OAuthProvider = "github"
	GitlabProvider    OAuthProvider = "gitlab"
	MicrosoftProvider OAuthProvider = "microsoft"
)

// OAuthAuthorizationRequest represents the authorization request
type OAuthAuthorizationRequest struct {
	Provider OAuthProvider `json:"provider" binding:"required"`
	Code     string        `json:"code" binding:"required"`
	State    string        `json:"state" binding:"required"`
}

// OAuthCallbackRequest represents the callback from OAuth provider
type OAuthCallbackRequest struct {
	Provider OAuthProvider `json:"provider" binding:"required"`
	Code     string        `json:"code" binding:"required"`
}

// OAuthTokenResponse represents the token response
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	IDToken      string `json:"id_token,omitempty"`
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider"`
}

type OAuthLoginResponse struct {
	Token    string        `json:"token"`
	UserInfo OAuthUserInfo `json:"user_info"`
}

// OAuthAuthorizationURLRequest represents the request to generate authorization URL
type OAuthAuthorizationURLRequest struct {
	Provider    OAuthProvider `json:"provider" binding:"required"`
	RedirectURI string        `json:"redirect_uri" binding:"required"`
}

// OAuthAuthorizationURLResponse represents the authorization URL response
type OAuthAuthorizationURLResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
}

// OAuthTokenRequest represents the request to exchange code for token
type OAuthTokenRequest struct {
	Provider OAuthProvider `json:"provider" binding:"required"`
	Code     string        `json:"code" binding:"required"`
}
