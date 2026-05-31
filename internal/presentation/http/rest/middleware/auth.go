package middleware

import (
	"context"
	"mizu/internal/application/session"
	"mizu/internal/presentation/http/cookie"
	"mizu/internal/presentation/http/rest/response"
	"net/http"
)

type actorIDKey struct{}

func Authenticated(store session.Store, cookie cookie.AuthCookie) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookie.CookieName())
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			sessionID, err := session.NewSessionID(cookie.Value)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			sess, err := store.Get(r.Context(), sessionID)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), actorIDKey{}, sess.UserID().String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ActorIDFromContext extracts the authenticated actor's ID from the request context.
// Should only be called in handlers protected by the Authenticated middleware.
func ActorIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(actorIDKey{}).(string)
	return id
}
