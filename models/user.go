package models

type User struct {
	Generic

	Email             string `gorm:"unique" json:"-"`
	FullName          string `gorm:"not null" json:"full_name"`
	FirebaseSubjectId string `gorm:"unique" json:"-"`
}
