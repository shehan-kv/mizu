package filestore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStore stores files on the local filesystem.
type LocalStore struct {
	root string
}

// New creates a new local filesystem store.
//
// Files are stored inside the "uploads" directory
// relative to the current working directory.
//
// The directory is automatically created if it
// does not already exist.
func NewLocalStore() (*LocalStore, error) {

	root := "uploads"

	err := os.MkdirAll(root, 0o755)
	if err != nil {
		return nil, fmt.Errorf("localfilestore.NewLocalStore: create upload directory: %w", err)
	}

	return &LocalStore{
		root: root,
	}, nil
}

// Save stores the file contents using the provided key
// as the filename.
func (s *LocalStore) Save(_ context.Context, key string, reader io.Reader) error {

	path := filepath.Join(s.root, key)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("localfilestore.Store.Save: create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("localfilestore.Store.Save: write file: %w", err)
	}

	return nil
}

// Open opens a stored file for reading.
//
// The caller must close the returned reader.
func (s *LocalStore) Open(_ context.Context, key string) (io.ReadCloser, error) {

	path := filepath.Join(s.root, key)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("localfilestore.Store.Open: %w", err)
	}

	return file, nil
}

// Delete removes a stored file.
func (s *LocalStore) Delete(_ context.Context, key string) error {

	path := filepath.Join(s.root, key)

	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("localfilestore.Store.Delete: %w", err)
	}

	return nil
}
