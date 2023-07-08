package controllers

import (
	"cofin/api"
	"cofin/core"
	"cofin/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

type CompaniesController struct{}

func (cc CompaniesController) GetCompanies(c *gin.Context) {
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		api.ResultError(c, []string{"invalidRequest"})
		return
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		api.ResultError(c, []string{"invalidRequest"})
		return
	}

	query := c.Query("query")

	db, _ := core.GetDB()

	var companies []models.Company
	var result *gorm.DB

	if len(query) > 0 {
		result = db.Where("name ILIKE ? OR ticker ILIKE ?", "%"+query+"%", query+"%").Offset(offset).Limit(limit).Order("total_volume desc").Find(&companies)
	} else {
		result = db.Offset(offset).Limit(limit).Order("total_volume desc").Find(&companies)
	}

	if result.Error != nil {
		api.ResultError(c, nil)
	}

	api.ResultData(c, companies)
}
