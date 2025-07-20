package session

import (
	"mizu/internal/logger"
	"os"
	"strings"
)

// GetSessionStore chooses a session store based on the SESSION environment variable
// Parameters:
//   - lg: a logger that implements the logger.Logger interface
//
// Returns:
//   - an implementation of the SessionStore
func GetSessionStore(lg logger.Logger) SessionStore {
	seDb := strings.TrimSpace(strings.ToLower(os.Getenv("SESSION")))

	switch seDb {
	case "redis":
		// TODO: Implement Redis
		lg.Info("Redis session store not implemented", "event", "session_init_failed")
	default:
		lg.Info("No session store selected")
		lg.Info("Using in-memory session store with default settings")

	}

	// Implement dynamic session store selection.
	// In-memory is always used for now
	return NewInMemorySession()
}
