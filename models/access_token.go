package models

type AccessToken struct {
	Generic

	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `json:"-"`

	Token string `gorm:"unique_index" json:"token"`
}
