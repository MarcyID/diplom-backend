package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ErrorResponse формат ошибки API
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse формат успешного ответа
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Error возвращает стандартизированную ошибку
func Error(c *gin.Context, statusCode int, code string, message string) {
	c.Header("Content-Type", "application/json")
	c.JSON(statusCode, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// ErrorWithDetails возвращает ошибку с дополнительными деталями
func ErrorWithDetails(c *gin.Context, statusCode int, code string, message string, details string) {
	c.Header("Content-Type", "application/json")
	c.JSON(statusCode, ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}

// Success возвращает успешный ответ с данными
func Success(c *gin.Context, data interface{}) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, SuccessResponse{
		Data: data,
	})
}

// SuccessWithMessage возвращает успешный ответ с сообщением
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// Common error codes
const (
	ErrBadRequest          = "BAD_REQUEST"
	ErrUnauthorized        = "UNAUTHORIZED"
	ErrForbidden           = "FORBIDDEN"
	ErrNotFound            = "NOT_FOUND"
	ErrConflict            = "CONFLICT"
	ErrInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrValidation          = "VALIDATION_ERROR"
)
