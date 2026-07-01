package apperror

type Code string

const (
	// Common
	CodeInternal     Code = "INTERNAL_SERVER_ERROR"
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeBadRequest   Code = "BAD_REQUEST"

	// Business
	CodeInsufficientBalance Code = "INSUFFICIENT_BALANCE"
	CodeOTPExpired          Code = "OTP_EXPIRED"
	CodeInvalidOTP          Code = "INVALID_OTP"
	CodeUserInactive        Code = "USER_INACTIVE"
)
