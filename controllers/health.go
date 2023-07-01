package controllers

import (
	"cofin/core"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HealthController struct{}

func (h HealthController) Status(c *gin.Context) {
	db, err := core.GetDB()

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}

	err = db.Raw(`SELECT 1`).Row().Err()

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
