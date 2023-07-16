package controllers

import (
	"cofin/internal/google_pki"
	"cofin/internal/stripe_api"
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
	DB        *gorm.DB
	Logger    *zap.SugaredLogger
	StripeAPI stripe_api.StripeAPI
}

func (ac AuthController) SignIn(c *gin.Context) {
	type signInParams struct {
		JWTToken string `json:"jwt_token"`
	}

	var payload signInParams
	if err := c.BindJSON(&payload); err != nil {
		RespondBadRequestErr(c, []error{err})
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
		RespondBadRequestErr(c, []error{err})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		RespondBadRequestErr(c, []error{ErrInvalidToken})
		return
	}
	email := claims["email"].(string)
	fullName := claims["name"].(string)
	firebaseSubjectID := claims["sub"].(string)

	var user *models.User
	var accessToken *models.AccessToken
	if err := ac.DB.Transaction(func(tx *gorm.DB) (err error) {
		stripeCustomerID, err := ac.StripeAPI.CreateCustomer(email, fullName)
		if err != nil {
			return err
		}

		user, err = models.GetUserByFirebaseSubjectID(tx, firebaseSubjectID)
		if err != nil {
			return err
		} else if user != nil {
			return nil
		}

		ac.Logger.Info("Creating user")
		user, err = models.CreateUser(tx, email, fullName, stripeCustomerID, firebaseSubjectID)
		if err != nil {
			return err
		}

		ac.Logger.Infow("Creating access token", "userID", user.ID)
		accessToken, err = models.CreateAccessToken(tx, user.ID, generateRandomString(128))
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		ac.Logger.Errorf("Error creating user: %w", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, accessToken)
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
