package middleware

import (
	"context"
	"mizu/internal/auth"
	"mizu/internal/db/models"
	"mizu/internal/db/store"
	"mizu/internal/event"
	"mizu/internal/logger"
	"mizu/internal/session"
	"net/http"
)

const userKey key = "user"

// Gets a user stored in the context by Authenticated middleware.
//
// Parameters:
//   - ctx: context to retrieve the user from
//
// Returns:
//   - *models.User if a user is found in context
//   - ErrNotFound: if not found
func GetUserFromContext(ctx context.Context) (*models.User, error) {
	val := ctx.Value(userKey)
	if user, ok := val.(*models.User); ok {
		return user, nil
	}
	return nil, ErrNotFound
}

// Factory function that returns a middleware function that
// verifies the session of a user. If session is valid, fetches
// user from database and stores in context.
//
// Parameters:
//   - lg: an implementation of logger.Logger
//   - seStore: an implementation of session.SessionStore
//   - usrStore: an implementation of store.UserStore
//
// Returns:
//   - a function that takes a http.HandlerFunc as parameter and returns a http.HandlerFunc
//
// Possible Response Codes:
//   - 401 Unauthorized - If session doesn't exist or fetching user from store fails
func Authenticated(lg logger.Logger, seStore session.SessionStore, usrStore store.UserStore) func(http.HandlerFunc) http.HandlerFunc {

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cid := GetCorrelationID(r.Context())

			cookie, err := r.Cookie(auth.AuthCookieName)
			if err != nil {
				lg.Warn("unauthenticated user",
					"event", event.EventUserNotAuthenticated,
					"correlation_id", cid,
					"scope", "middleware_authenticated",
					"err", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			session, err := seStore.GetSession(cookie.Value)
			if err != nil {
				lg.Warn("session not found",
					"event", event.EventNotFound,
					"correlation_id", cid,
					"scope", "middleware_authenticated",
					"err", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := usrStore.GetById(r.Context(), session.UserId)
			if err != nil {
				lg.Warn("could not retrieve user from database",
					"event", event.EventNotFound,
					"correlation_id", cid,
					"scope", "middleware_authenticated",
					"err", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)
			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}
