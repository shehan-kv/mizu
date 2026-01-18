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
	srv *service.UserService
}

// Creates a new instance of UserHandler
//
// Parameters:
//   - srv: a pointer to a UserService
//
// Returns:
//   - a pointer to a new UserHandler
func NewUserHandler(srv *service.UserService) *UserHandler {
	return &UserHandler{
		srv: srv,
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
func (h *UserHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /", mwChain.Handle(h.CreateUser))
	mux.Handle("GET /", mwChain.Handle(h.GetAll))
	mux.Handle("GET /self", mwChain.Handle(h.GetSelf))
	mux.Handle("POST /verify/onboard/{token}", mwChain.Handle(h.OnboardVerify))
	mux.Handle("POST /{userId}/verify-request", mwChain.Handle(h.CreateVerifyRequest))
	mux.Handle("PUT /{userId}/activate", mwChain.Handle(h.Activate))
	mux.Handle("PUT /{userId}/deactivate", mwChain.Handle(h.Deactivate))

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
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

	var createRequest dto.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !createRequest.Validate() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.srv.CreateUser(r.Context(), &createRequest); err != nil {
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
func (h *UserHandler) CreateVerifyRequest(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("userId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.srv.CreateVerifyRequest(r.Context(), parsedId); err != nil {
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
func (h *UserHandler) OnboardVerify(w http.ResponseWriter, r *http.Request) {

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

	err := h.srv.OnboardVerify(r.Context(), token, &verifyRequest)
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

// Handles getting current signed-in user information
//
// Method: GET
//
// Possible Response Codes:
//   - 500 InternalServerError - Server error
//   - 200 OK - Verified successfully
func (h *UserHandler) GetSelf(w http.ResponseWriter, r *http.Request) {

	resp, err := h.srv.GetSelf(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {

	keyword := r.URL.Query().Get("q")
	role := r.URL.Query().Get("role")
	strPage := r.URL.Query().Get("page")
	strLimit := r.URL.Query().Get("limit")

	var page int64
	var limit int64

	if strPage == "" {
		page = 1
	} else {
		parsedPage, err := strconv.ParseInt(strPage, 10, 64)
		if err != nil || parsedPage <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		page = parsedPage
	}

	if strLimit == "" {
		limit = 15
	} else {
		parsedLimit, err := strconv.ParseInt(strLimit, 10, 64)
		if err != nil || parsedLimit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	resp, err := h.srv.GetAll(r.Context(), &dto.UserSearch{
		Keyword: keyword,
		Role:    role,
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) Activate(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("userId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.srv.Activate(r.Context(), parsedId); err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Deactivate(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("userId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.srv.Deactivate(r.Context(), parsedId); err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
