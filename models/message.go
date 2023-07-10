package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
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

	UserID     uint `gorm:"index;not null"`
	User       User
	CompanyID  uint `gorm:"index;not null"`
	Company    Company
	Author     MessageAuthor
	Text       string
	Annotation JSON
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

type JSON json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface.
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

// Value return json value, implement driver.Valuer interface.
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
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

func GetMessagesForCompany(db *gorm.DB, companyID uint, offset, limit int) ([]Message, error) {
	var messages []Message
	if err := db.Where("company_id = ?", companyID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}
