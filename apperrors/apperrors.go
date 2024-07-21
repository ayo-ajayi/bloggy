package apperrors

import (
	"fmt"
)

type AppError struct {
	UserMessage string `json:"message"`
	StatusCode  int    `json:"status_code"`
	InternalErr error  `json:"-"`
}

func NewError(userMessage string, statusCode int, internalErr error) *AppError {
	return &AppError{
		UserMessage: userMessage,
		StatusCode:  statusCode,
		InternalErr: internalErr,
	}
}

func (e *AppError) Error() string {
	if e.InternalErr != nil {
		return fmt.Sprintf("%s: %v", e.UserMessage, e.InternalErr)
	}
	return e.UserMessage
}

func (e *AppError) Unwrap() error {
	return e.InternalErr
}

func (e *AppError) Sanitize() *AppError {
	return &AppError{
		UserMessage: e.UserMessage,
		StatusCode:  e.StatusCode,
	}
}
