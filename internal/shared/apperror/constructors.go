package apperror

func New(
	code Code,
	message string,
	err error,
) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func BadRequest(msg string) *Error {
	return New(CodeBadRequest, msg, nil)
}

func Validation(msg string) *Error {
	return New(CodeValidation, msg, nil)
}

func Unauthorized(msg string) *Error {
	return New(CodeUnauthorized, msg, nil)
}

func Forbidden(msg string) *Error {
	return New(CodeForbidden, msg, nil)
}

func NotFound(msg string) *Error {
	return New(CodeNotFound, msg, nil)
}

func Conflict(msg string) *Error {
	return New(CodeConflict, msg, nil)
}

func Internal(err error) *Error {
	return New(
		CodeInternal,
		"internal server error",
		err,
	)
}
