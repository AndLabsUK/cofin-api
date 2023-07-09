package main

import (
	"cofin/core"
	"cofin/models"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// connect to the database
	db, err := core.InitDB()
	if err != nil {
		panic(err)
	}

	// auto migrate the database
	err = db.Debug().AutoMigrate(
		&models.User{},
		&models.Company{},
		&models.Document{},
		&models.AccessToken{},
	)
	if err != nil {
		panic(err)
	}

	fetcher, err := NewDocumentFetcher(db)
	if err != nil {
		panic(err)
	}

	fetcher.Run()
}
