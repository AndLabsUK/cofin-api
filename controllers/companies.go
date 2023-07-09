package controllers

import (
	"cofin/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CompaniesController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (cc CompaniesController) GetCompanies(c *gin.Context) {
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		WriteBadRequestError(c, []error{err})
		return
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		WriteBadRequestError(c, []error{err})
		return
	}

	query := c.Query("query")

	var companies []models.Company
	var result *gorm.DB
	if len(query) > 0 {
		result = cc.DB.Where("name ILIKE ? OR ticker ILIKE ?", "%"+query+"%", query+"%").Offset(offset).Limit(limit).Order("total_volume desc").Find(&companies)
	} else {
		result = cc.DB.Offset(offset).Limit(limit).Order("total_volume desc").Find(&companies)
	}

	if result.Error != nil {
		cc.Logger.Errorf("Error querying companies: %w", result.Error)
		WriteInternalError(c)
		return
	}

	WriteSuccess(c, companies)
}
