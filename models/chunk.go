package models

import (
	"errors"

	"gorm.io/gorm"
)

type Chunk struct {
	Generic

	RawContent string `gorm:"not null" json:"raw_content"`
}

func CreateChunk(db *gorm.DB, rawContent string) (*Chunk, error) {
	chunk := Chunk{
		RawContent: rawContent,
	}

	err := db.Create(&chunk).Error
	if err != nil {
		return nil, err
	}

	return &chunk, nil
}

func GetChunkByID(db *gorm.DB, id uint) (*Chunk, error) {
	var chunk Chunk

	err := db.First(&chunk, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &chunk, nil
}
