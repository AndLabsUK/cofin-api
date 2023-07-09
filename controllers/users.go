package controllers

import (
	"cofin/core"
	"cofin/models"

	"github.com/gin-gonic/gin"
)

type UsersController struct{}

func (u UsersController) GetCurrentUser(c *gin.Context) {
	db, _ := core.GetDB()

	var user models.User
	userId := CurrentUserId(c)
	db.First(&user, userId)

	ResultData(c, user)
}
