package service

import "fmt"

// APIError - ошибка с кодом статуса HTTP
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError создаёт новую ошибку API
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewKinopoiskError создаёт ошибку от Kinopoisk API
func NewKinopoiskError(statusCode int, body string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf("Kinopoisk API error: status %d, body: %s", statusCode, body),
	}
}
