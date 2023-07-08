package core

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() (*gorm.DB, error) {

	databaseUrl := os.Getenv("DATABASE_URL")
	credentials, err := url.Parse(databaseUrl)

	username := credentials.User.Username()
	password, _ := credentials.User.Password()
	host, port, _ := net.SplitHostPort(credentials.Host)
	dbName := credentials.Path[1:]

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host,
		username,
		password,
		dbName,
		port,
		"prefer",
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db = gormDB

	return db, nil
}

func GetDB() (*gorm.DB, error) {
	if db == nil {
		return InitDB()
	}

	return db, nil
}
