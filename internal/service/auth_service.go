package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"mizu/internal/auth"
	"mizu/internal/db/store"
	"mizu/internal/dto"
	"mizu/internal/errdefs"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/session"
)

// Handles user authentication including signing in and creating/revoking sessions.
type AuthService struct {
	lg     logger.Logger
	usrSt  store.UserStore
	sessSt session.SessionStore
}

// Creates a new instance of AuthService.
// It takes a logger, a UserStore for user-related data operations,
// and a SessionStore for session management.
//
// Parameters:
//   - lg: logger that implements the logger.Logger interface
//   - usrSt: user store that implements the UserStore interface
//   - sessSt: session store that implements the SessionStore interface
//
// Returns:
//   - a pointer to a new AuthService
func NewAuthService(lg logger.Logger, usrSt store.UserStore, sessSt session.SessionStore) *AuthService {
	return &AuthService{
		lg:     lg,
		usrSt:  usrSt,
		sessSt: sessSt,
	}
}

// Authenticates a user using an email and a password.
// Manages session creation and revocation of existing sessions.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - cookie: the existing auth cookie, if any, to revoke old sessions.
//   - request: a pointer to a UserSignInRequest DTO
//
// Returns:
//   - session ID string upon successful authentication
//   - errdefs.ErrAuthUnauthorized if authentication fails
//   - errdefs.ErrAuthInternalError if internal errors occur
func (authserv *AuthService) SignIn(
	ctx context.Context,
	cookie *http.Cookie,
	request *dto.UserSignInRequest) (string, error) {

	correlationId := middleware.GetCorrelationID(ctx)

	authserv.lg.Info("user sign in attempt",
		"event", logger.EventAuthSigninAttempt, "correlation_id", correlationId)

	user, err := authserv.usrSt.GetByEmail(ctx, request.Email)
	if err != nil {
		authserv.lg.Warn("user not found",
			"event", logger.EventAuthUserNotFound, "correlation_id", correlationId)
		return "", errdefs.ErrAuthUnauthorized
	}

	if !user.IsActive {
		authserv.lg.Warn(
			"account deactivated",
			"event", logger.EventAuthAccountDisabled,
			"user_id", user.Id, "correlation_id", correlationId)
		return "", errdefs.ErrAuthUnauthorized
	}

	hash, err := authserv.usrSt.GetPasswordById(ctx, user.Id)
	if err != nil {
		authserv.lg.Warn("password not found",
			"event", logger.EventAuthPasswordNotFound,
			"user_id", user.Id, "correlation_id", correlationId)
		return "", errdefs.ErrAuthUnauthorized
	}

	isPasswordCorrect := auth.CompareHashAndPassword(hash, request.Password)

	if !isPasswordCorrect {
		authserv.lg.Warn("invalid credentials",
			"event", logger.EventAuthInvalidCredentials,
			"user_id", user.Id, "correlation_id", correlationId)
		return "", errdefs.ErrAuthUnauthorized
	}

	randomUuid, err := uuid.NewRandom()
	if err != nil {
		authserv.lg.Warn("could not create session id",
			"event", logger.EventSessionIdCreateFailed,
			"user_id", user.Id, "correlation_id", correlationId)
		return "", errdefs.ErrAuthInternalError
	}

	sessionId := randomUuid.String()

	// If an existing auth cookie is found,
	// revoke the old session id.
	if cookie != nil {
		err = authserv.sessSt.RevokeSession(cookie.Value)
		if err != nil {
			authserv.lg.Warn("could not revoke previous session",
				"event", logger.EventSessionRevokeFailed,
				"user_id", user.Id, "correlation_id", correlationId)
			return "", errdefs.ErrAuthInternalError
		}
	}

	err = authserv.usrSt.UpdateLastLogin(ctx, user.Id)
	if err != nil {
		authserv.lg.Warn("could not update last login",
			"event", logger.EventAuthLastSigninNotUpdated,
			"user_id", user.Id, "correlation_id", correlationId)
	}

	// If there's no existing cookie,
	// create a new session
	err = authserv.sessSt.SetSession(sessionId, user.Id)
	if err != nil {
		authserv.lg.Warn("could not set session",
			"event", logger.EventSessionSetFailed,
			"user_id", user.Id, "correlation_id", correlationId)
		return "", errdefs.ErrAuthInternalError
	}

	authserv.lg.Info("user authenticated",
		"event", logger.EventAuthUserAuthenticated,
		"user_id", user.Id, "correlation_id", correlationId)
	return sessionId, nil
}
