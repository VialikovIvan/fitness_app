package domain

import (
	"errors"
	"net/mail"
	"strings"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
)

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidPasswordLength = errors.New("invalid password length")
	ErrInvalidEmailFormat    = errors.New("invalid email format")
)

type User struct {
	ID           int64
	FirstName    string
	MiddleName   string
	LastName     string
	Email        string
}

func IsValidPasswordLength(plainPassword string) bool {
	return len(plainPassword) >= MinPasswordLength && len(plainPassword) <= MaxPasswordLength
}

// IsValidEmailFormat reports whether the email format is valid.
// Emails of type Name Lastname <validEmail> or simular are processed like invalid.
func IsValidEmailFormat(email string) bool {
	if strings.ContainsAny(email, "<>\"") {
		return false
	}

	_, err := mail.ParseAddress(email)
	return err == nil
}
