package db

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var validUsernameCharsRE = regexp.MustCompile("^[a-zA-Z0-9_]+$")

func CheckUsernameValidity(username string) error {
	if len(username) < 3 {
		return errors.New("username is too short (minimum length: 3)")
	}

	if len(username) > 32 {
		return errors.New("username is too long (maximum length: 32)")
	}

	if !validUsernameCharsRE.MatchString(username) {
		return errors.New("username must only contain letters, numbers, and underscores")
	}

	return nil
}

func CheckPasswordStrengthValidity(password string) error {
	if len(password) < 6 {
		return errors.New("password is too short (minimum length: 6)")
	}

	hasLetter := false
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return errors.New("password must contain at least one letter")
	}

	hasDigit := false
	for _, char := range password {
		if unicode.IsDigit(char) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}

func VerifyPassword(hashedPassword, inputPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(inputPassword))
	return err == nil
}
