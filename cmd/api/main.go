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
	err = db.Debug().AutoMigrate(&models.User{}, &models.Company{}, &models.Document{})
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
	usersController := controllers.UsersController{}
	informationRetriever, err := internal.NewInformationRetriever()
	if err != nil {
		panic(err)
	}

	generator, err := internal.NewGenerator()
	if err != nil {
		panic(err)
	}

	conversationController := controllers.ConversationController{
		InformationRetriever: informationRetriever,
		Generator:            generator,
	}

	r.GET("/health", healthController.Status)
	r.GET("/users", usersController.GetAll)
	r.GET("/conversation", conversationController.Respond)

	err = r.Run()
	if err != nil {
		return
	}
}
