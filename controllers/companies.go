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

func (cc CompaniesController) GetCompany(c *gin.Context) {
	ticker := c.Params.ByName("ticker")

	var company models.Company
	result := cc.DB.Where("ticker = ?", ticker).First(&company)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			RespondBadRequestErr(c, []error{result.Error})
			return
		}

		cc.Logger.Errorf("Error querying company: %w", result.Error)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, company)
}

func (cc CompaniesController) GetCompanies(c *gin.Context) {
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	query := c.Query("query")

	var companies []models.Company
	var result *gorm.DB
	if len(query) > 0 {
		result = cc.DB.Where("name ILIKE ? OR ticker ILIKE ?", "%"+query+"%", query+"%").Offset(offset).Limit(limit).Order("total_volume DESC").Find(&companies)
	} else {
		result = cc.DB.Offset(offset).Limit(limit).Order("total_volume DESC").Find(&companies)
	}

	if result.Error != nil {
		cc.Logger.Errorf("Error querying companies: %w", result.Error)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, companies)
}
