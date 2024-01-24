package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	IsAdmin   bool      `db:"is_admin"`
	CreatedAt time.Time `db:"created_at"`
}

func CreateUser(db *sqlx.DB, username string, password string) (*User, error) {
	var count uint
	q1 := db.Rebind("SELECT count(*) FROM users where username = ?")
	err := db.Get(&count, q1, username)
	if err != nil {
		return nil, errors.New("failed to check username")
	}
	if count > 0 {
		return nil, fmt.Errorf("%s is already taken", username)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	var user User
	q2 := db.Rebind("INSERT INTO users (username, password) VALUES (?, ?) RETURNING *")
	err = db.Get(&user, q2, username, string(hashedBytes))
	if err != nil {
		return nil, errors.New("failed to create new user")
	}

	return &user, nil
}

func RetrieveAllUsers(db *sqlx.DB) ([]User, error) {
	var users []User
	err := db.Select(&users, "SELECT * FROM users")
	if err != nil {
		return nil, errors.New("failed to retrieve users")
	}
	return users, nil
}

func RetrieveUser(db *sqlx.DB, username string) (*User, error) {
	var user User
	q := db.Rebind("SELECT * FROM users WHERE username = ?")
	err := db.Get(&user, q, username)
	if err != nil {
		return nil, errors.New("failed to retrieve user")
	}
	return &user, nil
}

func (user *User) UpdateUsername(db *sqlx.DB, newUsername string) error {
	q := db.Rebind("UPDATE users SET username = ? WHERE username = ? RETURNING *")
	err := db.Get(user, q, newUsername, user.Username)
	if err != nil {
		return errors.New("failed to update username")
	}
	return nil
}

func (user *User) UpdatePassword(db *sqlx.DB, newPassword string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	q := db.Rebind("UPDATE users SET password = ? WHERE username = ? RETURNING *")
	err = db.Get(user, q, string(hashedBytes), user.Username)
	if err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (user *User) ToggleAdmin(db *sqlx.DB) error {
	q := db.Rebind("UPDATE users SET is_admin = NOT is_admin WHERE username = ? RETURNING *")
	err := db.Get(user, q, user.Username)
	if err != nil {
		return errors.New("failed to update admin status")
	}
	return nil
}

func (user *User) Delete(db *sqlx.DB) error {
	q := db.Rebind("DELETE from users WHERE username = ?")
	_, err := db.Exec(q, user.Username)
	if err != nil {
		return errors.New("failed to delete user")
	}
	return nil
}

func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false
	}
	return true
}
