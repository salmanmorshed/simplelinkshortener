package database

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
	IsAdmin  bool
	Links    []Link `gorm:"foreignKey:UserID"`
}

func (user *User) UpdateUsername(db *gorm.DB, newUsername string) error {
	if q := db.Model(&user).Update("username", newUsername); q.Error != nil {
		return errors.New("failed to update username")
	}
	return nil
}

func (user *User) UpdatePassword(db *gorm.DB, newPassword string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}
	if q := db.Model(&user).Update("password", string(hashedBytes)); q.Error != nil {
		return errors.New("failed to update password")
	}
	return nil
}

func (user *User) ToggleAdmin(db *gorm.DB) error {
	if q := db.Model(&user).Update("is_admin", gorm.Expr("NOT is_admin")); q.Error != nil {
		return errors.New("failed to update admin status")
	}
	return nil
}

func (user *User) Delete(db *gorm.DB) error {
	if q := db.Delete(&user); q.Error != nil {
		return errors.New("failed to delete user")
	}
	return nil
}

func GetAllUsers(db *gorm.DB) ([]User, error) {
	var users []User
	q := db.Find(&users)
	if q.Error != nil {
		return nil, q.Error
	}
	return users, nil
}

func GetUserByID(db *gorm.DB, userID uint) (*User, error) {
	var userRecord User
	if q := db.First(&userRecord, userID); q.Error != nil {
		return nil, fmt.Errorf("user #%d does not exist", userID)
	}
	return &userRecord, nil
}

func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var userRecord User
	if q := db.Where("username = ?", username).First(&userRecord); q.Error != nil {
		return nil, fmt.Errorf("user %s does not exist", username)
	}
	return &userRecord, nil
}

func AuthenticateUser(db *gorm.DB, username string, password string) (*User, error) {
	var userRecord User
	if q := db.Where("username = ?", username).First(&userRecord); q.Error != nil {
		return nil, fmt.Errorf("failed to find user with given username: %s", username)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userRecord.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password for username: %s", username)
	}
	return &userRecord, nil
}

func CreateNewUser(db *gorm.DB, username string, password string) (*User, error) {
	var count int64
	q1 := db.Unscoped().Model(&User{}).Where("username = ?", username).Count(&count)
	if q1.Error == nil && count > 0 {
		return nil, fmt.Errorf("%s is already taken", username)
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashedPassword := string(hashedBytes)
	newUser := User{Username: username, Password: hashedPassword}
	if q := db.Create(&newUser); q.Error != nil {
		return nil, q.Error
	}
	return &newUser, nil
}
