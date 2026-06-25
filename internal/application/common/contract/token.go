package contract

import (
	"github.com/novriyantoAli/moodly/internal/pkg/jwt"
)

type TokenService interface {
	GenerateToken(userID uint, email, level string, roles []string) (string, error)
	GenerateRefreshToken(userID uint, email, level string, roles []string) (string, error)
	ValidateToken(token string) (*jwt.Claims, error)
	ValidateRefreshToken(token string) (*jwt.Claims, error)
}
