package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/helloirfanaditya/go-response"
)

type createUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type paymentRequest struct {
	Method string `json:"method" validate:"required,oneof=transfer credit_card"`
	Card   string `json:"card" validate:"required_if=Method credit_card"`
}

type registerRequest struct {
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// Domain-specific error codes — stable, machine-readable, prefixed by domain.
var (
	ErrInvalidAmount       = response.NewError(http.StatusBadRequest, "TRANSACTION_INVALID_AMOUNT", "amount must be greater than zero")
	ErrInsufficientBalance = response.NewError(http.StatusBadRequest, "TRANSACTION_INSUFFICIENT_BALANCE", "insufficient balance")
	ErrRecipientNotFound   = response.NewError(http.StatusNotFound, "TRANSACTION_RECIPIENT_NOT_FOUND", "recipient account not found")
	ErrStoreFailed         = response.NewError(http.StatusInternalServerError, "TRANSACTION_STORE_FAILED", "failed to store transaction")
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		response.Success(c, gin.H{"message": "pong"})
	})

	r.POST("/users", func(c *gin.Context) {
		var req createUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Validation(c, &req, err)
			return
		}
		response.Created(c, gin.H{"id": 1})
	})

	// Conditional validation: card wajib hanya kalau method = credit_card.
	r.POST("/payments", func(c *gin.Context) {
		var req paymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Validation(c, &req, err)
			return
		}
		response.Created(c, gin.H{"payment_id": 1})
	})

	// Cross-field validation: confirm_password harus sama dengan password.
	r.POST("/register", func(c *gin.Context) {
		var req registerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Validation(c, &req, err)
			return
		}
		response.Created(c, gin.H{"user_id": 1})
	})

	// Domain error dengan code spesifik.
	r.POST("/transfers", func(c *gin.Context) {
		response.Error(c, ErrInsufficientBalance)
	})

	// Field-level errors: business rule yang nggak bisa diungkapin validator tag.
	r.POST("/withdrawals", func(c *gin.Context) {
		response.Error(c, response.NewValidationError("withdrawal rejected",
			response.FieldError{
				Field:  "amount",
				Code:   "TRANSACTION_EXCEEDS_LIMIT",
				Params: map[string]any{"max": 5000000},
			},
			response.FieldError{
				Field: "recipient",
				Code:  "TRANSACTION_ACCOUNT_BLOCKED",
			},
		))
	})

	r.GET("/users", func(c *gin.Context) {
		response.Paginate(c, []gin.H{}, response.Meta{Page: 1, PerPage: 10, Total: 100})
	})

	r.GET("/users/:id", func(c *gin.Context) {
		response.Error(c, response.NotFound("user not found"))
	})

	r.DELETE("/users/:id", func(c *gin.Context) {
		response.NoContent(c)
	})

	r.GET("/boom", func(c *gin.Context) {
		response.Error(c, errors.New("unexpected failure"))
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
