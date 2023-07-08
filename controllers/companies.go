package controllers

import (
	"cofin/cmd/api"
	"cofin/core"
	"cofin/models"
	"github.com/gin-gonic/gin"
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

	db, _ := core.GetDB()

	var companies []models.Company
	result := db.Offset(offset).Limit(limit).Order("total_volume desc").Find(&companies)
	if result.Error != nil {
		api.ResultError(c, nil)
	}

	api.ResultData(c, companies)
}
