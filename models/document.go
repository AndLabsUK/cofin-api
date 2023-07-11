package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Documents are raw document inputs. We don't usually use them for search or
// retrieval. They are stored primarily for debugging and supportability
// reasons. In search, we use different representations of these documents, such
// as their chunks.
type Document struct {
	Generic

	CompanyID uint `gorm:"index;not null"`
	Company   Company
	// UUID is used to ensure consistency with vector storage. Since we don't
	// have atomic updates to vector storage, we always create the document only
	// once we succeed at uploading all of its chunks to vector storage. If some
	// chunks get uploaded and others fail to upload, they will be "dead" in the
	// vector storage since they won't have a matching document in the DB.
	//
	// We query chunks based on document UUID, so this is fine.
	UUID       uuid.UUID  `gorm:"index;not null"`
	FiledAt    time.Time  `gorm:"index;not null"`
	Kind       SourceKind `gorm:"index;not null"`
	OriginURL  string
	RawContent string
}

func CreateDocument(db *gorm.DB, company *Company, filedAt time.Time, kind SourceKind, originURL, rawContent string) (*Document, error) {
	document := Document{
		CompanyID:  company.ID,
		UUID:       uuid.New(),
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

func GetMostRecentCompanyDocumentOfKind(db *gorm.DB, companyID uint, kind SourceKind) (*Document, error) {
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

func GetRecentCompanyDocuments(db *gorm.DB, companyID uint, limit int) ([]Document, error) {
	var documents []Document
	err := db.Where("company_id = ?", companyID).Order("filed_at DESC").Limit(limit).Find(&documents).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return documents, nil
}

func GetDocumentByUUID(db *gorm.DB, documentUUID uuid.UUID) (*Document, error) {
	var document Document
	err := db.Where("uuid = ?", documentUUID).First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &document, nil
}
