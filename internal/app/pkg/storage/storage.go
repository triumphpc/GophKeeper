// Package storage contain storage implements for keeper
package storage

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/triumphpc/GophKeeper/internal/app/service/user"
	"github.com/triumphpc/GophKeeper/internal/app/service/userdata"
)

// FileInfo contains information of the file
type FileInfo struct {
	ID   uuid.UUID
	Type string
	Path string
	Meta string
}

// Storage interface describe methods for storage
type Storage interface {
	// Close storage connect
	Close()
	// CreateUser user in storage
	CreateUser(u *user.User) error
	// Find  finds a user by login
	Find(login string) (*user.User, error)
	// SaveText save text content for user
	SaveText(ctx context.Context, data *userdata.DataText, userId int) error
	// SaveCard save card content for user
	SaveCard(ctx context.Context, data *userdata.DataCard, userId int) error
	// SaveFile save relation with files
	SaveFile(ctx context.Context, fileInfo FileInfo, userId int) error
}

// FileStorage interface describe methods for file storage
type FileStorage interface {
	// Save saves a new file to the store
	Save(ctx context.Context, fileType string, fileData bytes.Buffer) (*FileInfo, error)
}
