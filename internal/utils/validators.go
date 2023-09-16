package utils

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

func CheckURLValidity(rawURL string, strict bool) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	if strict && !strings.Contains(parsedURL.Host, ".") {
		return false
	}

	return true
}

func CheckUsernameValidity(username string) error {
	if len(username) < 3 {
		return errors.New("username is too short (minimum length: 3)")
	}

	validUsernameRegex := regexp.MustCompile("^[a-zA-Z0-9_]+$")
	if !validUsernameRegex.MatchString(username) {
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
