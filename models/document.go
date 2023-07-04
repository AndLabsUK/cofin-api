package models

import (
	"cofin/internal"
	"time"

	"gorm.io/gorm"
)

// Documents are raw document inputs. We don't usually use them for search or
// retrieval. They are stored primarily for debugging and supportability
// reasons. In search, we use different representations of these documents, such
// as their chunks.
type Document struct {
	gorm.Model
	CompanyID  uint
	Company    Company
	FiledAt    time.Time           `gorm:"index;not null"`
	Kind       internal.SourceKind `gorm:"index;not null"`
	OriginURL  string
	RawContent string
}
