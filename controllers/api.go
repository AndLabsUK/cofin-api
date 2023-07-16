package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidToken   = errors.New("Invalid token")
	ErrInternalError  = errors.New("Internal error")
	ErrUnknownCompany = errors.New("Unknown company")
	ErrAccessDenied   = errors.New("Access denied")
	ErrUnpaidUser     = errors.New("Unpaid user")
	ErrUnknownUser    = errors.New("Unknown user")
)

type apiResponse struct {
	Errors []string `json:"errors,omitempty"`
	Data   any      `json:"data,omitempty"`
}

func RespondOK(c *gin.Context, obj any) {
	c.JSON(http.StatusOK, apiResponse{Data: obj})
}

func RespondBadRequestErr(c *gin.Context, errors []error) {
	RespondCustomStatusErr(c, http.StatusBadRequest, errors)
}

func RespondInternalErr(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, apiResponse{Errors: []string{ErrInternalError.Error()}})
}

func RespondCustomStatusErr(c *gin.Context, status int, errors []error) {
	errStrings := make([]string, len(errors))
	for i, err := range errors {
		errStrings[i] = err.Error()
	}
	c.AbortWithStatusJSON(status, apiResponse{Errors: errStrings})
}
