package session

import (
	appsession "mizu/internal/application/session"
	"mizu/internal/domain/iam"
	"testing"
)

func newTestSession(t *testing.T, id string, userID string) *appsession.Session {
	t.Helper()

	sessionID, err := appsession.NewSessionID(id)
	if err != nil {
		t.Fatalf("failed to create session ID: %v", err)
	}

	uid, err := iam.NewUserID(userID)
	if err != nil {
		t.Fatalf("failed to create user ID: %v", err)
	}

	return appsession.NewSession(sessionID, uid)
}
