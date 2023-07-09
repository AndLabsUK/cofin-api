package controllers

import (
	"cofin/core"
	"cofin/integrations"
	"cofin/models"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthController struct{}

func (a AuthController) SignIn(c *gin.Context) {
	type signInParams struct {
		JWTToken string `json:"jwt_token"`
	}

	var payload signInParams
	if bindingError := c.BindJSON(&payload); bindingError != nil {
		ResultError(c, []string{"invalidRequest"})
		return
	}

	token, err := jwt.Parse(payload.JWTToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected JWT signing method: %v", token.Header["alg"])
		}

		kid := token.Header["kid"].(string)

		googlePki := integrations.GooglePKI{}
		key, err := googlePki.GetPublicKeyForKid(kid)
		if err != nil {
			return nil, err
		}

		n, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return nil, err
		}

		e, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return nil, err
		}

		modulus := new(big.Int).SetBytes(n)
		exponent := big.NewInt(0).SetBytes(e).Uint64()

		rsaPubKey := &rsa.PublicKey{
			N: modulus,
			E: int(exponent),
		}

		return rsaPubKey, nil
	})
	if err != nil {
		ResultError(c, []string{"invalidRequest"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		ResultError(c, []string{"invalidRequest"})
		return
	}

	email := claims["email"].(string)
	name := claims["name"].(string)
	sub := claims["sub"].(string)

	db, err := core.GetDB()
	if err != nil {
		ResultError(c, nil)
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
				ResultError(c, nil)
				return
			}
		} else {
			ResultError(c, nil)
			return
		}
	}

	accessToken := models.AccessToken{
		UserID: user.ID,
		Token:  generateRandomString(128),
	}
	result := db.Create(&accessToken)
	if result.Error != nil {
		ResultError(c, nil)
		return
	}

	ResultData(c, accessToken)
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
