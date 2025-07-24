package service

import (
	"context"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/project"
	"mizu/internal/errdefs"
	"mizu/internal/logger"
	"mizu/internal/middleware"
)

// Handles project related operations.
type ProjectService struct {
	lg    logger.Logger
	prjSt store.ProjectStore
}

// Creates a new instance of ProjectService.
// It takes a logger, a ProjectStore for project-related data operations.
//
// Parameters:
//   - lg: logger that implements the logger.Logger interface
//   - prjSt: project store that implements the ProjectStore interface
//
// Returns:
//   - a pointer to a new ProjectService
func NewProjectService(lg logger.Logger, prjSt store.ProjectStore) *ProjectService {
	return &ProjectService{
		lg:    lg,
		prjSt: prjSt,
	}
}

// Creates a new project and assigns current signed-in user
// to the project
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - request: a pointer to ProjectCreateRequest DTO.
//
// Returns:
//   - errdefs.ErrAlreadyExists if project already exists in database.
//   - errdefs.ErrProjectInternalError if internal errors occurs.
func (prjSrv *ProjectService) CreateProject(ctx context.Context, request *dto.ProjectCreateRequest) error {

	cid := middleware.GetCorrelationID(ctx)

	signedInUser, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		prjSrv.lg.Warn("getting signed in user from context failed",
			"event", logger.EventAuthInternalError,
			"correlation_id", cid,
			"err", err)
		return errdefs.ErrProjectInternalError
	}

	_, err = prjSrv.prjSt.CreateOne(ctx, &params.ProjectCreateParams{
		Name:    request.Name,
		Status:  request.Status,
		Members: append(request.Members, signedInUser.Id)})

	if err != nil {
		if errors.Is(err, errdefs.ErrDbUniqueViolation) {
			prjSrv.lg.Warn("project with the same name exists",
				"event", logger.EventProjectAlreadyExists,
				"correlation_id", cid,
				"project_name", request.Name,
				"err", err)
			return errdefs.ErrAlreadyExists
		}

		prjSrv.lg.Warn("failed to create project",
			"event", logger.EventProjectCreateFailed,
			"correlation_id", cid,
			"project_name", request.Name,
			"err", err)
		return errdefs.ErrProjectInternalError
	}

	prjSrv.lg.Info("created project successfully",
		"event", logger.EventProjectCreated,
		"correlation_id", cid,
		"project_name", request.Name)

	return nil
}
