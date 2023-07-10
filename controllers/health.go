package controllers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HealthController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (h HealthController) Status(c *gin.Context) {
	err := h.DB.Raw(`SELECT 1`).Row().Err()
	if err != nil {
		h.Logger.Errorf("Error checking database health: %v", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, nil)
}
