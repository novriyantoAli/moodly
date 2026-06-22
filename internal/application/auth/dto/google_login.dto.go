package dto

type GoogleLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiredAt    int64  `json:"expired_at"`
}

type GoogleLoginRequest struct {
	IDToken   string `json:"id_token"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}
