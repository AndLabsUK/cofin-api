package models

import (
	"gorm.io/gorm"
)

// DocumentChunks are text chunks used for semantic search and question
// answering. They are derived from Documents.
type DocumentChunk struct {
	gorm.Model
	DocumentID uint
	Document   Document
	RawContent string
}
