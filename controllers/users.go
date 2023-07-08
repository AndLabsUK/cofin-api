package controllers

import (
	"cofin/api"
	"cofin/api/middleware"
	"cofin/core"
	"cofin/models"
	"github.com/gin-gonic/gin"
)

type UsersController struct{}

func (u UsersController) GetCurrentUser(c *gin.Context) {
	db, _ := core.GetDB()

	var user models.User
	userId := middleware.CurrentUserId(c)
	db.First(&user, userId)

	api.ResultData(c, user)
}
