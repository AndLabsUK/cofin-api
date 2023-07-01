package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"cofin/controllers"
	"cofin/core"
	"cofin/models"
)

func main() {
	godotenv.Load()

	// connect to the database

	db, err := core.InitDB()
	if err != nil {
		panic(err)
	}

	// auto migrate the database

	err = db.Debug().AutoMigrate(&models.User{})
	if err != nil {
		panic(err)
	}

	// set up http server

	r := gin.Default()

	err = r.SetTrustedProxies(nil)
	if err != nil {
		return
	}

	healthController := controllers.HealthController{}
	usersController := controllers.UsersController{}

	r.GET("/health", healthController.Status)
	r.GET("/users", usersController.GetAll)

	err = r.Run()
	if err != nil {
		return
	}
}
