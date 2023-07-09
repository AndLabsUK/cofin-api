package controllers

import (
	"cofin/core"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func (h HealthController) Status(c *gin.Context) {
	db, err := core.GetDB()

	if err != nil {
		ResultError(c, nil)
		return
	}

	err = db.Raw(`SELECT 1`).Row().Err()

	if err != nil {
		ResultError(c, nil)
		return
	}

	ResultSuccess(c)
}
