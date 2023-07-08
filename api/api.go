package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiResponse struct {
	Errors []string `json:"errors,omitempty"`
	Data   any      `json:"data,omitempty"`
}

func ResultData(c *gin.Context, obj any) {
	c.JSON(http.StatusOK, ApiResponse{Data: obj})
}

func ResultSuccess(c *gin.Context) {
	c.JSON(http.StatusOK, ApiResponse{})
}

func ResultError(c *gin.Context, errors []string) {
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, ApiResponse{Errors: errors})
	} else {
		c.JSON(http.StatusInternalServerError, ApiResponse{Errors: []string{"unknownError"}})
	}
}

func ResultErrorWithData(c *gin.Context, errors []string, obj any) {
	c.JSON(http.StatusBadRequest, ApiResponse{Errors: errors, Data: obj})
}
