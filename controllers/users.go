package controllers

import (
	"cofin/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UsersController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (uc UsersController) GetCurrentUser(c *gin.Context) {
	user, err := models.GetUserByID(uc.DB, CurrentUserID(c))
	if err != nil {
		uc.Logger.Errorf("Error querying users: %w", err)
		RespondInternalErr(c)
		return
	}

	if user == nil {
		RespondCustomStatusErr(c, http.StatusNotFound, []error{ErrUnknownUser})
		return
	}

	RespondOK(c, user)
}
