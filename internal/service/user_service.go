package service

import (
	"context"
	"database/sql"
	"errors"
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

// Creates a new user, an onboarding request and
// sends an email
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - request: a pointer to UserCreateRequest DTO.
//
// Returns:
//   - ErrAlreadyExists: if user already exists in database.
//   - ErrInternalError: if internal errors occur.
func (usrSrv *UserService) CreateUser(ctx context.Context, request *dto.UserCreateRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)

	id, err := usrSrv.usrSt.CreateOne(ctx, &params.UserCreateParams{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Title:     sql.NullString{String: request.Title},
		Email:     request.Email,
		Role:      request.Role,
		IsActive:  request.IsActive,
		Image:     sql.NullString{Valid: false},
	})

	if err != nil {
		if errors.Is(err, store.ErrUniqueViolation) {
			usrSrv.lg.Error("user already exists",
				"event", event.EventAlreadyExists,
				"correlation_id", correlationId,
				"scope", "user_service",
				"err", err)
			return ErrAlreadyExists
		}

		usrSrv.lg.Error("could not create user",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)
		return ErrInternalError
	}

	usrSrv.lg.Info("user created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", correlationId,
		"scope", "user_service",
		"user_id", id)

	token, err := uuid.NewRandom()
	if err != nil {
		usrSrv.lg.Error("could not create user onboard request token",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)
		return ErrInternalError
	}

	if err := usrSrv.usrSt.CreateOnboardRequest(ctx, &params.UserOnboardRequestCreateParams{
		UserId:  id,
		Token:   token.String(),
		IsValid: true,
	}); err != nil {

		usrSrv.lg.Error("could not create user onboard request",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)
		return ErrInternalError
	}

	usrSrv.lg.Info("user onboard request created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", correlationId,
		"scope", "user_service",
		"user_id", id)

	if err := usrSrv.emlSndr.SendOnboardingRequest(ctx, &emlPrms.OnboardingRequest{
		FirstName:     request.FirstName,
		LastName:      request.LastName,
		Email:         request.Email,
		Token:         token.String(),
		CorrelationId: correlationId,
	}); err != nil {

		usrSrv.lg.Error("failed to send onboarding request email",
			"event", event.EventEmailSendFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"user_id", id)
		return ErrInternalError
	}

	return nil
}
