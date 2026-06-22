package dto

type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiredAt    int64  `json:"expired_at"`
	UserID       uint   `json:"user_id"`
}
