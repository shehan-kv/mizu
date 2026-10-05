package filestore

import (
	"context"
	"io"
)

// Store persists and retrieves binary file data.
type Store interface {
	// Save stores the file stream using the provided key.
	//
	// The key is expected to be unique and stable.
	// Existing files with the same key may be overwritten
	// depending on the implementation.
	Save(ctx context.Context, key string, reader io.Reader) error

	// Open opens a previously stored file for reading.
	//
	// The caller must close the returned reader.
	Open(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes a stored file.
	Delete(ctx context.Context, key string) error
}
