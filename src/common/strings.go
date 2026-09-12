package common

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bp := []byte(password)
	hp, err := bcrypt.GenerateFromPassword(bp, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hp), nil
}

func ComparePasswords(hashedPw string, plainPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPw), []byte(plainPw))
	if err != nil {
		return false
	}
	return true
}

func CheckPassword(password string, minLength int, maxLength int, includeChars bool, includeDigits bool, includeLowercase bool, includeUppercase bool) bool {
	if len(password) < minLength || len(password) > maxLength {
		return false
	}

	if includeChars && !HasLetter(password) {
		return false
	}

	if includeDigits && !HasDigits(password) {
		return false
	}

	if includeLowercase && !HasLower(password) {
		return false
	}

	if includeUppercase && !HasUpper(password) {
		return false
	}

	return true
}

func HasUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func HasLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func HasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func HasDigits(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
