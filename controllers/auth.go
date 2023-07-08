package controllers

import (
	"cofin/cmd/main/api"
	"cofin/core"
	"cofin/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type AuthController struct{}

func (a AuthController) SignIn(c *gin.Context) {

	type signInParams struct {
		JWTToken string `json:"jwt_token"`
	}

	var payload signInParams
	if bindingError := c.BindJSON(&payload); bindingError != nil {
		api.ResultError(c, []string{"invalidRequest"})
		return
	}

	////////
	token, err := jwt.Parse(payload.JWTToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected JWT signing method: %v", token.Header["alg"])
		}

		return nil, nil //TODO: Verify signature for JWT token using google public key (IMPORTANT)
	})
	if err == nil { //REPLACE to err != nil
		api.ResultError(c, []string{"invalidRequest"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok { //|| !token.Valid {
		api.ResultError(c, []string{"invalidRequest"})
		return
	}

	email := claims["email"].(string)
	name := claims["name"].(string)
	sub := claims["sub"].(string)

	db, err := core.GetDB()
	if err != nil {
		api.ResultError(c, nil)
		return
	}

	var user models.User
	tx := db.First(&user, "firebase_subject_id = ?", sub)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			user = models.User{
				Email:             email,
				FullName:          name,
				FirebaseSubjectId: sub,
			}

			result := db.Create(&user)
			if result.Error != nil {
				api.ResultError(c, nil)
				return
			}
		} else {
			api.ResultError(c, nil)
			return
		}
	}

	accessToken := models.AccessToken{
		UserID: user.ID,
		Token:  generateRandomString(128),
	}
	result := db.Create(&accessToken)
	if result.Error != nil {
		api.ResultError(c, nil)
		return
	}

	api.ResultData(c, accessToken)
}

func generateRandomString(l int) string {
	var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, l)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}
