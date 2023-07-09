package main

import (
	"cofin/controllers"
	"cofin/core"
	"cofin/internal/retrieval"
	"cofin/models"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
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

	server := createServer(db)
	server.Run()
}

func createServer(db *gorm.DB) *gin.Engine {
	// set up http server
	engine := gin.Default()
	err := engine.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://"+os.Getenv("UI_DOMAIN"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, Accept, Origin, Cache-Control, X-Requested-With, X-User-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	generator, err := retrieval.NewGenerator()
	if err != nil {
		panic(err)
	}

	logger, err := core.NewLogger()
	if err != nil {
		panic(err)
	}

	conversationController := controllers.ConversationController{
		DB:        db,
		Generator: generator,
		Logger:    logger,
	}

	router := controllers.Router{
		HealthController:       &controllers.HealthController{},
		AuthController:         &controllers.AuthController{},
		UsersController:        &controllers.UsersController{},
		CompaniesController:    &controllers.CompaniesController{},
		ConversationController: &conversationController,
	}

	router.RegisterRoutes(engine)
	return engine
}
