package models

import (
	"gorm.io/gorm"
)

type AccessToken struct {
	Generic

	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `json:"user"`

	Token string `gorm:"unique_index" json:"token"`
}

func CreateAccessToken(db *gorm.DB, userID uint, token string) (*AccessToken, error) {
	accessToken := &AccessToken{
		UserID: userID,
		Token:  token,
	}

	if err := db.Create(accessToken).Error; err != nil {
		return nil, err
	}

	// Return access token with user object preloaded.
	if err := db.Preload("User").First(&accessToken, accessToken.ID).Error; err != nil {
		return nil, err
	}

	if err := setMessageAllowance(db, &accessToken.User); err != nil {
		return nil, err
	}

	return accessToken, nil
}
