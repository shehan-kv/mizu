package iam

import (
	"encoding/json"
	"errors"
	"mizu/internal/application/iam"
	"mizu/internal/application/logger"
	domainiam "mizu/internal/domain/iam"
	domainverification "mizu/internal/domain/verification"
	"mizu/internal/presentation/http/cookie"
	"mizu/internal/presentation/http/rest/middleware"
	"mizu/internal/presentation/http/rest/page"
	"mizu/internal/presentation/http/rest/query"
	"mizu/internal/presentation/http/rest/response"
	"net/http"
)

type IAMHandler struct {
	iamSrv     *iam.Service
	authCookie cookie.AuthCookie
	log        logger.Logger
}

func NewIAMHandler(iamSrv *iam.Service, authCookie cookie.AuthCookie, log logger.Logger) *IAMHandler {
	return &IAMHandler{iamSrv: iamSrv, authCookie: authCookie, log: log}
}

func (h *IAMHandler) NewMux(authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /users", authMiddleware(http.HandlerFunc(h.CreateUser)))
	mux.Handle("GET /users", authMiddleware(http.HandlerFunc(h.ListUsers)))
	mux.Handle("GET /users/me", authMiddleware(http.HandlerFunc(h.GetMe)))
	mux.Handle("GET /users/verifications/{verificationID}", http.HandlerFunc(h.ValidateVerification))
	mux.Handle("POST /users/verifications/{verificationID}/confirm", http.HandlerFunc(h.VerifyAccount))
	mux.Handle("POST /users/verifications/{userID}/regenerate-verification", authMiddleware(http.HandlerFunc(h.RegenerateVerification)))
	mux.Handle("PUT /users/{userID}/activate", authMiddleware(http.HandlerFunc(h.ActivateUser)))
	mux.Handle("PUT /users/{userID}/deactivate", authMiddleware(http.HandlerFunc(h.DeactivateUser)))
	mux.Handle("GET /users/{userID}", authMiddleware(http.HandlerFunc(h.GetUser)))
	mux.Handle("DELETE /users/{userID}", authMiddleware(http.HandlerFunc(h.DeleteUser)))

	// Auth routes are unauthenticated
	mux.HandleFunc("POST /users/auth/sign-in", h.SignIn)
	mux.HandleFunc("POST /users/auth/sign-out", h.SignOut)

	return mux
}

func (h *IAMHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	extSessionID := h.authCookie.Extract(r)

	result, err := h.iamSrv.SignIn(r.Context(), req.Email, req.Password, extSessionID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	http.SetCookie(w, h.authCookie.New(result.ID, req.RememberMe))
	response.WriteJSON(w, http.StatusOK, SignInResponse{Status: "success", Role: result.Role})
}

func (h *IAMHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	extSessionID := h.authCookie.Extract(r)

	if extSessionID == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.iamSrv.SignOut(r.Context(), *extSessionID); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	http.SetCookie(w, h.authCookie.Expire())
	w.WriteHeader(http.StatusOK)
}

func (h *IAMHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	err := h.iamSrv.CreateUser(r.Context(), iam.CreateUserParams{
		ActorID:    actorID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Title:      req.Title,
		Role:       req.Role,
		IsActive:   req.IsActive,
		ProjectIDs: req.ProjectIDs,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *IAMHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.iamSrv.ListUsers(r.Context(), iam.ListUsersParams{
		ActorID: actorID,
		Keyword: query.ExtractString(r, "q"),
		Role:    query.ExtractString(r, "role"),
		Limit:   p.Limit,
		Offset:  p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	uResp := make([]UserResponse, len(result.Items))
	for i := range result.Items {
		uResp[i] = toUserResponse(&result.Items[i])
	}

	resp := page.PaginatedResponse[UserResponse]{
		Items:      uResp,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *IAMHandler) GetMe(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.iamSrv.GetUser(r.Context(), actorID, actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toUserResponse(result))
}

func (h *IAMHandler) ValidateVerification(w http.ResponseWriter, r *http.Request) {

	err := h.iamSrv.VerificationExists(r.Context(), r.PathValue("verificationID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *IAMHandler) VerifyAccount(w http.ResponseWriter, r *http.Request) {

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Password != req.ConfirmPassword {
		response.WriteError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	err := h.iamSrv.VerifyAccount(r.Context(), r.PathValue("verificationID"), req.Password)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *IAMHandler) GetUser(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.iamSrv.GetUser(r.Context(), actorID, r.PathValue("userID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toUserResponse(result))
}

func (h *IAMHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	err := h.iamSrv.Activate(r.Context(), r.PathValue("userID"), actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *IAMHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	err := h.iamSrv.Deactivate(r.Context(), r.PathValue("userID"), actorID)

	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *IAMHandler) RegenerateVerification(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	err := h.iamSrv.RegenerateVerification(r.Context(), r.PathValue("userID"), actorID)

	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *IAMHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {

	actorID := middleware.ActorIDFromContext(r.Context())

	err := h.iamSrv.Delete(r.Context(), r.PathValue("userID"), actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *IAMHandler) writeServiceError(w http.ResponseWriter, method string, path string, err error) {
	switch {

	// 404
	case errors.Is(err, domainiam.ErrUserNotFound):
		response.WriteError(w, http.StatusNotFound, "user not found")

	case errors.Is(err, domainverification.ErrVerificationNotFound):
		response.WriteError(w, http.StatusNotFound, "user not found")

	// 401
	case errors.Is(err, domainiam.ErrUserInvalidCredentials):
		response.WriteError(w, http.StatusUnauthorized, "invalid credentials")

	// 403
	case errors.Is(err, domainiam.ErrUserInactive):
		response.WriteError(w, http.StatusForbidden, "user is inactive")

	case errors.Is(err, domainiam.ErrUserCannotDeleteSelf):
		response.WriteError(w, http.StatusForbidden, "you cannot delete your own account")

	// 409
	case errors.Is(err, domainiam.ErrUserEmailAlreadyExists):
		response.WriteError(w, http.StatusConflict, "email already exists")

	case errors.Is(err, domainiam.ErrUserAlreadyVerified):
		response.WriteError(w, http.StatusConflict, "user already verified")

	case errors.Is(err, domainiam.ErrUserConcurrentModification):
		response.WriteError(w, http.StatusConflict, "concurrent modification")

	// 400
	case errors.Is(err, domainiam.ErrUserFirstNameCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "firstname cannot be empty")

	case errors.Is(err, domainiam.ErrUserLastNameCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "lastname cannot be empty")

	case errors.Is(err, domainiam.ErrUserInvalidEmailAddress):
		response.WriteError(w, http.StatusBadRequest, "invalid email address")

	case errors.Is(err, domainiam.ErrUserEmailCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "email cannot be empty")

	case errors.Is(err, domainiam.ErrUserPasswordTooShort):
		response.WriteError(w, http.StatusBadRequest, "password too short")

	case errors.Is(err, domainiam.ErrUserPasswordTooLong):
		response.WriteError(w, http.StatusBadRequest, "password too long")

	case errors.Is(err, domainiam.ErrUserIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "user id cannot be empty")

	case errors.Is(err, domainiam.ErrUserInvalidRole):
		response.WriteError(w, http.StatusBadRequest, "invalid role")

	case errors.Is(err, domainverification.ErrVerificationIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "verification id cannot be empty")

	case errors.Is(err, domainverification.ErrVerificationUserIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "verification user id cannot be empty")

	default:
		h.log.Error(
			"internal server error",
			"method", method,
			"path", path,
			"error", err,
		)

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}

func toUserResponse(user *iam.UserDTO) UserResponse {
	return UserResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		Title:      user.Title,
		Role:       user.Role,
		Image:      user.Image,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
	}
}
