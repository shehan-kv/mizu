package v1

import (
	"encoding/json"
	"errors"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/user"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
	"strconv"
)

// Handles user-related HTTP requests.
//
// Uses an UserService to perform
// user operations
type UserHandler struct {
	usrSrv *service.UserService
}

// Creates a new instance of UserHandler
//
// Parameters:
//   - usrSrv: a pointer to a UserService
//
// Returns:
//   - a pointer to a new UserHandler
func NewUserHandler(usrSrv *service.UserService) *UserHandler {
	return &UserHandler{
		usrSrv: usrSrv,
	}
}

// Creates a ServeMux for the user routes and middleware.
// Defines the routes and handler function for each route.
// Registers middleware for the routes.
//
// Parameters:
//   - lg: an implementation of logger.Logger
//   - seSt: an implementation of session.SessionStore
//   - usrSt: an implementation of store.UserStore
//
// Returns:
//   - a *http.ServeMux
func (usrHndl *UserHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /", mwChain.Handle(usrHndl.CreateUser))
	mux.Handle("POST /verify/onboard/{token}", mwChain.Handle(usrHndl.OnboardVerify))
	mux.Handle("POST /{userId}/verify-request", mwChain.Handle(usrHndl.CreateVerifyRequest))

	return mux
}

// Handles creating a user.
//
// Expects a JSON body of UserCreateRequest DTO.
//
// Method: POST
//
// Possible Response Codes:
//   - 400 BadRequest – Invalid input, missing fields or constraint violations
//   - 409 Conflict - Already exists
//   - 500 InternalServerError - Server error
//   - 201 OK - Created successfully
func (usrHndl *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

	var createRequest dto.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !createRequest.Validate() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := usrHndl.usrSrv.CreateUser(r.Context(), &createRequest); err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Handles creating a user verify request.
//
// Method: POST
//
// Possible Response Codes:
//   - 400 BadRequest – Invalid input or user doesn't exist in the database
//   - 500 InternalServerError - Server error
//   - 201 OK - Created successfully
func (usrHndl *UserHandler) CreateVerifyRequest(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("userId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := usrHndl.usrSrv.CreateVerifyRequest(r.Context(), parsedId); err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Handles verifying a newly onboarded user account.
// Currently the user has to provide a password
// and password confirmation during the verification step
//
// Method: POST
//
// Possible Response Codes:
//   - 400 BadRequest – Invalid input or user doesn't exist in the database
//   - 500 InternalServerError - Server error
//   - 200 OK - Verified successfully
func (usrHndl *UserHandler) OnboardVerify(w http.ResponseWriter, r *http.Request) {

	token := r.PathValue("token")

	var verifyRequest dto.UserOnboardVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&verifyRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !verifyRequest.Validate() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := usrHndl.usrSrv.OnboardVerify(r.Context(), token, &verifyRequest)
	if err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
