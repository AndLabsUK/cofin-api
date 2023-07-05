package models

import "gorm.io/gorm"

type Company struct {
	gorm.Model
	// Company name.
	Name string `gorm:"not null"`
	// Ticker symbol of the company. It is unique.
	Ticker string `gorm:"unique_index"`
	// SEC company identifier. We can search by CIK, it is unique for US
	// companies. Some (non-US companies) might not have it.
	CIK string `gorm:"unique_index"`
}
