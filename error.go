package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCoder is implemented by errors that define their own HTTP response.
//
// response.Error maps any error implementing ErrorCoder to its StatusCode and
// ResponseCode, which keeps the error mapping extensible: services can define
// their own error types without changing this library.
type ErrorCoder interface {
	error
	// StatusCode returns the HTTP status code for the error.
	StatusCode() int
	// ResponseCode returns the stable machine-readable response code.
	ResponseCode() string
}

// ResponseError is the standard application error. Create one with NewError or
// one of the typed constructors such as NotFound.
type ResponseError struct {
	status int
	code   string
	msg    string
}

// Error returns the human-readable message.
func (e *ResponseError) Error() string { return e.msg }

// StatusCode returns the HTTP status code.
func (e *ResponseError) StatusCode() int { return e.status }

// ResponseCode returns the stable machine-readable response code.
func (e *ResponseError) ResponseCode() string { return e.code }

// NewError builds a ResponseError with a custom HTTP status and response code.
// Prefer the typed constructors for the standard error kinds.
func NewError(status int, code, message string) *ResponseError {
	return &ResponseError{status: status, code: code, msg: message}
}

// BadRequest returns a 400 Bad Request error with code BAD_REQUEST.
func BadRequest(message string) *ResponseError {
	return NewError(http.StatusBadRequest, CodeBadRequest, message)
}

// Unauthorized returns a 401 Unauthorized error with code UNAUTHORIZED.
func Unauthorized(message string) *ResponseError {
	return NewError(http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden returns a 403 Forbidden error with code FORBIDDEN.
func Forbidden(message string) *ResponseError {
	return NewError(http.StatusForbidden, CodeForbidden, message)
}

// NotFound returns a 404 Not Found error with code NOT_FOUND.
func NotFound(message string) *ResponseError {
	return NewError(http.StatusNotFound, CodeNotFound, message)
}

// Conflict returns a 409 Conflict error with code CONFLICT.
func Conflict(message string) *ResponseError {
	return NewError(http.StatusConflict, CodeConflict, message)
}

// UnprocessableEntity returns a 422 Unprocessable Entity error with code
// VALIDATION_ERROR.
func UnprocessableEntity(message string) *ResponseError {
	return NewError(http.StatusUnprocessableEntity, CodeValidationError, message)
}

// InternalServerError returns a 500 Internal Server Error with code
// INTERNAL_SERVER_ERROR.
func InternalServerError(message string) *ResponseError {
	return NewError(http.StatusInternalServerError, CodeInternalServerError, message)
}

// errorBody is the JSON payload for error responses.
type errorBody struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error converts an application error into a JSON error response.
//
// Errors implementing ErrorCoder (including ResponseError) are mapped to their
// own HTTP status and response code. Any other error falls back to 500
// INTERNAL_SERVER_ERROR.
//
// Error panics when called with a nil error.
func Error(c *gin.Context, err error) {
	if err == nil {
		panic("response: Error called with a nil error")
	}

	var target ErrorCoder
	if errors.As(err, &target) {
		c.JSON(target.StatusCode(), errorBody{
			Success: false,
			Code:    target.ResponseCode(),
			Message: target.Error(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, errorBody{
		Success: false,
		Code:    CodeInternalServerError,
		Message: err.Error(),
	})
}
