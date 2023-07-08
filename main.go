package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"cofin/controllers"
	"cofin/core"
	"cofin/internal"
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
	err = db.Debug().AutoMigrate(
		&models.User{},
		&models.Company{},
		&models.Document{},
		&models.AccessToken{},
	)
	if err != nil {
		panic(err)
	}

	// set up http server
	r := gin.Default()
	err = r.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	healthController := controllers.HealthController{}
	authController := controllers.AuthController{}
	usersController := controllers.UsersController{}

	generator, err := internal.NewGenerator()
	if err != nil {
		panic(err)
	}

	conversationController := controllers.ConversationController{
		DB:        db,
		Generator: generator,
	}

	router := Router{
		healthController:       &healthController,
		authController:         &authController,
		usersController:        &usersController,
		conversationController: &conversationController,
	}

	router.RegisterRoutes(r)

	err = r.Run()
	if err != nil {
		return
	}
}
