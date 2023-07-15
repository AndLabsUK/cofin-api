package controllers

import (
	"cofin/models"
	"net/http"
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

	company, err := models.GetCompanyByTicker(cc.DB, ticker)
	if err != nil {
		cc.Logger.Errorf("Error querying company: %w", err)
		RespondInternalErr(c)
		return
	} else if company == nil {
		RespondCustomStatusErr(c, http.StatusNotFound, []error{ErrUnknownCompany})
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
	companies, err := models.FindCompanies(cc.DB, query, limit, offset)
	if err != nil {
		cc.Logger.Errorf("Error querying companies: %w", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, companies)
}
