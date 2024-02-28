package database

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Link struct {
	ID        uint      `db:"id"`
	URL       string    `db:"url"`
	Visits    uint      `db:"visits"`
	CreatedBy string    `db:"created_by"`
	CreatedAt time.Time `db:"created_at"`
}

type User struct {
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	IsAdmin   bool      `db:"is_admin"`
	CreatedAt time.Time `db:"created_at"`
}

func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false
	}
	return true
}
