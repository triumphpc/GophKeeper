// Package jwt manager to generate and verify access token for users.
package jwt

import (
	"github.com/golang-jwt/jwt"
	"time"
)

// Manager is a JSON web token manager
type Manager struct {
	secretKey     string
	tokenDuration time.Duration
}

// New returns a new JWT manager
func New(secretKey string, tokenDuration time.Duration) *Manager {
	return &Manager{secretKey, tokenDuration}
}

// UserClaims is a custom JWT claims that contains some user's information
type UserClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
}
