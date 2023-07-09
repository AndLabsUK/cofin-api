package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidToken  = errors.New("Invalid token")
	ErrInternalError = errors.New("Internal error")
	ErrUnknownTicker = errors.New("Unknown ticker")
)

type ApiResponse struct {
	Errors []error `json:"errors,omitempty"`
	Data   any     `json:"data,omitempty"`
}

func WriteSuccess(c *gin.Context, obj any) {
	c.JSON(http.StatusOK, ApiResponse{Data: obj})
}

func WriteBadRequestError(c *gin.Context, errors []error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, ApiResponse{Errors: errors})
}

func WriteInternalError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, ApiResponse{Errors: []error{ErrInternalError}})
}
