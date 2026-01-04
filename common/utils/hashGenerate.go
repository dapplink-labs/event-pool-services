package utils

import (
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidPasswordA(s string) bool {
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString
	hasLower := regexp.MustCompile(`[a-z]`).MatchString
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString
	hasSpecial := regexp.MustCompile(`[^A-Za-z0-9]`).MatchString

	return hasDigit(s) && hasLower(s) && hasUpper(s) && hasSpecial(s)
}
