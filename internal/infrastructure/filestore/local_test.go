package filestore

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStore(t *testing.T) {
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	store, err := NewLocalStore()
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}

	t.Run("creates uploads directory", func(t *testing.T) {
		info, err := os.Stat(store.root)
		if err != nil {
			t.Fatalf("failed to stat uploads directory: %v", err)
		}

		if !info.IsDir() {
			t.Fatal("expected uploads to be a directory")
		}
	})

	t.Run("saves file contents", func(t *testing.T) {
		key := "test.txt"
		content := "hello, world"

		if err := store.Save(
			context.Background(),
			key,
			strings.NewReader(content),
		); err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(store.root, key))
		if err != nil {
			t.Fatalf("failed to read saved file: %v", err)
		}

		if string(data) != content {
			t.Fatalf("expected %q, got %q", content, string(data))
		}
	})

	t.Run("opens saved file", func(t *testing.T) {
		key := "open.txt"
		content := "hello, world"

		if err := store.Save(
			context.Background(),
			key,
			strings.NewReader(content),
		); err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		reader, err := store.Open(context.Background(), key)
		if err != nil {
			t.Fatalf("Open() returned error: %v", err)
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(data) != content {
			t.Fatalf("expected %q, got %q", content, string(data))
		}
	})

	t.Run("deletes saved file", func(t *testing.T) {
		key := "delete.txt"

		if err := store.Save(
			context.Background(),
			key,
			strings.NewReader("content"),
		); err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		if err := store.Delete(context.Background(), key); err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(store.root, key)); !os.IsNotExist(err) {
			t.Fatal("expected file to be deleted")
		}
	})
}
