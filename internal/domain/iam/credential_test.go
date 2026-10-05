package iam

import "testing"

func TestNewCredential(t *testing.T) {
	userID := UserID("user-123")
	hash := "hashed-password"

	credential := NewCredential(userID, hash)

	if credential.UserID() != userID {
		t.Errorf("expected user ID %q, got %q", userID, credential.UserID())
	}

	if credential.Hash() != hash {
		t.Errorf("expected hash %q, got %q", hash, credential.Hash())
	}

	if credential.Version() != 1 {
		t.Errorf("expected version 1, got %d", credential.Version())
	}
}

func TestRestoreCredential(t *testing.T) {
	userID := UserID("user-123")
	hash := "hashed-password"
	version := 5

	credential := RestoreCredential(userID, hash, version)

	if credential.UserID() != userID {
		t.Errorf("expected user ID %q, got %q", userID, credential.UserID())
	}

	if credential.Hash() != hash {
		t.Errorf("expected hash %q, got %q", hash, credential.Hash())
	}

	if credential.Version() != version {
		t.Errorf("expected version %d, got %d", version, credential.Version())
	}
}

func TestCredentialUpdateHash(t *testing.T) {
	credential := NewCredential(UserID("user-123"), "old-hash")

	credential.UpdateHash("new-hash")

	if credential.Hash() != "new-hash" {
		t.Errorf("expected updated hash %q, got %q", "new-hash", credential.Hash())
	}
}
