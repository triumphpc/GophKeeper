// Package storage contain storage implements for keeper
package storage

import "github.com/triumphpc/GophKeeper/internal/app/service"

// Storage interface describe methods for storage
type Storage interface {
	// Close storage connect
	Close()
	// Save user in storage
	Save(u *service.User) error
	// Find  finds a user by username
	Find(username string) (*service.User, error)
}
