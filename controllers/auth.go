package controllers

import (
	"cofin/internal/google_pki"
	"cofin/models"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (a AuthController) SignIn(c *gin.Context) {
	type signInParams struct {
		JWTToken string `json:"jwt_token"`
	}

	var payload signInParams
	if err := c.BindJSON(&payload); err != nil {
		WriteBadRequestError(c, []error{err})
		return
	}

	token, err := jwt.Parse(payload.JWTToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected JWT signing method: %v", token.Header["alg"])
		}

		kid := token.Header["kid"].(string)

		googlePki := google_pki.GooglePKI{}
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
		WriteBadRequestError(c, []error{err})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		WriteBadRequestError(c, []error{ErrInvalidToken})
		return
	}

	email := claims["email"].(string)
	name := claims["name"].(string)
	sub := claims["sub"].(string)

	var user models.User
	// TODO: rewrite as a transaction.
	tx := a.DB.First(&user, "firebase_subject_id = ?", sub)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			user = models.User{
				Email:             email,
				FullName:          name,
				FirebaseSubjectId: sub,
				IsSubscribed:      false,
			}

			result := a.DB.Create(&user)
			if result.Error != nil {
				a.Logger.Errorf("Error creating user: %w", result.Error)
				WriteInternalError(c)
				return
			}

		} else {
			a.Logger.Errorf("Error getting user: %w", tx.Error)
			WriteInternalError(c)
			return
		}
	}

	stripe := integrations.Stripe{}
	go stripe.CreateCustomer(&user, db)

	accessToken := models.AccessToken{
		UserID: user.ID,
		Token:  generateRandomString(128),
	}
	result := a.DB.Create(&accessToken)
	if result.Error != nil {
		a.Logger.Errorf("Error creating access token: %w", result.Error)
		return
	}

	WriteSuccess(c, accessToken)
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
