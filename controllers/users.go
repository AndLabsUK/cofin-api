package controllers

import (
	"cofin/core"
	"cofin/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UsersController struct{}

func (u UsersController) GetAll(c *gin.Context) {
	db, _ := core.GetDB()

	var users []models.User
	db.Find(&users)

	c.JSON(http.StatusOK, gin.H{"users": users})
}
