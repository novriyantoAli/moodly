package contract

type TokenService interface {
	GenerateToken(
		userID uint,
		email string,
		level string,
	) (string, error)

	GenerateRefreshToken(
		userID uint,
		email string,
		level string,
	) (string, error)
}
