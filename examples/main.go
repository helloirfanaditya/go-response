package main

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/helloirfanaditya/go-response"
)

type createUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

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
