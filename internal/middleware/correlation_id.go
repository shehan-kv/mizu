package middleware

import (
	"context"
	"mizu/internal/event"
	"mizu/internal/logger"
	"net/http"

	"github.com/google/uuid"
)

// For the context
type key string

const correlationIDKey key = "correlationID"
const correlationHeader = "X-Correlation-ID"

// Gets correlation id from the context
//
// Parameters:
//   - ctx: context to get the correlation id from
//
// Returns:
//   - the correlation id stored in the context
//   - empty string if not found
func GetCorrelationID(ctx context.Context) string {
	val := ctx.Value(correlationIDKey)
	if cid, ok := val.(string); ok {
		return cid
	}
	return ""
}

// Factory function that returns the correlation id middleware.
//
// Parameters:
//   - lg: a logger that implements the logger.Logger interface
//
// Returns:
//   - a middleware function that adds a correlation id
func CorrelationId(lg logger.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cid := r.Header.Get(correlationHeader)
			if cid == "" {
				randomUuid, err := uuid.NewRandom()
				if err != nil {
					lg.Error("could not create correlation id",
						"event", event.EventInternalError,
						"scope", "middleware_correlation_id",
						"err", err)

					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				cid = randomUuid.String()
				w.Header().Set(correlationHeader, cid)
				ctx := context.WithValue(r.Context(), correlationIDKey, cid)
				r = r.WithContext(ctx)
			}

			next(w, r)
		}
	}
}
