package service

import (
	"context"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/dto/common"
	dto "mizu/internal/dto/project"
	"mizu/internal/event"
	"mizu/internal/logger"
	"mizu/internal/middleware"
)

// Handles project related operations.
type ProjectService struct {
	lg      logger.Logger
	prjSt   store.ProjectStore
	invSt   store.InvoiceStore
	contSt  store.ContractStore
	chReqSt store.ChangeRequestStore
	fileSt  store.FileStore
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
func NewProjectService(
	lg logger.Logger,
	prjSt store.ProjectStore,
	invSt store.InvoiceStore,
	contSt store.ContractStore,
	chReqSt store.ChangeRequestStore,
	fileSt store.FileStore,
) *ProjectService {

	return &ProjectService{
		lg:      lg,
		prjSt:   prjSt,
		invSt:   invSt,
		contSt:  contSt,
		chReqSt: chReqSt,
		fileSt:  fileSt,
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
//   - ErrAlreadyExists: if project already exists in database.
//   - ErrInternalError: if internal errors occur.
func (prjSrv *ProjectService) CreateProject(ctx context.Context, request *dto.ProjectCreateRequest) error {

	cid := middleware.GetCorrelationID(ctx)

	signedInUser, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		prjSrv.lg.Warn("getting signed in user from context failed",
			"event", event.EventInternalError,
			"correlation_id", cid,
			"scope", "project_service",
			"err", err)
		return ErrInternalError
	}

	_, err = prjSrv.prjSt.CreateOne(ctx, &params.ProjectCreate{
		Name:    request.Name,
		Status:  request.Status,
		Members: append(request.Members, signedInUser.Id)})

	if err != nil {
		if errors.Is(err, store.ErrUniqueViolation) {
			prjSrv.lg.Warn("project with the same name exists",
				"event", event.EventAlreadyExists,
				"correlation_id", cid,
				"project_name", request.Name,
				"scope", "project_service",
				"err", err)
			return ErrAlreadyExists
		}

		prjSrv.lg.Warn("failed to create project",
			"event", event.EventCreateFailed,
			"correlation_id", cid,
			"project_name", request.Name,
			"scope", "project_service",
			"err", err)
		return ErrInternalError
	}

	prjSrv.lg.Info("created project successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", cid,
		"project_name", request.Name,
		"scope", "project_service")

	return nil
}

// Creates a new project task with assigned users.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - request: a pointer to TaskCreateRequest DTO.
//
// Returns:
//   - ErrAlreadyExists: if task already exists in database.
//   - ErrBadRequest: if request parameter violates constraints (eg:- a task for a project that doesn't exist).
//   - ErrInternalError: if internal errors occur.
func (prjSrv *ProjectService) CreateTask(ctx context.Context, projectId int64, request *dto.TaskCreateRequest) error {

	cid := middleware.GetCorrelationID(ctx)

	_, err := prjSrv.prjSt.CreateTask(ctx, &params.TaskCreate{
		ProjectId:            projectId,
		Priority:             request.Priority,
		Status:               request.Status,
		Name:                 request.Name,
		Description:          request.Description,
		EstimatedTimeMinutes: request.EstimatedTimeMinutes,
		Assignees:            request.Assignees})

	if err != nil {
		if errors.Is(err, store.ErrUniqueViolation) {
			prjSrv.lg.Warn("task with the same name exists in the project",
				"event", event.EventAlreadyExists,
				"correlation_id", cid,
				"task_name", request.Name,
				"scope", "project_service",
				"err", err)
			return ErrAlreadyExists
		}

		if errors.Is(err, store.ErrForeignKeyViolation) {
			prjSrv.lg.Warn("task foreign key constraint violated",
				"event", event.EventCreateFailed,
				"correlation_id", cid,
				"task_name", request.Name,
				"scope", "project_service",
				"err", err)
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			prjSrv.lg.Warn("task not-null constraint violated",
				"event", event.EventCreateFailed,
				"correlation_id", cid,
				"task_name", request.Name,
				"scope", "project_service",
				"err", err)
			return ErrBadRequest
		}

		prjSrv.lg.Warn("failed to create project task",
			"event", event.EventCreateFailed,
			"correlation_id", cid,
			"task_name", request.Name,
			"scope", "project_service",
			"err", err)
		return ErrInternalError
	}

	prjSrv.lg.Info("created task successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", cid,
		"task_name", request.Name,
		"scope", "project_service")

	return nil
}

// Gets a paginated list of available projects assigned to
// the current user that meets the specified search query.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - query: a pointer to ProjectSearchQuery DTO.
//
// Returns:
//   - ErrInternalError: if internal errors occur.
func (prjSrv *ProjectService) GetProjects(ctx context.Context,
	query *dto.ProjectSearchQuery) (*common.Page[[]dto.ProjectsStatsResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	user, err := middleware.GetUserFromContext(ctx)

	if err != nil {
		prjSrv.lg.Error("could not get user from context",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	projects, err := prjSrv.prjSt.GetWithStats(ctx, &params.ProjectsSearch{
		Keyword: query.Keyword,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
		UserId:  user.Id,
	})

	if err != nil {
		prjSrv.lg.Error("could not get projects list",
			"event", event.EventGetFailed,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", user.Id,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	projectResponses := []dto.ProjectsStatsResponse{}
	for _, response := range projects.Items {
		stat := dto.ProjectsStatsResponse{
			Id:             response.Id,
			Name:           response.Name,
			Status:         response.Status,
			CreatedAt:      response.CreatedAt,
			TotalTasks:     response.TotalTasks,
			TasksCompleted: response.TasksCompleted,
			TotalInvoices:  response.TotalInvoices,
			InvoicesPaid:   response.InvoicesPaid,
			TotalQuotes:    response.TotalQuotes,
		}

		projectResponses = append(projectResponses, stat)
	}

	response := &common.Page[[]dto.ProjectsStatsResponse]{
		Count: projects.Total,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  projectResponses,
	}

	return response, nil
}

func (prjSrv *ProjectService) GetTasksByProject(
	ctx context.Context,
	projectId int64,
	query *dto.TaskSearchQuery) (*common.Page[[]dto.TaskResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)

	if err != nil {
		prjSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	taskSearch := params.TaskSearch{
		Keyword:  query.Keyword,
		Status:   query.Status,
		Priority: query.Priority,
		Offset:   (query.Page - 1) * query.Limit,
		Limit:    query.Limit,
	}

	tasks, err := prjSrv.prjSt.GetTasksByProjectId(ctx, projectId, &taskSearch)
	if err != nil {
		prjSrv.lg.Error("could not get tasks list",
			"event", event.EventGetFailed,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	taskMap := make(map[int64]*dto.TaskResponse)
	orderedTaskIds := []int64{}

	for _, task := range tasks {

		taskResp := dto.TaskResponse{
			Id:             task.Id,
			ProjectId:      task.ProjectId,
			Name:           task.Name,
			Status:         task.Status,
			Priority:       task.Priority,
			Description:    task.Description,
			CreatedAt:      task.CreatedAt,
			EstTimeMinutes: task.EstTimeMinutes,
		}

		assigneeResp := dto.TaskAssigneeResponse{
			Id:        *task.UserId,
			FirstName: *task.FirstName,
			LastName:  *task.LastName,
			Image:     task.Image,
			Title:     task.Title,
		}

		if _, ok := taskMap[task.Id]; !ok {

			orderedTaskIds = append(orderedTaskIds, task.Id)

			if task.UserId != nil {
				taskResp.Assignees = []dto.TaskAssigneeResponse{assigneeResp}
			} else {
				taskResp.Assignees = make([]dto.TaskAssigneeResponse, 0)
			}

			taskMap[task.Id] = &taskResp

		} else {
			taskMap[task.Id].Assignees = append(taskMap[task.Id].Assignees, assigneeResp)
		}
	}

	count, err := prjSrv.prjSt.CountTasksByProjectId(ctx, projectId, &taskSearch)
	if err != nil {
		prjSrv.lg.Error("could not get tasks count",
			"event", event.EventGetFailed,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	resp := common.Page[[]dto.TaskResponse]{
		Count: count,
		Page:  query.Page,
		Limit: query.Limit,
		Data:  make([]dto.TaskResponse, len(orderedTaskIds)),
	}

	for i, taskId := range orderedTaskIds {
		resp.Data[i] = *taskMap[taskId]
	}

	return &resp, nil
}

func (prjSrv *ProjectService) GetOneById(
	ctx context.Context,
	projectId int64) (*dto.ProjectDetailsResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)

	if err != nil {
		prjSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"err", err)
		return nil, ErrInternalError
	}

	project, err := prjSrv.prjSt.GetById(ctx, projectId)
	if err != nil {
		prjSrv.lg.Error("could not get project by id",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := dto.ProjectDetailsResponse{
		Id:        project.Id,
		Name:      project.Name,
		CreatedAt: project.CreatedAt,
		Status:    project.Status,
	}

	taskCount, err := prjSrv.prjSt.CountTasksByProjectId(ctx, projectId, &params.TaskSearch{})
	if err != nil {
		prjSrv.lg.Error("could not get project tasks count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.TaskCount = taskCount

	taskCompletedCount, err := prjSrv.prjSt.CountTasksByProjectId(ctx, projectId, &params.TaskSearch{
		Status: "completed",
	})
	if err != nil {
		prjSrv.lg.Error("could not get completed project tasks count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.TaskCompletedCount = taskCompletedCount

	invoiceCount, err := prjSrv.invSt.CountSummaryByProjectId(ctx, projectId, &params.InvoiceSearch{
		Type: params.InvoiceTypeInvoice,
	})
	if err != nil {
		prjSrv.lg.Error("could not get invoices count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.InvoiceCount = invoiceCount

	invoicePaidCount, err := prjSrv.invSt.CountSummaryByProjectId(ctx, projectId, &params.InvoiceSearch{
		Type:   params.InvoiceTypeInvoice,
		Status: params.InvoiceStatusPaid,
	})
	if err != nil {
		prjSrv.lg.Error("could not get paid invoices count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.InvoicePaidCount = invoicePaidCount

	quoteCount, err := prjSrv.invSt.CountSummaryByProjectId(ctx, projectId, &params.InvoiceSearch{
		Type: params.InvoiceTypeQuote,
	})
	if err != nil {
		prjSrv.lg.Error("could not get invoices count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.QuoteCount = quoteCount

	contractCount, err := prjSrv.contSt.CountByProjectId(ctx, projectId, &params.ContractSearch{})
	if err != nil {
		prjSrv.lg.Error("could not get contract count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.ContractCount = contractCount

	contractSignedCount, err := prjSrv.contSt.CountByProjectId(ctx, projectId, &params.ContractSearch{
		Status: params.ContractStatusSigned,
	})
	if err != nil {
		prjSrv.lg.Error("could not get contract count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.ContractSignedCount = contractSignedCount

	chReqCount, err := prjSrv.chReqSt.CountByProjectId(ctx, projectId, &params.ChangeRequestSearch{})
	if err != nil {
		prjSrv.lg.Error("could not get change request count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.ChangeReqCount = chReqCount

	chReqClosedCount, err := prjSrv.chReqSt.CountByProjectId(ctx, projectId, &params.ChangeRequestSearch{
		Status: params.ChangeRequestClosed,
	})
	if err != nil {
		prjSrv.lg.Error("could not get change request closed count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.ChangeReqClosedCount = chReqClosedCount

	fileCount, err := prjSrv.fileSt.CountByProjectId(ctx, projectId, &params.FileSearch{})
	if err != nil {
		prjSrv.lg.Error("could not get file count",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.FileCount = fileCount

	members, err := prjSrv.prjSt.GetMembersByProjectId(ctx, projectId)
	if err != nil {
		prjSrv.lg.Error("could not get project members",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"user_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp.Members = make([]dto.ProjectMemberResponse, len(members))

	for i, v := range members {
		resp.Members[i] = dto.ProjectMemberResponse{
			Id:        v.Id,
			FirstName: v.FirstName,
			LastName:  v.LastName,
			Role:      v.Role,
			Title:     v.Title,
			Image:     v.Image,
		}
	}

	return &resp, nil
}

func (prjSrv *ProjectService) GetTaskCompleteCountByProjectId(
	ctx context.Context,
	projectId int64) ([]dto.TaskMetricResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)

	if err != nil {
		prjSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"project_id", projectId,
			"err", err)
		return nil, ErrInternalError
	}

	metrics, err := prjSrv.prjSt.GetTaskMetricsByProjectId(
		ctx,
		projectId,
		params.TaskStatusCompleted)

	if err != nil {
		prjSrv.lg.Error("could not get completed task metrics",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"project_id", projectId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := make([]dto.TaskMetricResponse, len(metrics))

	for i, v := range metrics {
		resp[i] = dto.TaskMetricResponse{Key: v.Key, Value: v.Value}
	}

	return resp, nil
}

func (prjSrv *ProjectService) GetCreatedCount(ctx context.Context) ([]dto.ProjectMetricResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)

	if err != nil {
		prjSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"err", err)
		return nil, ErrInternalError
	}

	metrics, err := prjSrv.prjSt.GetCreatedMetricsByUserId(ctx, actor.Id)
	if err != nil {
		prjSrv.lg.Error("could not get project created metrics",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	resp := make([]dto.ProjectMetricResponse, len(metrics))

	for i, v := range metrics {
		resp[i] = dto.ProjectMetricResponse{Key: v.Key, Value: v.Value}
	}

	return resp, nil
}

func (prjSrv *ProjectService) SetStatusStarted(ctx context.Context, projectId int64) error {
	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)

	if err != nil {
		prjSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"err", err)
		return ErrInternalError
	}

	project, err := prjSrv.prjSt.GetById(ctx, projectId)
	if err != nil {
		prjSrv.lg.Error("could not get project by id ",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"project_id", projectId,
			"err", err)
		return ErrInternalError
	}

	if project.Status == params.ProjectStatusStarted {
		return ErrAlreadyExists
	}

	err = prjSrv.prjSt.SetStatusById(ctx, projectId, params.ProjectStatusStarted)
	if err != nil {
		prjSrv.lg.Error("could not set project as started",
			"event", event.EventInternalError,
			"scope", "project_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"project_id", projectId,
			"err", err)
		return ErrInternalError
	}

	return nil
}
