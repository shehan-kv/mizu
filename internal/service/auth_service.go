package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"mizu/internal/auth"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/auth"
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

// Result of a successful sign-in
type SignInResult struct {
	SessionId string
	Role      string
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
	request *dto.SignInRequest) (*SignInResult, error) {

	correlationId := middleware.GetCorrelationID(ctx)

	authserv.lg.Info("user sign in attempt",
		"event", logger.EventAuthSigninAttempt, "correlation_id", correlationId)

	user, err := authserv.usrSt.GetByEmail(ctx, request.Email)
	if err != nil {
		authserv.lg.Warn("user not found",
			"event", logger.EventAuthUserNotFound, "correlation_id", correlationId)
		return nil, errdefs.ErrAuthUnauthorized
	}

	role, err := authserv.usrSt.GetRoleById(ctx, user.Role)
	if err != nil {
		authserv.lg.Warn("role not found",
			"event", logger.EventAuthRoleNotFound, "correlation_id", correlationId)
		return nil, errdefs.ErrAuthUnauthorized
	}

	if !user.IsActive {
		authserv.lg.Warn(
			"account deactivated",
			"event", logger.EventAuthAccountDisabled,
			"user_id", user.Id, "correlation_id", correlationId)
		return nil, errdefs.ErrAuthUnauthorized
	}

	hash, err := authserv.usrSt.GetPasswordById(ctx, user.Id)
	if err != nil {
		authserv.lg.Warn("password not found",
			"event", logger.EventAuthPasswordNotFound,
			"user_id", user.Id, "correlation_id", correlationId)
		return nil, errdefs.ErrAuthUnauthorized
	}

	isPasswordCorrect := auth.CompareHashAndPassword(hash, request.Password)

	if !isPasswordCorrect {
		authserv.lg.Warn("invalid credentials",
			"event", logger.EventAuthInvalidCredentials,
			"user_id", user.Id, "correlation_id", correlationId)
		return nil, errdefs.ErrAuthUnauthorized
	}

	randomUuid, err := uuid.NewRandom()
	if err != nil {
		authserv.lg.Warn("could not create session id",
			"event", logger.EventSessionIdCreateFailed,
			"user_id", user.Id, "correlation_id", correlationId)
		return nil, errdefs.ErrAuthInternalError
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
			return nil, errdefs.ErrAuthInternalError
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
		return nil, errdefs.ErrAuthInternalError
	}

	authserv.lg.Info("user authenticated",
		"event", logger.EventAuthUserAuthenticated,
		"user_id", user.Id, "correlation_id", correlationId)

	return &SignInResult{SessionId: sessionId, Role: role.Name}, nil
}

// Sign-out a user
// Revokes existing session identified by the cookie.
// This endpoint is idempotent
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - cookie: the existing auth cookie to revoke old session.
//
// Returns:
//   - errdefs.ErrAuthInternalError if internal errors occur
func (authserv *AuthService) SignOut(ctx context.Context, cookie *http.Cookie) error {

	correlationId := middleware.GetCorrelationID(ctx)

	session, err := authserv.sessSt.GetSession(cookie.Value)
	if err != nil {
		// Consider a user is already logged out if session doesn't
		// exist in session store
		if errors.Is(err, errdefs.ErrSessionNotFound) {
			authserv.lg.Warn("previous session doesn't exist",
				"event", logger.EventSessionNotFound,
				"correlation_id", correlationId)
			return nil
		}

		authserv.lg.Warn("could not get previous session",
			"event", logger.EventSessionNotFound,
			"correlation_id", correlationId)
		return errdefs.ErrAuthInternalError
	}

	err = authserv.sessSt.RevokeSession(cookie.Value)
	if err != nil {
		authserv.lg.Warn("could not revoke previous session",
			"event", logger.EventSessionRevokeFailed,
			"user_id", session.UserId, "correlation_id", correlationId)
		return errdefs.ErrAuthInternalError
	}

	return nil
}
