package iam

import (
	"encoding/json"
	"errors"
	"io"
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
	"strings"
)

type IAMHandler struct {
	iamSrv     *iam.Service
	authCookie cookie.AuthCookie
	log        logger.Logger

	maxFileSizeMB int64
}

func NewIAMHandler(
	iamSrv *iam.Service,
	authCookie cookie.AuthCookie,
	log logger.Logger,
	maxFileSizeMB int64,
) *IAMHandler {

	return &IAMHandler{
		iamSrv:        iamSrv,
		authCookie:    authCookie,
		log:           log,
		maxFileSizeMB: maxFileSizeMB,
	}
}

func (h *IAMHandler) NewMux(authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /users", authMiddleware(http.HandlerFunc(h.CreateUser)))
	mux.Handle("GET /users", authMiddleware(http.HandlerFunc(h.ListUsers)))
	mux.Handle("GET /users/me", authMiddleware(http.HandlerFunc(h.GetMe)))
	mux.Handle("GET /users/profile-images/{userID}", authMiddleware(http.HandlerFunc(h.GetProfileImage)))

	// Unauthenticated
	mux.Handle("GET /users/verifications/{verificationID}", http.HandlerFunc(h.ValidateVerification))
	mux.Handle("POST /users/verifications/{verificationID}/confirm", http.HandlerFunc(h.VerifyAccount))

	// Unauthenticated
	mux.Handle("POST /users/recovery", http.HandlerFunc(h.GenerateRecovery))
	mux.Handle("GET /users/recovery/{recoveryToken}", http.HandlerFunc(h.ValidateRecovery))
	mux.Handle("POST /users/recovery/{recoveryToken}/confirm", http.HandlerFunc(h.ConfirmRecovery))

	mux.Handle("POST /users/verifications/{userID}/regenerate-verification", authMiddleware(http.HandlerFunc(h.RegenerateVerification)))
	mux.Handle("PUT /users/{userID}/activate", authMiddleware(http.HandlerFunc(h.ActivateUser)))
	mux.Handle("PUT /users/{userID}/deactivate", authMiddleware(http.HandlerFunc(h.DeactivateUser)))
	mux.Handle("GET /users/{userID}", authMiddleware(http.HandlerFunc(h.GetUser)))
	mux.Handle("PUT /users/{userID}", authMiddleware(http.HandlerFunc(h.UpdateUser)))
	mux.Handle("DELETE /users/{userID}", authMiddleware(http.HandlerFunc(h.DeleteUser)))

	// Unauthenticated
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
		ActorID:    actorID,
		Keyword:    query.ExtractString(r, "q"),
		Role:       query.ExtractString(r, "role"),
		IsActive:   query.ExtractBool(r, "isActive"),
		IsVerified: query.ExtractBool(r, "isVerified"),
		Limit:      p.Limit,
		Offset:     p.Offset,
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

func (h *IAMHandler) GetProfileImage(w http.ResponseWriter, r *http.Request) {

	result, err := h.iamSrv.GetProfileImage(r.Context(), r.PathValue("userID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	defer result.Reader.Close()

	w.Header().Set("Content-Type", result.MimeType)
	w.Header().Set("Content-Disposition", "inline")

	_, err = io.Copy(w, result.Reader)
	if err != nil {
		return
	}
}

func (h *IAMHandler) ValidateRecovery(w http.ResponseWriter, r *http.Request) {

	err := h.iamSrv.RecoveryExists(r.Context(), r.PathValue("recoveryToken"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
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

func (h *IAMHandler) ConfirmRecovery(w http.ResponseWriter, r *http.Request) {

	var req ConfirmRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Password != req.ConfirmPassword {
		response.WriteError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	err := h.iamSrv.ConfirmRecovery(r.Context(), r.PathValue("recoveryToken"), req.Password)
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

func (h *IAMHandler) GenerateRecovery(w http.ResponseWriter, r *http.Request) {
	var req GenerateRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.iamSrv.RegenerateRecovery(r.Context(), req.Email)

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

func (h *IAMHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())
	userID := r.PathValue("userID")

	r.Body = http.MaxBytesReader(w, r.Body, h.maxFileSizeMB<<20)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))

	params := iam.UpdateUserParams{
		ActorID:   actorID,
		UserID:    userID,
		FirstName: r.FormValue("firstName"),
		LastName:  r.FormValue("lastName"),
		Email:     r.FormValue("email"),
		Role:      r.FormValue("role"),
	}

	if title != "" {
		params.Title = &title
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			response.WriteError(w, http.StatusBadRequest, "invalid image upload")
			return
		}
	} else {
		defer file.Close()

		params.Image = &iam.ProfileImageParams{
			FileName: header.Filename,
			MimeType: header.Header.Get("Content-Type"),
			Size:     header.Size,
			Reader:   file,
		}
	}

	err = h.iamSrv.Update(r.Context(), params)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	case errors.Is(err, domainiam.ErrUserImageNotFound):
		response.WriteError(w, http.StatusNotFound, "user image not found")

	case errors.Is(err, domainiam.ErrRecoveryNotFound):
		response.WriteError(w, http.StatusNotFound, "recovery not found")

	case errors.Is(err, domainverification.ErrVerificationNotFound):
		response.WriteError(w, http.StatusNotFound, "user not found")

	// 401
	case errors.Is(err, domainiam.ErrUserInvalidCredentials):
		response.WriteError(w, http.StatusUnauthorized, "invalid credentials")

	case errors.Is(err, domainiam.ErrCredentialNotFound):
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

	case errors.Is(err, domainiam.ErrCredentialConcurrentModification):
		response.WriteError(w, http.StatusConflict, "concurrent modification")

	case errors.Is(err, domainiam.ErrRecoveryConcurrentModification):
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

	case errors.Is(err, domainiam.ErrRecoveryTokenCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "recovery token cannot be empty")

	case errors.Is(err, domainiam.ErrUserInvalidRole):
		response.WriteError(w, http.StatusBadRequest, "invalid role")

	case errors.Is(err, domainiam.ErrUserImageNameCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "invalid image")

	case errors.Is(err, domainiam.ErrUserInvalidImageMimeType):
		response.WriteError(w, http.StatusBadRequest, "invalid image type")

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
		HasImage:   user.HasImage,
		IsActive:   user.IsActive,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
		LastSignIn: user.LastSignIn,
	}
}
