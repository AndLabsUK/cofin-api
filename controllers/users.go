package controllers

import (
	"cofin/core"
	"cofin/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UsersController struct{}

func (u UsersController) GetAll(c *gin.Context) {
	db, _ := core.GetDB()

	var users []models.User
	db.Find(&users)

	c.JSON(http.StatusOK, gin.H{"users": users})
}
