package v1

import (
	"encoding/json"
	"errors"
	"mizu/internal/auth"
	"mizu/internal/dto"
	"mizu/internal/errdefs"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"net/http"
)

// Handles authentication-related HTTP requests.
//
// Uses an AuthService to perform
// authentication operations
type AuthHandler struct {
	authSrv *service.AuthService
}

// Creates a new instance of AuthHandler
//
// Parameters:
//   - authSrv: a pointer to a AuthService
//
// Returns:
//   - a pointer to a new AuthHandler
func NewAuthHandler(authSrv *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authSrv: authSrv,
	}
}

// Creates a ServeMux for the authentication routes and middleware.
// Defines the routes and handler function for each route.
// Registers middleware for the routes.
//
// Parameters:
//   - lg: an implementation of logger.Logger
//
// Returns:
//   - a *http.ServeMux
func (athHndl *AuthHandler) GetMux(lg logger.Logger) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(middleware.CorrelationId(lg))

	mux := http.NewServeMux()

	mux.Handle("POST /sign-in", mwChain.Handle(athHndl.SignIn))

	return mux
}

// Handles user sign in by email and a password.
// Complies with the http.HandlerFunc.
//
// Parameters:
//   - w: http.ResponseWriter
//   - r: *http.Request
func (athHndl *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {

	var signinRequest dto.UserSignInRequest

	err := json.NewDecoder(r.Body).Decode(&signinRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isValid := signinRequest.Validate()
	if !isValid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cookie, _ := r.Cookie(auth.AuthCookieName)
	token, err := athHndl.authSrv.SignIn(r.Context(), cookie, &signinRequest)

	if err != nil {
		if errors.Is(err, errdefs.ErrAuthUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if errors.Is(err, errdefs.ErrAuthInternalError) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, auth.GetAuthCookie(token, signinRequest.RememberMe))
	w.WriteHeader(http.StatusOK)
}
