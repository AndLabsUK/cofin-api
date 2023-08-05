package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type MessageAuthor string

const (
	UserAuthor MessageAuthor = "user"
	AIAuthor   MessageAuthor = "ai"
)

type Message struct {
	Generic

	UserID     uint          `gorm:"index;not null" json:"user_id"`
	User       User          `json:"-"`
	CompanyID  uint          `gorm:"index;not null" json:"company_id"`
	Company    Company       `json:"-"`
	Author     MessageAuthor `json:"author"`
	Text       string        `json:"text"`
	Annotation JSON          `gorm:"type:jsonb" json:"annotation"`
}

type Source struct {
	ID        uint       `json:"id" binding:"required"`
	Kind      SourceKind `json:"kind" binding:"required"`
	FiledAt   time.Time  `json:"filed_at" binding:"required"`
	OriginURL string     `json:"origin_url" binding:"required"`
}

// Annotation is a serialisable struct that adds metadata to the message row.
type Annotation struct {
	// DocumentIDs describe documents used as the source for the answer.
	Sources []Source `json:"sources"`
}

func CreateUserMessage(db *gorm.DB, userID, companyID uint, text string) (*Message, error) {
	var message = Message{
		UserID:    userID,
		CompanyID: companyID,
		Author:    UserAuthor,
		Text:      text,
	}

	if err := db.Create(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func CreateAIMessage(db *gorm.DB, userID, companyID uint, text string, sources []Source) (*Message, error) {
	annotation := Annotation{
		Sources: sources,
	}

	marshalledAnnotation, err := json.Marshal(annotation)
	if err != nil {
		return nil, err
	}

	var message = Message{
		UserID:     userID,
		CompanyID:  companyID,
		Author:     AIAuthor,
		Text:       text,
		Annotation: marshalledAnnotation,
	}

	if err := db.Create(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func GetMessagesForCompanyInverseChronological(db *gorm.DB, userID, companyID uint, offset, limit int) ([]Message, error) {
	var messages []Message
	if err := db.Where("user_id = ? AND company_id = ?", userID, companyID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

func GetMessagesForCompanyChronological(db *gorm.DB, userID, companyID uint, offset, limit int) ([]Message, error) {
	var messages []Message
	if err := db.Where("user_id = ? AND company_id = ?", userID, companyID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&messages).Error; err != nil {
		return nil, err
	}

	var chronologicalMessages []Message = make([]Message, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		chronologicalMessages = append(chronologicalMessages, messages[i])
	}

	return chronologicalMessages, nil
}

func CountUserGenerations(db *gorm.DB, userID uint) (int64, error) {
	var count int64
	if err := db.Model(&Message{}).Where("user_id = ? AND author = ?", userID, AIAuthor).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
