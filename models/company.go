package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Company struct {
	Generic

	// Company name.
	Name string `gorm:"not null" json:"name"`
	// Ticker symbol of the company. It is unique.
	Ticker string `gorm:"unique_index" json:"ticker"`
	// SEC company identifier. We can search by CIK, it is unique for US
	// companies. Some (non-US companies) might not have it.
	CIK string `gorm:"unique_index" json:"-"`
	// Last time we fetched the company's documents.
	LastFetchedAt time.Time `json:"-"`

	Currency    string  `json:"currency"`
	Price       float64 `json:"price"`
	Change      float64 `json:"change"`
	TotalVolume float64 `json:"total_volume"`
}

func GetCompanyByID(db *gorm.DB, companyID uint) (*Company, error) {
	var company Company
	err := db.Where("id = ?", companyID).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &company, nil
}

// Get company by ticker.
func GetCompanyByTicker(db *gorm.DB, ticker string) (*Company, error) {
	ticker = strings.ToUpper(ticker)

	var company Company
	err := db.Where("ticker = ?", strings.ToUpper(ticker)).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &company, nil
}

// Create company.
func CreateCompany(db *gorm.DB, name, ticker, cik string, lastFetchedAt time.Time) (*Company, error) {
	var company = Company{
		Name:          name,
		Ticker:        strings.ToUpper(ticker),
		CIK:           cik,
		LastFetchedAt: lastFetchedAt,
	}

	if err := db.Create(&company).Error; err != nil {
		return nil, err
	}

	return &company, nil
}

func FindCompanies(db *gorm.DB, query string, offset, limit int) ([]Company, error) {
	var companies []Company
	var err error
	if len(query) > 0 {
		q := "%" + query + "%"
		err = db.Where("last_fetched_at > ? AND (name ILIKE ? OR ticker ILIKE ?)", time.Time{}, q, q).Offset(offset).Limit(limit).Order("total_volume DESC").Find(&companies).Error
	} else {
		err = db.Where("last_fetched_at > ?", time.Time{}).Offset(offset).Limit(limit).Order("total_volume DESC").Find(&companies).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return companies, nil
}
