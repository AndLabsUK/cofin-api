package middleware

import (
	"cofin/cmd/api"
	"cofin/core"
	"cofin/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequireAuth(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if len(token) > 0 {

		db, err := core.GetDB()
		if err == nil {
			var accessToken models.AccessToken
			tx := db.First(&accessToken, "token = ?", token)
			if tx.Error == nil {
				c.Set("userId", accessToken.UserID)
				c.Next()
				return
			}
		}
	}

	c.JSON(http.StatusForbidden, api.ApiResponse{Errors: []string{"accessDenied"}})
	c.Abort()
	return
}

func CurrentUserId(c *gin.Context) uint {
	return c.GetUint("userId")
}

func CurrentUser(c *gin.Context) *models.User {
	userId := CurrentUserId(c)
	if userId == 0 {
		return nil
	}

	db, err := core.GetDB()
	if err != nil {
		return nil
	}

	var user models.User
	tx := db.First(&user, userId)

	if tx.Error != nil {
		return nil
	}

	return &user
}
