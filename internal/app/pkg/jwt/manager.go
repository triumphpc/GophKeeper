// Package jwt manager to generate and verify access token for users.
package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/triumphpc/GophKeeper/internal/app/service/user"
	"time"
)

// Manager is a JSON web token manager
type Manager struct {
	secretKey     string
	tokenDuration time.Duration
	claims        *UserClaims
}

// New returns a new JWT manager
func New(secretKey string, tokenDuration time.Duration) *Manager {
	return &Manager{secretKey, tokenDuration, &UserClaims{}}
}

// UserClaims is a custom JWT claims that contains some user's information
type UserClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Role     string `json:"role"`
	Id       int    `json:"id"`
}

// Claims return current user claims
func (m *Manager) Claims() *UserClaims {
	return m.claims
}

// Generate generates and signs a new token for a user
func (m *Manager) Generate(user *user.User) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(m.tokenDuration).Unix(),
		},
		Username: user.Username,
		Role:     user.Role,
		Id:       user.Id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.secretKey))
}

// Verify verifies the access token string and return a user claim if the token is valid
func (m *Manager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	m.claims = claims

	return claims, nil
}
