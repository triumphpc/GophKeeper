// Package service implement User entity
package service

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// User contains user's information
type User struct {
	Username       string
	HashedPassword string
}

// NewUser returns a new user
func NewUser(username string, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	user := &User{
		Username:       username,
		HashedPassword: string(hashedPassword),
	}

	return user, nil
}

// IsCorrectPassword checks if the provided password is correct or not
func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil
}
