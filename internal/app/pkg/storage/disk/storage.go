package disk

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/storage"
	"os"
	"sync"
)

// FileStore stores files on disk, and its info on memory
type FileStore struct {
	mutex      sync.RWMutex
	fileFolder string
	files      map[string]*storage.FileInfo
}

// New returns a new FileStore
func New(fileFolder string) *FileStore {
	return &FileStore{
		fileFolder: fileFolder,
		files:      make(map[string]*storage.FileInfo),
	}
}

// Save adds a new file
func (store *FileStore) Save(ctx context.Context, fileType string, fileData bytes.Buffer) (*storage.FileInfo, error) {
	fileID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("cannot generate image id: %w", err)
	}

	path := fmt.Sprintf("%s/%s%s", store.fileFolder, fileID, fileType)

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("cannot create file: %w", err)
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		return nil, fmt.Errorf("cannot write image to file: %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	fileInfo := &storage.FileInfo{
		ID:   fileID,
		Type: fileType,
		Path: path,
	}

	store.files[fileID.String()] = fileInfo

	return fileInfo, nil
}
