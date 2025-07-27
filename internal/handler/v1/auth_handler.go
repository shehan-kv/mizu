package v1

import (
	"encoding/json"
	"errors"
	"mizu/internal/auth"
	dto "mizu/internal/dto/auth"
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
	mux.Handle("POST /sign-out", mwChain.Handle(athHndl.SignOut))

	return mux
}

// Handles user signing-in.
//
// Expects a JSON body of SignInRequest DTO.
//
// Method: POST
//
// Returns:
//   - SignInResponse DTO on successful sign-in
//
// Possible Response Codes:
//   - 400 BadRequest – Invalid input or missing fields
//   - 401 StatusUnauthorized - If user isn't permitted to sign-in
//   - 500 InternalServerError - Server error
//   - 200 OK - Signed-in successfully
func (athHndl *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {

	var signinRequest dto.SignInRequest

	if err := json.NewDecoder(r.Body).Decode(&signinRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isValid := signinRequest.Validate(); !isValid {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cookie, _ := r.Cookie(auth.AuthCookieName)
	result, err := athHndl.authSrv.SignIn(r.Context(), cookie, &signinRequest)

	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if errors.Is(err, service.ErrInternalError) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, auth.GetAuthCookie(result.SessionId, signinRequest.RememberMe))
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.SignInResponse{Status: "success", Role: result.Role})
}

// Handles user signing-out.
//
// Method: POST
//
// Possible Response Codes:
//   - 500 InternalServerError - Server error
//   - 200 OK - Signed-out successfully
func (athHndl *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie(auth.AuthCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = athHndl.authSrv.SignOut(r.Context(), cookie); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, auth.GetAuthDeleteCookie())
	w.WriteHeader(http.StatusOK)
}
