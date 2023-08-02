package models

import "time"

type Company struct {
	Generic

	// Company name.
	Name string `gorm:"not null" json:"name"`
	// Ticker symbol of the company. It is unique.
	Ticker string `gorm:"unique_index" json:"ticker"`
	// SEC company identifier. We can search by CIK, it is unique for US
	// companies. Some (non-US companies) might not have it.
	CIK string `gorm:"unique_index" json:"-"`
	// Last time we fetched the company's documents.
	LastFetchedAt time.Time `json:"-"`

	Currency    string  `json:"currency"`
	Price       float64 `json:"price"`
	Change      float64 `json:"change"`
	TotalVolume float64 `json:"total_volume"`
}
