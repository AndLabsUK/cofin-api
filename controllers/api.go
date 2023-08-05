package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrInternalError  = errors.New("internal error")
	ErrUnknownCompany = errors.New("unknown company")
	ErrAccessDenied   = errors.New("access denied")
	ErrUnpaidUser     = errors.New("unpaid user")
	ErrUnknownUser    = errors.New("unknown user")
	ErrBadInput       = errors.New("bad input")
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
