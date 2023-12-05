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

func (user *User) DeleteUser(db *gorm.DB) error {
	if q := db.Delete(&user); q.Error != nil {
		return errors.New("failed to delete user")
	}
	return nil
}

func (user *User) GetLinkCount(db *gorm.DB) int {
	return int(db.Model(*user).Association("Links").Count())
}

func (user *User) FetchLinks(db *gorm.DB, links *[]Link, limit int, offset int, inverseOrdering bool) error {
	var order string
	if inverseOrdering {
		order = "created_at desc"
	} else {
		order = "created_at"
	}
	err := db.Model(*user).Limit(limit).Offset(offset).Order(order).Association("Links").Find(links)
	if err != nil {
		return fmt.Errorf("failed to fetch links")
	}
	return nil
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

func GetUserByUsername(db *gorm.DB, username string) (*User, error) {
	var userRecord User
	if q := db.Where("username = ?", username).First(&userRecord); q.Error != nil {
		return nil, fmt.Errorf("user %s does not exist", username)
	}
	return &userRecord, nil
}

func GetAllUsers(db *gorm.DB) ([]User, error) {
	var users []User
	q := db.Find(&users)
	if q.Error != nil {
		return nil, q.Error
	}
	return users, nil
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
