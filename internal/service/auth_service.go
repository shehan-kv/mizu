package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"mizu/internal/auth"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/auth"
	"mizu/internal/event"
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
//   - *SignInResult: a pointer to SignInResult
//   - ErrUnauthorized: if authentication fails
//   - ErrInternalError: if internal errors occur
func (s *AuthService) SignIn(
	ctx context.Context,
	cookie *http.Cookie,
	request *dto.SignInRequest) (*SignInResult, error) {

	correlationId := middleware.GetCorrelationID(ctx)

	s.lg.Info("user sign in attempt",
		"event", event.EventSigninAttempt,
		"correlation_id", correlationId,
		"scope", "auth_service")

	user, err := s.usrSt.GetByEmail(ctx, request.Email)
	if err != nil {
		s.lg.Warn("user not found",
			"event", event.EventNotFound,
			"correlation_id", correlationId,
			"scope", "auth_service",
			"err", err)
		return nil, ErrUnauthorized
	}

	role, err := s.usrSt.GetRoleById(ctx, user.Role)
	if err != nil {
		s.lg.Warn("role not found",
			"event", event.EventNotFound,
			"correlation_id", correlationId,
			"scope", "auth_service",
			"err", err)
		return nil, ErrUnauthorized
	}

	if !user.IsActive {
		s.lg.Warn("account deactivated",
			"event", event.EventAccountDisabled,
			"user_id", user.Id,
			"correlation_id", correlationId,
			"scope", "auth_service")
		return nil, ErrUnauthorized
	}

	hash, err := s.usrSt.GetPasswordById(ctx, user.Id)
	if err != nil {
		s.lg.Warn("password not found",
			"event", event.EventNotFound,
			"user_id", user.Id,
			"correlation_id", correlationId,
			"scope", "auth_service",
			"err", err)
		return nil, ErrUnauthorized
	}

	isPasswordCorrect := auth.CompareHashAndPassword(hash, request.Password)

	if !isPasswordCorrect {
		s.lg.Warn("invalid credentials",
			"event", event.EventInvalidCredentials,
			"user_id", user.Id,
			"correlation_id", correlationId,
			"scope", "auth_service")
		return nil, ErrUnauthorized
	}

	randomUuid, err := uuid.NewRandom()
	if err != nil {
		s.lg.Warn("could not create session id",
			"event", event.EventCreateFailed,
			"user_id", user.Id,
			"correlation_id", correlationId,
			"scope", "auth_service",
			"err", err)
		return nil, ErrInternalError
	}

	sessionId := randomUuid.String()

	// If an existing auth cookie is found,
	// revoke the old session id.
	if cookie != nil {
		err = s.sessSt.RevokeSession(cookie.Value)
		if err != nil {
			s.lg.Warn("could not revoke previous session",
				"event", event.EventInternalError,
				"user_id", user.Id,
				"correlation_id", correlationId,
				"scope", "auth_service",
				"err", err)
			return nil, ErrInternalError
		}
	}

	err = s.usrSt.UpdateLastLogin(ctx, user.Id)
	if err != nil {
		s.lg.Warn("could not update last login",
			"event", event.EventInternalError,
			"user_id", user.Id,
			"correlation_id", correlationId,
			"scope", "auth_service",
			"err", err)
	}

	// If there's no existing cookie,
	// create a new session
	err = s.sessSt.SetSession(sessionId, user.Id)
	if err != nil {
		s.lg.Warn("could not set session",
			"event", event.EventInternalError,
			"user_id", user.Id,
			"correlation_id", correlationId,
			"scope", "auth_service",
			"err", err)
		return nil, ErrInternalError
	}

	s.lg.Info("user authenticated",
		"event", event.EventUserAuthenticated,
		"user_id", user.Id,
		"correlation_id", correlationId,
		"scope", "auth_service")

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
//   - AuthInternalError: if internal errors occur
func (s *AuthService) SignOut(ctx context.Context, cookie *http.Cookie) error {

	correlationId := middleware.GetCorrelationID(ctx)

	existingSession, err := s.sessSt.GetSession(cookie.Value)
	if err != nil {
		// Consider a user is already logged out if session doesn't
		// exist in session store
		if errors.Is(err, session.ErrNotFound) {
			s.lg.Warn("previous session doesn't exist",
				"event", event.EventNotFound,
				"correlation_id", correlationId,
				"scope", "auth_service",
				"err", err)
			return nil
		}

		s.lg.Warn("could not get previous session",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "auth_service",
			"err", err)
		return ErrInternalError
	}

	err = s.sessSt.RevokeSession(cookie.Value)
	if err != nil {
		s.lg.Warn("could not revoke previous session",
			"event", event.EventInternalError,
			"user_id", existingSession.UserId,
			"correlation_id", correlationId,
			"scope", "auth_service",
			"err", err)
		return ErrInternalError
	}

	return nil
}
