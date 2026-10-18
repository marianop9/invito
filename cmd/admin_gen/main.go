package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash from a plain-text password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plain-text password with a stored hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func main() {
	password := "Alemania2012"

	// Hash the password
	hash, err := HashPassword(password)
	if err != nil {
		panic(err)
	}

	fmt.Println("Hash:", hash)

	// Verify the password
	match := CheckPasswordHash(password, hash)
	fmt.Println("Match:", match) // true
}
