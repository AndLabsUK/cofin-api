package models

import (
	"errors"

	"gorm.io/gorm"
)

const MAX_MESSAGES_UNPAID int64 = 10

type User struct {
	Generic

	Email                     string `gorm:"unique" json:"-"`
	FullName                  string `gorm:"not null" json:"full_name"`
	FirebaseSubjectID         string `gorm:"unique" json:"-"`
	StripeCustomerID          string `gorm:"unique" json:"-"`
	IsSubscribed              bool   `gorm:"not null; default:false" json:"is_subscribed"`
	RemainingMessageAllowance int64  `gorm:"-" sql:"-" json:"remaining_message_allowance"`
}

func GetUserByID(db *gorm.DB, userID uint) (*User, error) {
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if err := setMessageAllowance(db, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func SetUserSubscriptionByStripeCustomerID(db *gorm.DB, stripeCustomerID string, isSubscribed bool) error {
	return db.Model(&User{}).Where("stripe_customer_id = ?", stripeCustomerID).Update("is_subscribed", isSubscribed).Error
}

func GetUserByStripeClientID(db *gorm.DB, stripeCustomerID string) (*User, error) {
	var user User
	if err := db.First(&user, "stripe_customer_id = ?", stripeCustomerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if err := setMessageAllowance(db, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByFirebaseSubjectID(db *gorm.DB, firebaseSubjectID string) (*User, error) {
	var user User
	if err := db.First(&user, "firebase_subject_id = ?", firebaseSubjectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if err := setMessageAllowance(db, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func CreateUser(db *gorm.DB, email, fullName, stripeCustomerID, firebaseSubjectID string) (*User, error) {
	user := &User{
		Email:             email,
		FullName:          fullName,
		FirebaseSubjectID: firebaseSubjectID,
		StripeCustomerID:  stripeCustomerID,
		IsSubscribed:      false,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	if err := setMessageAllowance(db, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Set remaining message allowance counter on the user object.
func setMessageAllowance(db *gorm.DB, user *User) error {
	if user.IsSubscribed {
		user.RemainingMessageAllowance = -1
	} else if messageCount, err := CountUserGenerations(db, user.ID); err != nil {
		return err
	} else {
		if remaining := MAX_MESSAGES_UNPAID - messageCount; remaining > 0 {
			user.RemainingMessageAllowance = remaining
		} else {
			user.RemainingMessageAllowance = 0
		}
	}

	return nil
}
