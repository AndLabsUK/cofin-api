package controllers

import (
	"cofin/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UsersController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (u UsersController) GetCurrentUser(c *gin.Context) {
	var user models.User
	userId := CurrentUserId(c)
	u.DB.First(&user, userId)

	WriteSuccess(c, user)
}
