package apperror

import (
	"errors"
	"net/http"
)

type Response struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

func ToHTTP(err error) (int, Response) {

	var appErr *Error

	if errors.As(err, &appErr) {

		return status(appErr.Code), Response{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
	}

	return http.StatusInternalServerError, Response{
		Code:    CodeInternal,
		Message: "internal server error",
	}
}

func status(code Code) int {

	switch code {

	case CodeBadRequest:
		return http.StatusBadRequest

	case CodeValidation:
		return http.StatusBadRequest

	case CodeUnauthorized:
		return http.StatusUnauthorized

	case CodeForbidden:
		return http.StatusForbidden

	case CodeConflict:
		return http.StatusConflict

	case CodeNotFound:
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}
