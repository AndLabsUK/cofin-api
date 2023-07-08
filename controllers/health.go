package controllers

import (
	"cofin/cmd/api"
	"cofin/core"
	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func (h HealthController) Status(c *gin.Context) {
	db, err := core.GetDB()

	if err != nil {
		api.ResultError(c, nil)
		return
	}

	err = db.Raw(`SELECT 1`).Row().Err()

	if err != nil {
		api.ResultError(c, nil)
		return
	}

	api.ResultSuccess(c)
}
