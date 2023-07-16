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
	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	company, err := models.GetCompanyByID(cc.DB, uint(companyID))
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
	if ticker := c.Query("ticker"); ticker != "" {
		company, err := models.GetCompanyByTicker(cc.DB, ticker)
		if err != nil {
			cc.Logger.Errorf("Error querying company: %w", err)
			RespondInternalErr(c)
			return
		}

		if company == nil {
			RespondCustomStatusErr(c, http.StatusNotFound, []error{ErrUnknownCompany})
			return
		}

		RespondOK(c, company)
		return
	}

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
	companies, err := models.FindCompanies(cc.DB, query, offset, limit)
	if err != nil {
		cc.Logger.Errorf("Error querying companies: %w", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, companies)
}

func (cc CompaniesController) GetCompanyDocuments(c *gin.Context) {
	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	company, err := models.GetCompanyByID(cc.DB, uint(companyID))
	if err != nil {
		cc.Logger.Errorf("Error querying company: %w", err)
		RespondInternalErr(c)
		return
	} else if company == nil {
		RespondCustomStatusErr(c, http.StatusNotFound, []error{ErrUnknownCompany})
		return
	}

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

	documents, err := models.GetCompanyDocumentsInverseChronological(cc.DB, company.ID, offset, limit)
	if err != nil {
		cc.Logger.Errorf("Error querying company documents: %w", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, documents)
}
