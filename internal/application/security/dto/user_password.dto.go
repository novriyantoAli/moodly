package dto

type SetPasswordRequest struct {
	UserID   uint
	Username string
	Password string
}

type VerifyPasswordRequest struct {
	Username string
	Password string
}

type ChangePasswordRequest struct {
	UserID          uint
	CurrentPassword string
	NewPassword     string
	ConfirmPassword string
}

type UserPasswordResponse struct {
	UserID        uint
	Username      string
	FailedAttempt int
	IsLocked      bool
}
