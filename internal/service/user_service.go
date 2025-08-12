package service

import (
	"context"
	"errors"
	"mizu/internal/auth"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/user"
	"mizu/internal/email"
	emlPrms "mizu/internal/email/params"
	"mizu/internal/event"
	"mizu/internal/logger"
	"mizu/internal/middleware"

	"github.com/google/uuid"
)

// Handles user related operations.
type UserService struct {
	lg      logger.Logger
	usrSt   store.UserStore
	emlSndr email.EmailSender
}

// Creates a new instance of UserService.
// It takes a logger, a UserStore for user-related data operations.
//
// Parameters:
//   - lg: logger that implements the logger.Logger interface
//   - usrSt: user store that implements the UserStore interface
//
// Returns:
//   - a pointer to a new UserService
func NewUserService(lg logger.Logger, usrSt store.UserStore, emlSndr email.EmailSender) *UserService {
	return &UserService{
		lg:      lg,
		usrSt:   usrSt,
		emlSndr: emlSndr,
	}
}

// Creates a new user, a user verify request and
// sends an email
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - request: a pointer to UserCreateRequest DTO.
//
// Returns:
//   - ErrAlreadyExists: if user already exists in database.
//   - ErrBadRequest: if required fields are missing.
//   - ErrInternalError: if internal errors occur.
func (usrSrv *UserService) CreateUser(ctx context.Context, request *dto.UserCreateRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)

	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		usrSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err)
		return ErrInternalError
	}

	token, err := uuid.NewRandom()
	if err != nil {
		usrSrv.lg.Error("could not create user verify request token",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)
		return ErrInternalError
	}

	userId, err := usrSrv.usrSt.Onboard(ctx, &params.UserOnboard{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Title:     request.Title,
		Email:     request.Email,
		Role:      request.Role,
		IsActive:  request.IsActive,
		Image:     nil,
		ActorID:   actor.Id,
		Token:     token.String(),
	})

	if err != nil {
		if errors.Is(err, store.ErrUniqueViolation) {
			usrSrv.lg.Error("user already exists",
				"event", event.EventAlreadyExists,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"err", err)
			return ErrAlreadyExists
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			usrSrv.lg.Error("required field not found",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		usrSrv.lg.Error("could not create user",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)
		return ErrInternalError
	}

	usrSrv.lg.Info("user created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", correlationId,
		"scope", "user_service",
		"actor_id", actor.Id,
		"user_id", userId)

	if err := usrSrv.emlSndr.SendVerifyRequest(ctx, &emlPrms.VerifyRequest{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		Token:     token.String(),
	}); err != nil {

		usrSrv.lg.Warn("failed to send user verify request email",
			"event", event.EventEmailSendFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", userId,
			"err", err)
	}

	return nil
}

// Creates a new user verify request and
// sends an email. Replaces the verify request
// if one already exists.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - userId: ID of user to create the request for
//
// Returns:
//   - ErrInternalError: if internal errors occur.
//   - ErrBadRequest: if user not found.
func (usrSrv *UserService) CreateVerifyRequest(ctx context.Context, userId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)

	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		usrSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	user, err := usrSrv.usrSt.GetById(ctx, userId)
	if err != nil {
		usrSrv.lg.Error("user not found to create verify request",
			"event", event.EventNotFound,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", userId,
			"err", err)
		return ErrBadRequest
	}

	token, err := uuid.NewRandom()
	if err != nil {
		usrSrv.lg.Error("could not create user verify request token",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", user.Id,
			"err", err)
		return ErrInternalError
	}

	err = usrSrv.usrSt.CreateOnboardReq(ctx, &params.UserOnboardReqCreate{
		UserId:  user.Id,
		Token:   token.String(),
		IsValid: true,
	})
	if err != nil {
		usrSrv.lg.Error("could not create user verify request",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", user.Id,
			"err", err)
		return ErrInternalError
	}

	usrSrv.lg.Info("user verify request created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", correlationId,
		"scope", "user_service",
		"user_id", user.Id,
		"actor_id", actor.Id)

	if err := usrSrv.emlSndr.SendVerifyRequest(ctx, &emlPrms.VerifyRequest{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Token:     token.String(),
	}); err != nil {

		usrSrv.lg.Error("failed to send verify request email",
			"event", event.EventEmailSendFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"user_id", user.Id,
			"actor_id", actor.Id,
			"err", err)
	}

	return nil
}

// Verifies a newly onboarded user, sets the password and deletes the
// exisisting verification token. The current verfication step requires
// the user to provide a password and confirmation.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - request: a pointer to UserVerifyRequest DTO.
//
// Returns:
//   - ErrInternalError: if internal errors occur.
//   - ErrBadRequest: if user not found.
func (usrSrv *UserService) OnboardVerify(ctx context.Context, token string, request *dto.UserOnboardVerifyRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		usrSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	existingToken, err := usrSrv.usrSt.GetVerifyRequestByToken(ctx, token)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			usrSrv.lg.Error("verification request token does not exist in database",
				"event", event.EventNotFound,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		usrSrv.lg.Error("could not get verification request by token",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		usrSrv.lg.Error("could not hash password",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", existingToken.UserId,
			"err", err)
		return ErrInternalError
	}

	err = usrSrv.usrSt.OnboardVerify(ctx, &params.UserOnboardVerify{
		UserId:         existingToken.UserId,
		HashedPassword: hashedPassword})
	if err != nil {
		usrSrv.lg.Error("could not verify user",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", existingToken.UserId,
			"err", err)
		return ErrInternalError
	}

	return nil
}
