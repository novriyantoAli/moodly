package dto

type SetPINRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	PIN    string `json:"pin" binding:"required,min=4,max=6,numeric"`
}

type VerifyPINRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	PIN    string `json:"pin" binding:"required"`
}

type UserPINResponse struct {
	UserID        uint `json:"user_id"`
	FailedAttempt int  `json:"failed_attempt"`
	IsLocked      bool `json:"is_locked"`
}
