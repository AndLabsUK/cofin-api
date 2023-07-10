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

type apiResponse struct {
	Errors []error `json:"errors,omitempty"`
	Data   any     `json:"data,omitempty"`
}

func RespondOK(c *gin.Context, obj any) {
	c.JSON(http.StatusOK, apiResponse{Data: obj})
}

func RespondBadRequestErr(c *gin.Context, errors []error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, apiResponse{Errors: errors})
}

func RespondInternalErr(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, apiResponse{Errors: []error{ErrInternalError}})
}
