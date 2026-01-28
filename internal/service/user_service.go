package service

import (
	"context"
	"errors"
	"mizu/internal/auth"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/dto/common"
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
func (s *UserService) CreateUser(ctx context.Context, request *dto.UserCreateRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)

	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err)
		return ErrInternalError
	}

	token, err := uuid.NewRandom()
	if err != nil {
		s.lg.Error("could not create user verify request token",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)
		return ErrInternalError
	}

	userId, err := s.usrSt.Onboard(ctx, &params.UserOnboard{
		FirstName:  request.FirstName,
		LastName:   request.LastName,
		Title:      request.Title,
		Email:      request.Email,
		Role:       request.Role,
		IsActive:   request.IsActive,
		IsVerified: false,
		Image:      nil,
		ActorID:    actor.Id,
		Token:      token.String(),
		Projects:   request.Projects,
		// Verify if the requesting user is authorized to assign
		// the specified projects to a new user
	})

	if err != nil {
		if errors.Is(err, store.ErrUniqueViolation) {
			s.lg.Error("user already exists",
				"event", event.EventAlreadyExists,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"err", err)
			return ErrAlreadyExists
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			s.lg.Error("required field not found",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		s.lg.Error("could not create user",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)
		return ErrInternalError
	}

	s.lg.Info("user created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", correlationId,
		"scope", "user_service",
		"actor_id", actor.Id,
		"user_id", userId)

	if err := s.emlSndr.SendVerifyRequest(ctx, &emlPrms.VerifyRequest{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		Token:     token.String(),
	}); err != nil {

		s.lg.Warn("failed to send user verify request email",
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
func (s *UserService) CreateVerifyRequest(ctx context.Context, userId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)

	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	user, err := s.usrSt.GetById(ctx, userId)
	if err != nil {
		s.lg.Error("user not found to create verify request",
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
		s.lg.Error("could not create user verify request token",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", user.Id,
			"err", err)
		return ErrInternalError
	}

	err = s.usrSt.CreateOnboardReq(ctx, &params.UserOnboardReqCreate{
		UserId:  user.Id,
		Token:   token.String(),
		IsValid: true,
	})
	if err != nil {
		s.lg.Error("could not create user verify request",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", user.Id,
			"err", err)
		return ErrInternalError
	}

	s.lg.Info("user verify request created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", correlationId,
		"scope", "user_service",
		"user_id", user.Id,
		"actor_id", actor.Id)

	if err := s.emlSndr.SendVerifyRequest(ctx, &emlPrms.VerifyRequest{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Token:     token.String(),
	}); err != nil {

		s.lg.Error("failed to send verify request email",
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
func (s *UserService) OnboardVerify(ctx context.Context, token string, request *dto.UserOnboardVerifyRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	existingToken, err := s.usrSt.GetOnboardReqByToken(ctx, token)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			s.lg.Error("verification request token does not exist in database",
				"event", event.EventNotFound,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"err", err)
			return ErrBadRequest
		}

		s.lg.Error("could not get verification request by token",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err)
		return ErrInternalError
	}

	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		s.lg.Error("could not hash password",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"user_id", existingToken.UserId,
			"err", err)
		return ErrInternalError
	}

	err = s.usrSt.OnboardVerify(ctx, &params.UserOnboardVerify{
		UserId:         existingToken.UserId,
		HashedPassword: hashedPassword})
	if err != nil {
		s.lg.Error("could not verify user",
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

// Gets current signed-in user information.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//
// Returns:
//   - *dto.UserSelfResponse
//   - ErrInternalError: if internal errors occur.
func (s *UserService) GetSelf(ctx context.Context) (*dto.UserSelfResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get user from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)

		return nil, ErrInternalError
	}

	role, err := s.usrSt.GetRoleById(ctx, user.Role)
	if err != nil {
		s.lg.Error("could not get user role",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", user.Id,
			"err", err,
		)

		return nil, ErrInternalError
	}

	selfResponse := dto.UserSelfResponse{
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		Title:      user.Title.String,
		Image:      user.Image.String,
		Role:       role.Name,
		IsVerified: user.IsVerified,
	}

	return &selfResponse, nil
}

func (s *UserService) GetAll(
	ctx context.Context,
	query *dto.UserSearch) (*common.Page[[]dto.UserResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get user from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)

		return nil, ErrInternalError
	}

	search := params.UserSearch{
		Keyword: query.Keyword,
		Role:    query.Role,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	}

	users, err := s.usrSt.GetAll(ctx, &search)
	if err != nil {
		s.lg.Error("could not get users",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", user.Id,
			"err", err,
		)

		return nil, ErrInternalError
	}

	count, err := s.usrSt.CountAll(ctx, &search)
	if err != nil {
		s.lg.Error("could not get user count",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", user.Id,
			"err", err,
		)

		return nil, ErrInternalError
	}

	resp := common.Page[[]dto.UserResponse]{
		Count: count,
		Page:  query.Page,
		Limit: query.Page,
		Data:  make([]dto.UserResponse, len(users)),
	}

	for i, v := range users {
		resp.Data[i] = dto.UserResponse{
			Id:         v.Id,
			FirstName:  v.FirstName,
			LastName:   v.LastName,
			Title:      v.Title,
			Email:      v.Email,
			Role:       v.Role,
			Image:      v.Image,
			CreatedAt:  v.CreatedAt,
			LastLogin:  v.LastLogin,
			IsActive:   v.IsActive,
			IsVerified: v.IsVerified,
		}
	}

	return &resp, nil
}

func (s *UserService) Activate(ctx context.Context, userID int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get user from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)

		return ErrInternalError
	}

	user, err := s.usrSt.GetById(ctx, userID)
	if err != nil {
		s.lg.Error("could not get user",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)

		return ErrInternalError
	}

	if user.IsActive {
		return ErrAlreadyExists
	}

	err = s.usrSt.SetActive(ctx, userID, true)
	if err != nil {
		s.lg.Error("could not activate user",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)

		return ErrInternalError
	}

	return nil
}

func (s *UserService) Deactivate(ctx context.Context, userID int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get user from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)

		return ErrInternalError
	}

	user, err := s.usrSt.GetById(ctx, userID)
	if err != nil {
		s.lg.Error("could not get user",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)

		return ErrInternalError
	}

	if !user.IsActive {
		return ErrAlreadyExists
	}

	err = s.usrSt.SetActive(ctx, userID, false)
	if err != nil {
		s.lg.Error("could not deactivate user",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)

		return ErrInternalError
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, userID int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get user from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)

		return ErrInternalError
	}

	err = s.usrSt.DeleteById(ctx, userID)
	if err != nil {
		s.lg.Error("could not delete user",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)

		return ErrInternalError
	}

	return nil
}

func (s *UserService) SetProjects(ctx context.Context, userID int64, request *dto.ProjectSetRequest) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		s.lg.Error("could not get user from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"err", err,
		)

		return ErrInternalError
	}

	err = s.usrSt.RemoveProjects(ctx, userID)
	if err != nil {
		s.lg.Error("could not remove user projects",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"err", err,
		)

		return ErrInternalError
	}

	if len(request.Projects) > 0 {
		err = s.usrSt.SetProjects(ctx, userID, request.Projects)
		if err != nil {
			if errors.Is(err, store.ErrForeignKeyViolation) {
				s.lg.Error("invalid user projects",
					"event", event.EventInternalError,
					"correlation_id", correlationId,
					"scope", "user_service",
					"actor_id", actor.Id,
					"err", err,
				)
				return ErrBadRequest
			}

			s.lg.Error("could not set user projects",
				"event", event.EventInternalError,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"err", err,
			)

			return ErrInternalError
		}
	}

	return nil
}
