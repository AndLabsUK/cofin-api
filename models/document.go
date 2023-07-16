package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"gorm.io/gorm"
)

// SourceKind describes a category of source data for API response -- can be an
// SEC filing, an investor call transcript, etc. SourceType implements the
// Scanner interface and the Stringer interface.
type SourceKind string

const (
	Q10 SourceKind = "10-Q"
	K10 SourceKind = "10-K"
)

type Quarter uint8

const (
	Q1 Quarter = iota + 1
	Q2
	Q3
	Q4
)

func (st *SourceKind) Scan(value interface{}) error {
	*st = SourceKind(value.(string))
	return nil
}

func (st *SourceKind) Value() (driver.Value, error) {
	return string(*st), nil
}

func (st *SourceKind) String() string {
	return string(*st)
}

// Document section (chapter).
type Section string

const (
	// 10-K sections
	K10Business                                             Section = "1"
	K10RiskFactors                                          Section = "1A"
	K10UnresolvedStaffComments                              Section = "1B"
	K10Properties                                           Section = "2"
	K10LegalProceedings                                     Section = "3"
	K10MineSafetyDisclosures                                Section = "4"
	K10MarketForRegistrantsCommonEquity                     Section = "5"
	K10SelectedFinancialData                                Section = "6"
	K10ManagementsDiscussion                                Section = "7"
	K10QuantitativeAndQualitativeDisclosuresAboutMarketRisk Section = "7A"
	K10FinancialStatementsAndSupplementaryData              Section = "8"
	K10ChangesInAndDisagreementsWithAccountants             Section = "9"
	K10ControlsAndProcedures                                Section = "9A"
	K10OtherInformation                                     Section = "9B"
	K10DirectorsExecutiveOfficersAndCorporateGovernance     Section = "10"
	K10ExecutiveCompensation                                Section = "11"
	K10SecurityOwnership                                    Section = "12"
	K10CertainRelationships                                 Section = "13"
	K10PrincipalAccountantFeesAndServices                   Section = "14"

	// 10-Q sections
	Q10FinancialStatements  Section = "part1item1"
	Q10ManagementDiscussion Section = "part1item2"
	Q10MarketRisk           Section = "part1item3"
	Q10Controls             Section = "part1item4"
	Q10LegalProceedings     Section = "part2item1"
	Q10RiskFactors          Section = "part2item1a"
	Q10Unregistered         Section = "part2item2"
	Q10Defaults             Section = "part2item3"
	Q10MineSafety           Section = "part2item4"
	Q10OtherInformation     Section = "part2item5"
	Q10Exhibits             Section = "part2item6"
)

var (
	K10Sections = []Section{
		K10Business,
		K10RiskFactors,
		K10UnresolvedStaffComments,
		K10Properties,
		K10LegalProceedings,
		K10MineSafetyDisclosures,
		K10MarketForRegistrantsCommonEquity,
		K10SelectedFinancialData,
		K10ManagementsDiscussion,
		K10QuantitativeAndQualitativeDisclosuresAboutMarketRisk,
		K10FinancialStatementsAndSupplementaryData,
		K10ChangesInAndDisagreementsWithAccountants,
		K10ControlsAndProcedures,
		K10OtherInformation,
		K10DirectorsExecutiveOfficersAndCorporateGovernance,
		K10ExecutiveCompensation,
		K10SecurityOwnership,
		K10CertainRelationships,
		K10PrincipalAccountantFeesAndServices,
	}

	Q10Sections = []Section{
		Q10FinancialStatements,
		Q10ManagementDiscussion,
		Q10MarketRisk,
		Q10Controls,
		Q10LegalProceedings,
		Q10RiskFactors,
		Q10Unregistered,
		Q10Defaults,
		Q10MineSafety,
		Q10OtherInformation,
		Q10Exhibits,
	}
)

// Documents are raw document inputs.
type Document struct {
	Generic

	CompanyID  uint       `gorm:"index;not null"`
	Company    Company    `json:"-"`
	FiledAt    time.Time  `gorm:"index;not null"`
	Kind       SourceKind `gorm:"index;not null"`
	OriginURL  string
	RawContent string `json:"-"`
}

func CreateDocument(db *gorm.DB, company *Company, filedAt time.Time, kind SourceKind, originURL, rawContent string) (*Document, error) {
	document := Document{
		CompanyID:  company.ID,
		FiledAt:    filedAt,
		Kind:       kind,
		OriginURL:  originURL,
		RawContent: rawContent,
	}

	if err := db.Create(&document).Error; err != nil {
		return nil, err
	}

	return &document, nil
}

func GetCompanyDocumentsOfKindInverseChronological(db *gorm.DB, companyID uint, kind SourceKind) (*Document, error) {
	var document Document
	err := db.Where("company_id = ? AND kind = ?", companyID, kind).Order("filed_at DESC").First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &document, nil
}

func GetCompanyDocumentsInverseChronological(db *gorm.DB, companyID uint, offset, limit int) ([]Document, error) {
	var documents []Document
	err := db.Preload("Company").Where("company_id = ?", companyID).Order("filed_at DESC").Offset(offset).Limit(limit).Find(&documents).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return documents, nil
}

func GetDocumentByID(db *gorm.DB, documentID uint) (*Document, error) {
	var document Document
	err := db.Where("id = ?", documentID).First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &document, nil
}
