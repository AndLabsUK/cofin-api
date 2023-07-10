package models

type User struct {
	Generic

	Email             string `gorm:"unique" json:"-"`
	FullName          string `gorm:"not null" json:"full_name"`
	FirebaseSubjectId string `gorm:"unique" json:"-"`
	StripeCustomerId  string `gorm:"unique" json:"-"`
	IsSubscribed      bool   `gorm:"not null; default:false" json:"is_subscribed"`
}
