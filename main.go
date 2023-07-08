package main

import (
	"context"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"cofin/controllers"
	"cofin/core"
	"cofin/fetcher"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if runFetcher := os.Getenv("RUN_FETCHER"); runFetcher != "" {
		runFetcherBool, err := strconv.ParseBool(runFetcher)
		if err != nil {
			panic(err)
		}

		if runFetcherBool {
			fetcher, err := fetcher.NewFetcher(db)
			if err != nil {
				panic(err)
			}

			go fetcher.Loop(ctx)
		}
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
