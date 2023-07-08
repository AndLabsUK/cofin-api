package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"

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
	engine := gin.Default()
	err = engine.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://"+os.Getenv("UI_DOMAIN"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, Accept, Origin, Cache-Control, X-Requested-With, X-Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	healthController := controllers.HealthController{}
	authController := controllers.AuthController{}
	usersController := controllers.UsersController{}
	companiesController := controllers.CompaniesController{}

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
		companiesController:    &companiesController,
		conversationController: &conversationController,
	}

	router.RegisterRoutes(engine)

	err = engine.Run()
	if err != nil {
		return
	}
}
