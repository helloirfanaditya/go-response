// Package response provides a consistent JSON response format for Gin
// applications. It standardizes success, error, validation, and pagination
// responses so every service produces predictable output.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response codes shared by every response body.
const (
	// CodeSuccess indicates the request completed successfully.
	CodeSuccess = "SUCCESS"
	// CodeCreated indicates a resource was created.
	CodeCreated = "CREATED"
	// CodeBadRequest indicates the request was malformed.
	CodeBadRequest = "BAD_REQUEST"
	// CodeUnauthorized indicates authentication is missing or failed.
	CodeUnauthorized = "UNAUTHORIZED"
	// CodeForbidden indicates the request is not allowed.
	CodeForbidden = "FORBIDDEN"
	// CodeNotFound indicates the requested resource does not exist.
	CodeNotFound = "NOT_FOUND"
	// CodeConflict indicates the request conflicts with the current state.
	CodeConflict = "CONFLICT"
	// CodeValidationError indicates the request failed validation.
	CodeValidationError = "VALIDATION_ERROR"
	// CodeInternalServerError indicates an unexpected server failure.
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
)

const (
	messageSuccess          = "Success"
	messageCreated          = "Created"
	messageValidationFailed = "Validation failed"
)

// successBody is the JSON payload for 2xx responses.
type successBody struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Success writes a 200 OK response with code SUCCESS.
//
// Data is always included in the payload. Pass nil only when the payload is
// intentionally null; pass an empty slice ([]T{}) when the payload should
// serialize as [].
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successBody{
		Success: true,
		Code:    CodeSuccess,
		Message: messageSuccess,
		Data:    data,
	})
}

// Created writes a 201 Created response with code CREATED.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, successBody{
		Success: true,
		Code:    CodeCreated,
		Message: messageCreated,
		Data:    data,
	})
}

// NoContent writes a 204 No Content response without a body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}
