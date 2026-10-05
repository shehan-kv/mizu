package auth

import (
	"strings"
	"testing"

	"mizu/internal/domain/iam"
)

func newTestPlainPassword(t *testing.T, value string) iam.PlainPassword {
	t.Helper()

	password, err := iam.NewPlainPassword(value)
	if err != nil {
		t.Fatal(err)
	}

	return password
}

func TestDefaultPasswordHasher_Hash(t *testing.T) {
	hasher := NewDefaultPasswordHasher()
	password := newTestPlainPassword(t, "correct-password")

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() returned an error: %v", err)
	}

	if !strings.HasPrefix(hash, bcryptHashPrefix) {
		t.Fatalf("Hash() = %q, expected prefix %q", hash, bcryptHashPrefix)
	}

	if len(hash) <= len(bcryptHashPrefix) {
		t.Fatalf("Hash() = %q, expected a bcrypt hash after the prefix", hash)
	}
}

func TestDefaultPasswordHasher_Verify(t *testing.T) {
	hasher := NewDefaultPasswordHasher()
	password := newTestPlainPassword(t, "correct-password")

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() returned an error: %v", err)
	}

	t.Run("returns true for correct password", func(t *testing.T) {
		if !hasher.Verify(hash, password) {
			t.Fatal("Verify() = false, expected true")
		}
	})

	t.Run("returns false for incorrect password", func(t *testing.T) {
		wrongPassword := newTestPlainPassword(t, "wrong-password")

		if hasher.Verify(hash, wrongPassword) {
			t.Fatal("Verify() = true, expected false")
		}
	})
}

func TestDefaultPasswordHasher_Verify_UnsupportedHashPrefix(t *testing.T) {
	hasher := NewDefaultPasswordHasher()
	password := newTestPlainPassword(t, "correct-password")

	if hasher.Verify("99some-hash", password) {
		t.Fatal("Verify() = true, expected false for unsupported hash prefix")
	}
}

func TestDefaultPasswordHasher_Verify_MalformedHash(t *testing.T) {
	hasher := NewDefaultPasswordHasher()
	password := newTestPlainPassword(t, "correct-password")

	tests := []struct {
		name string
		hash string
	}{
		{
			name: "empty hash",
			hash: "",
		},
		{
			name: "one character hash",
			hash: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if hasher.Verify(tt.hash, password) {
				t.Fatal("Verify() = true, expected false")
			}
		})
	}
}

func TestDefaultPasswordHasher_Hash_GeneratesDifferentHashes(t *testing.T) {
	hasher := NewDefaultPasswordHasher()
	password := newTestPlainPassword(t, "same-password")

	hash1, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("first Hash() returned an error: %v", err)
	}

	hash2, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("second Hash() returned an error: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("Hash() generated identical hashes for the same password")
	}
}
