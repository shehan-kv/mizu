package task

import (
	"context"
	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/application/shared"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"mizu/internal/domain/task"
	"time"
)

type Service struct {
	userRepo    iam.UserRepository
	taskRepo    task.Repository
	projectRepo project.Repository

	taskSrv  *task.Service
	authzSrv *authz.Service

	internalBus eventbus.InternalBus

	idGen common.IDGenerator

	logger logger.Logger
}

func NewService(
	userRepo iam.UserRepository,
	taskRepo task.Repository,
	projectRepo project.Repository,
	taskSrv *task.Service,
	authzSrv *authz.Service,
	internalBus eventbus.InternalBus,
	idGen common.IDGenerator,
	logger logger.Logger,
) *Service {
	return &Service{
		userRepo:    userRepo,
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		taskSrv:     taskSrv,
		authzSrv:    authzSrv,
		internalBus: internalBus,
		idGen:       idGen,
		logger:      logger,
	}
}

func (s *Service) Create(ctx context.Context, params CreateTaskParams) error {
	now := time.Now()

	pID, err := project.NewProjectID(params.ProjectID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	priority, err := task.NewPriority(params.Priority)
	if err != nil {
		return err
	}

	status, err := task.NewStatus(params.Status)
	if err != nil {
		return err
	}

	name, err := task.NewName(params.Name)
	if err != nil {
		return err
	}

	estMinutes, err := task.NewMinutes(params.EstimatedMinutes)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	if !p.AcceptsTasks() {
		return project.ErrProjectNotAcceptingTasks
	}

	assignees := make([]iam.UserID, len(params.AssigneeIDs))
	for i := range params.AssigneeIDs {
		id, err := iam.NewUserID(params.AssigneeIDs[i])
		if err != nil {
			return err
		}

		assignees[i] = id
	}

	if err := s.taskSrv.EnsureValidAssignees(ctx, assignees, p); err != nil {
		return err
	}

	id, err := s.idGen.Generate()
	if err != nil {
		return err
	}

	taskID, err := task.NewTaskID(id)
	if err != nil {
		return err
	}

	t, err := task.NewTask(
		taskID,
		pID,
		priority,
		status,
		name,
		params.Description,
		estMinutes,
		assignees,
		now,
	)
	if err != nil {
		return err
	}

	return s.taskRepo.Add(ctx, t)
}

func (s *Service) MoveTaskToBacklog(ctx context.Context, taskID string, actorID string) error {

	now := time.Now()

	tID, err := task.NewTaskID(taskID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	t, err := s.taskRepo.Get(ctx, tID)
	if err != nil {
		return err
	}

	if !t.HasAssignee(actor) {
		return task.ErrTaskNotAssignedToUser
	}

	if err := t.MoveToBacklog(now); err != nil {
		return err
	}

	if err := s.taskRepo.Save(ctx, t); err != nil {
		return err
	}

	s.publishEvents(ctx, t.PullEvents())

	return nil
}

func (s *Service) MoveTaskToInProgress(ctx context.Context, taskID string, actorID string) error {

	now := time.Now()

	tID, err := task.NewTaskID(taskID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	t, err := s.taskRepo.Get(ctx, tID)
	if err != nil {
		return err
	}

	if !t.HasAssignee(actor) {
		return task.ErrTaskNotAssignedToUser
	}

	if err := t.MoveToInProgress(now); err != nil {
		return err
	}

	if err := s.taskRepo.Save(ctx, t); err != nil {
		return err
	}

	s.publishEvents(ctx, t.PullEvents())

	return nil
}

func (s *Service) MoveTaskToCompleted(ctx context.Context, taskID string, actorID string) error {

	now := time.Now()

	tID, err := task.NewTaskID(taskID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	t, err := s.taskRepo.Get(ctx, tID)
	if err != nil {
		return err
	}

	if !t.HasAssignee(actor) {
		return task.ErrTaskNotAssignedToUser
	}

	if err := t.MoveToCompleted(now); err != nil {
		return err
	}

	if err := s.taskRepo.Save(ctx, t); err != nil {
		return err
	}

	s.publishEvents(ctx, t.PullEvents())

	return nil
}

func (s *Service) ListTasks(ctx context.Context, params ListTaskParams) (*shared.Collection[TaskDTO], error) {

	pID, err := project.NewProjectID(params.ProjectID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(params.ActorID)
	if err != nil {
		return nil, err
	}

	filter := task.TaskFilter{
		ProjectID: pID,
		Keyword:   params.Keyword,
	}

	if params.Status != nil {
		status, err := task.NewStatus(*params.Status)
		if err != nil {
			return nil, err
		}
		filter.Status = &status
	}

	if params.Priority != nil {
		priority, err := task.NewPriority(*params.Priority)
		if err != nil {
			return nil, err
		}
		filter.Priority = &priority
	}

	page, err := common.NewPage(params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return nil, err
	}

	if !p.HasMember(actor) {
		return nil, project.ErrNotProjectMember
	}

	t, err := s.taskRepo.List(ctx, filter, page)
	if err != nil {
		return nil, err
	}

	count, err := s.taskRepo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	assigneeSet := make(map[iam.UserID]struct{})

	for i := range t {
		for _, assigneeID := range t[i].Assignees() {
			assigneeSet[assigneeID] = struct{}{}
		}
	}

	assigneeIDs := make([]iam.UserID, 0, len(assigneeSet))

	for id := range assigneeSet {
		assigneeIDs = append(assigneeIDs, id)
	}

	users, err := s.userRepo.ListByIDs(ctx, assigneeIDs, iam.UserFilter{})
	if err != nil {
		return nil, err
	}

	assigneeDTOMap := make(map[iam.UserID]AssigneeDTO, len(users))

	for i := range users {
		assigneeDTOMap[users[i].ID()] = AssigneeDTO{
			ID:        users[i].ID().String(),
			FirstName: users[i].FirstName(),
			LastName:  users[i].LastName(),
			Title:     users[i].Title(),
			HasImage:  users[i].Image() != nil,
			Role:      users[i].Role().String(),
		}
	}

	dto := make([]TaskDTO, len(t))

	for i := range t {

		taskAssigneeIDs := t[i].Assignees()

		assignees := make([]AssigneeDTO, 0, len(taskAssigneeIDs))

		for _, assigneeID := range taskAssigneeIDs {

			assigneeDTO, ok := assigneeDTOMap[assigneeID]
			if !ok {
				return nil, task.ErrTaskAssigneeNotFound
			}

			assignees = append(assignees, assigneeDTO)
		}

		dto[i] = TaskDTO{
			ID:               t[i].ID().String(),
			ProjectID:        t[i].ProjectID().String(),
			Priority:         t[i].Priority().String(),
			Status:           t[i].Status().String(),
			Name:             t[i].Name().String(),
			Description:      t[i].Description(),
			EstimatedMinutes: t[i].EstimatedMinutes().Int(),
			Assignees:        assignees,
			CreatedAt:        t[i].CreatedAt(),
			UpdatedAt:        t[i].UpdatedAt(),
		}
	}

	return &shared.Collection[TaskDTO]{
		Items:      dto,
		TotalCount: count,
	}, nil
}

func (s *Service) GetCompleteCount(ctx context.Context, projectID string, actorID string) ([]MetricDTO, error) {

	pID, err := project.NewProjectID(projectID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	p, err := s.projectRepo.Get(ctx, pID)
	if err != nil {
		return nil, err
	}

	if !p.HasMember(actor) {
		return nil, project.ErrNotProjectMember
	}

	m, err := s.taskRepo.ListCompletedPerDay(ctx, pID)
	if err != nil {
		return nil, err
	}

	dto := make([]MetricDTO, len(m))
	for i := range m {
		dto[i] = MetricDTO{
			Key:   m[i].Key(),
			Value: m[i].Value(),
		}
	}

	return dto, nil
}

func (s *Service) ListAssignees(ctx context.Context, taskID string, actorID string) ([]AssigneeDTO, error) {

	tID, err := task.NewTaskID(taskID)
	if err != nil {
		return nil, err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return nil, err
	}

	t, err := s.taskRepo.Get(ctx, tID)
	if err != nil {
		return nil, err
	}

	p, err := s.projectRepo.Get(ctx, t.ProjectID())
	if err != nil {
		return nil, err
	}

	if !p.HasMember(actor) {
		return nil, project.ErrNotProjectMember
	}

	users, err := s.userRepo.ListByIDs(ctx, t.Assignees(), iam.UserFilter{})
	if err != nil {
		return nil, err
	}

	dto := make([]AssigneeDTO, len(users))
	for i := range users {
		dto[i] = AssigneeDTO{
			ID:        users[i].ID().String(),
			FirstName: users[i].FirstName(),
			LastName:  users[i].LastName(),
			Title:     users[i].Title(),
			HasImage:  users[i].Image() != nil,
			Role:      users[i].Role().String(),
		}
	}

	return dto, nil
}

func (s *Service) ReplaceAssignees(ctx context.Context, taskID string, assigneeIDs []string, actorID string) error {

	now := time.Now()

	tID, err := task.NewTaskID(taskID)
	if err != nil {
		return err
	}

	actor, err := iam.NewUserID(actorID)
	if err != nil {
		return err
	}

	assignees := make([]iam.UserID, len(assigneeIDs))
	for i := range assigneeIDs {
		id, err := iam.NewUserID(assigneeIDs[i])
		if err != nil {
			return err
		}

		assignees[i] = id
	}

	if err := s.authzSrv.RequireAdministratorOrStaff(ctx, actor); err != nil {
		return err
	}

	t, err := s.taskRepo.Get(ctx, tID)
	if err != nil {
		return err
	}

	p, err := s.projectRepo.Get(ctx, t.ProjectID())
	if err != nil {
		return err
	}

	if !p.HasMember(actor) {
		return project.ErrNotProjectMember
	}

	if err := s.taskSrv.EnsureValidAssignees(ctx, assignees, p); err != nil {
		return err
	}

	t.ReplaceAssignees(assignees, now)

	return s.taskRepo.Save(ctx, t)
}

func (s *Service) publishEvents(ctx context.Context, events []common.Event) {
	for _, event := range events {
		if err := s.internalBus.Publish(ctx, event); err != nil {
			s.logger.Warn("failed to publish event",
				"event_type", event.EventType(),
				"err", err,
			)
		}
	}
}
