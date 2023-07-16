package controllers

import (
	"cofin/core"
	"cofin/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAuth(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if len(token) > 0 {

		db, err := core.GetDB()
		if err == nil {
			var accessToken models.AccessToken
			tx := db.First(&accessToken, "token = ?", token)
			if tx.Error == nil {
				c.Set("userID", accessToken.UserID)
				c.Next()
				return
			}
		}
	}

	RespondCustomStatusErr(c, http.StatusForbidden, []error{ErrAccessDenied})
}

func CurrentUserID(c *gin.Context) uint {
	return c.GetUint("userID")
}

func CurrentUser(c *gin.Context) *models.User {
	userID := CurrentUserID(c)
	if userID == 0 {
		return nil
	}

	db, err := core.GetDB()
	if err != nil {
		return nil
	}

	user, _ := models.GetUserByID(db, userID)
	return user
}
