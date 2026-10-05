package task

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/domain/common"
	domainiam "mizu/internal/domain/iam"
	domainproject "mizu/internal/domain/project"
	domaintask "mizu/internal/domain/task"
)

type fakeUserRepository struct {
	getByIDUser     *domainiam.User
	getByIDErr      error
	getByIDCalls    int
	getByIDUserID   domainiam.UserID
	existsAll       bool
	existsAllErr    error
	existsAllCalls  int
	existsAllIDs    []domainiam.UserID
	listByIDsUsers  []*domainiam.User
	listByIDsErr    error
	listByIDsCalls  int
	listByIDsIDs    []domainiam.UserID
	listByIDsFilter domainiam.UserFilter
}

var _ domainiam.UserRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) Add(context.Context, *domainiam.User) error {
	return nil
}

func (f *fakeUserRepository) Exists(context.Context, domainiam.UserID) (bool, error) {
	return false, nil
}

func (f *fakeUserRepository) ExistsAll(
	_ context.Context,
	ids []domainiam.UserID,
) (bool, error) {
	f.existsAllCalls++
	f.existsAllIDs = append([]domainiam.UserID(nil), ids...)

	if f.existsAllErr != nil {
		return false, f.existsAllErr
	}

	return f.existsAll, nil
}

func (f *fakeUserRepository) GetByID(
	_ context.Context,
	id domainiam.UserID,
) (*domainiam.User, error) {
	f.getByIDCalls++
	f.getByIDUserID = id

	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}

	return f.getByIDUser, nil
}

func (f *fakeUserRepository) GetByEmail(
	context.Context,
	domainiam.Email,
) (*domainiam.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) List(
	context.Context,
	domainiam.UserFilter,
	common.Page,
) ([]*domainiam.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) ListByIDs(
	_ context.Context,
	ids []domainiam.UserID,
	filter domainiam.UserFilter,
) ([]*domainiam.User, error) {
	f.listByIDsCalls++
	f.listByIDsIDs = append([]domainiam.UserID(nil), ids...)
	f.listByIDsFilter = filter

	if f.listByIDsErr != nil {
		return nil, f.listByIDsErr
	}

	return f.listByIDsUsers, nil
}

func (f *fakeUserRepository) IsAnyAdministrator(
	context.Context,
	[]domainiam.UserID,
) (bool, error) {
	return false, nil
}

func (f *fakeUserRepository) HasAdministrator(context.Context) (bool, error) {
	return false, nil
}

func (f *fakeUserRepository) Count(
	context.Context,
	domainiam.UserFilter,
) (int, error) {
	return 0, nil
}

func (f *fakeUserRepository) Save(context.Context, *domainiam.User) error {
	return nil
}

func (f *fakeUserRepository) Remove(context.Context, *domainiam.User) error {
	return nil
}

type fakeTaskRepository struct {
	addErr   error
	addCalls int
	added    *domaintask.Task

	getErr   error
	getCalls int
	getID    domaintask.TaskID
	task     *domaintask.Task

	listErr    error
	listCalls  int
	list       []*domaintask.Task
	listFilter domaintask.TaskFilter
	listPage   common.Page

	countErr    error
	countCalls  int
	count       int
	countFilter domaintask.TaskFilter

	saveErr   error
	saveCalls int
	saved     *domaintask.Task

	listCompletedPerDayErr   error
	listCompletedPerDayCalls int
	listCompletedPerDayID    domainproject.ProjectID
	metrics                  []domaintask.Metric
}

var _ domaintask.Repository = (*fakeTaskRepository)(nil)

func (f *fakeTaskRepository) Add(
	_ context.Context,
	t *domaintask.Task,
) error {
	f.addCalls++
	f.added = t
	return f.addErr
}

func (f *fakeTaskRepository) Get(
	_ context.Context,
	id domaintask.TaskID,
) (*domaintask.Task, error) {
	f.getCalls++
	f.getID = id

	if f.getErr != nil {
		return nil, f.getErr
	}

	return f.task, nil
}

func (f *fakeTaskRepository) GetStatsByProject(
	context.Context,
	domainproject.ProjectID,
) (domaintask.Stats, error) {
	return domaintask.Stats{}, nil
}

func (f *fakeTaskRepository) List(
	_ context.Context,
	filter domaintask.TaskFilter,
	page common.Page,
) ([]*domaintask.Task, error) {
	f.listCalls++
	f.listFilter = filter
	f.listPage = page

	if f.listErr != nil {
		return nil, f.listErr
	}

	return f.list, nil
}

func (f *fakeTaskRepository) ListCompletedPerDay(
	_ context.Context,
	projectID domainproject.ProjectID,
) ([]domaintask.Metric, error) {
	f.listCompletedPerDayCalls++
	f.listCompletedPerDayID = projectID

	if f.listCompletedPerDayErr != nil {
		return nil, f.listCompletedPerDayErr
	}

	return f.metrics, nil
}

func (f *fakeTaskRepository) ListStatsByProjects(
	context.Context,
	[]domainproject.ProjectID,
) (map[domainproject.ProjectID]domaintask.Stats, error) {
	return nil, nil
}

func (f *fakeTaskRepository) Count(
	_ context.Context,
	filter domaintask.TaskFilter,
) (int, error) {
	f.countCalls++
	f.countFilter = filter

	if f.countErr != nil {
		return 0, f.countErr
	}

	return f.count, nil
}

func (f *fakeTaskRepository) Save(
	_ context.Context,
	t *domaintask.Task,
) error {
	f.saveCalls++
	f.saved = t
	return f.saveErr
}

func (f *fakeTaskRepository) Remove(
	context.Context,
	*domaintask.Task,
) error {
	return nil
}

type fakeProjectRepository struct {
	getErr   error
	getCalls int
	getID    domainproject.ProjectID
	project  *domainproject.Project

	saveAllErr   error
	saveAllCalls int
	savedAll     []*domainproject.Project
}

var _ domainproject.Repository = (*fakeProjectRepository)(nil)

func (f *fakeProjectRepository) Add(
	context.Context,
	*domainproject.Project,
) error {
	return nil
}

func (f *fakeProjectRepository) Get(
	_ context.Context,
	id domainproject.ProjectID,
) (*domainproject.Project, error) {
	f.getCalls++
	f.getID = id

	if f.getErr != nil {
		return nil, f.getErr
	}

	return f.project, nil
}

func (f *fakeProjectRepository) List(
	context.Context,
	domainproject.Filter,
	common.Page,
) ([]*domainproject.Project, error) {
	return nil, nil
}

func (f *fakeProjectRepository) ListByIDs(
	context.Context,
	[]domainproject.ProjectID,
) ([]*domainproject.Project, error) {
	return nil, nil
}

func (f *fakeProjectRepository) ListByMember(
	context.Context,
	domainiam.UserID,
) ([]*domainproject.Project, error) {
	return nil, nil
}

func (f *fakeProjectRepository) ListCreatedPerDay(
	context.Context,
	domainiam.UserID,
) ([]domainproject.Metric, error) {
	return nil, nil
}

func (f *fakeProjectRepository) Count(
	context.Context,
	domainproject.Filter,
) (int, error) {
	return 0, nil
}

func (f *fakeProjectRepository) Save(
	context.Context,
	*domainproject.Project,
) error {
	return nil
}

func (f *fakeProjectRepository) SaveAll(
	_ context.Context,
	projects []*domainproject.Project,
) error {
	f.saveAllCalls++
	f.savedAll = projects
	return f.saveAllErr
}

func (f *fakeProjectRepository) Remove(context.Context, *domainproject.Project) error {
	return nil
}

type fakeInternalBus struct {
	publishCalls int
	events       []common.Event
	publishErr   error
}

var _ eventbus.InternalBus = (*fakeInternalBus)(nil)

func (f *fakeInternalBus) Publish(_ context.Context, event common.Event) error {
	f.publishCalls++
	f.events = append(f.events, event)

	return f.publishErr
}

func (f *fakeInternalBus) Subscribe(common.EventType, eventbus.EventHandler) {
}

type fakeIDGenerator struct {
	id    string
	calls int
	err   error
}

var _ common.IDGenerator = (*fakeIDGenerator)(nil)

func (f *fakeIDGenerator) Generate() (string, error) {
	f.calls++

	if f.err != nil {
		return "", f.err
	}

	return f.id, nil
}

type fakeLogger struct {
	warnCalls int
	warnings  []string
}

var _ logger.Logger = (*fakeLogger)(nil)

func (f *fakeLogger) Info(string, ...any) {}

func (f *fakeLogger) Warn(msg string, _ ...any) {
	f.warnCalls++
	f.warnings = append(f.warnings, msg)
}

func (f *fakeLogger) Error(string, ...any) {}

func (f *fakeLogger) Fatal(string, ...any) {}

var (
	errRepository   = errors.New("repository failed")
	errIDGeneration = errors.New("id generation failed")
	errUserLookup   = errors.New("user lookup failed")
	errEventPublish = errors.New("event publish failed")
)

func newTestUserID(t *testing.T, id string) domainiam.UserID {
	t.Helper()

	userID, err := domainiam.NewUserID(id)
	if err != nil {
		t.Fatalf("failed to create user ID: %v", err)
	}

	return userID
}

func newTestProjectID(t *testing.T, id string) domainproject.ProjectID {
	t.Helper()

	projectID, err := domainproject.NewProjectID(id)
	if err != nil {
		t.Fatalf("failed to create project ID: %v", err)
	}

	return projectID
}

func newTestTaskID(t *testing.T, id string) domaintask.TaskID {
	t.Helper()

	taskID, err := domaintask.NewTaskID(id)
	if err != nil {
		t.Fatalf("failed to create task ID: %v", err)
	}

	return taskID
}

func newTestUser(
	t *testing.T,
	id string,
	email string,
	role domainiam.Role,
	isActive bool,
) *domainiam.User {
	t.Helper()

	userID := newTestUserID(t, id)

	name, err := domainiam.NewName("Test", "User")
	if err != nil {
		t.Fatalf("failed to create name: %v", err)
	}

	userEmail, err := domainiam.NewEmail(email)
	if err != nil {
		t.Fatalf("failed to create email: %v", err)
	}

	now := time.Now()

	return domainiam.NewUser(
		userID,
		name,
		userEmail,
		nil,
		role,
		isActive,
		userID,
		now,
	)
}

func newTestProject(
	t *testing.T,
	status domainproject.Status,
	members ...string,
) *domainproject.Project {
	t.Helper()

	id := newTestProjectID(t, "project-1")

	name, err := domainproject.NewName("Test Project")
	if err != nil {
		t.Fatalf("failed to create project name: %v", err)
	}

	memberIDs := make([]domainiam.UserID, 0, len(members))

	for _, id := range members {
		memberIDs = append(memberIDs, newTestUserID(t, id))
	}

	createdBy := newTestUserID(t, "admin-1")

	p, err := domainproject.NewProject(
		id,
		name,
		status,
		createdBy,
		memberIDs,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	return p
}

func newTestTask(
	t *testing.T,
	status domaintask.Status,
	assignees ...string,
) *domaintask.Task {
	t.Helper()

	taskID := newTestTaskID(t, "task-1")
	projectID := newTestProjectID(t, "project-1")

	priority, err := domaintask.NewPriority("medium")
	if err != nil {
		t.Fatalf("failed to create priority: %v", err)
	}

	name, err := domaintask.NewName("Test Task")
	if err != nil {
		t.Fatalf("failed to create task name: %v", err)
	}

	minutes, err := domaintask.NewMinutes(120)
	if err != nil {
		t.Fatalf("failed to create estimated minutes: %v", err)
	}

	assigneeIDs := make([]domainiam.UserID, 0, len(assignees))

	for _, id := range assignees {
		assigneeIDs = append(assigneeIDs, newTestUserID(t, id))
	}

	tm, err := domaintask.NewTask(
		taskID,
		projectID,
		priority,
		status,
		name,
		"Test description",
		minutes,
		assigneeIDs,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	return tm
}

func newTestService() (
	*Service,
	*fakeUserRepository,
	*fakeTaskRepository,
	*fakeProjectRepository,
	*fakeInternalBus,
	*fakeIDGenerator,
	*fakeLogger,
) {
	userRepo := &fakeUserRepository{}
	taskRepo := &fakeTaskRepository{}
	projectRepo := &fakeProjectRepository{}
	bus := &fakeInternalBus{}
	idGen := &fakeIDGenerator{id: "task-1"}
	logger := &fakeLogger{}

	taskSrv := domaintask.NewService(userRepo, projectRepo)
	authzSrv := authz.NewService(userRepo)

	s := NewService(
		userRepo,
		taskRepo,
		projectRepo,
		taskSrv,
		authzSrv,
		bus,
		idGen,
		logger,
	)

	return s, userRepo, taskRepo, projectRepo, bus, idGen, logger
}

func TestServiceCreate(t *testing.T) {
	t.Run("returns authorization error", func(t *testing.T) {
		s, userRepo, _, _, _, _, _ := newTestService()

		actor := newTestUser(
			t,
			"client-1",
			"client@example.com",
			domainiam.RoleClient,
			true,
		)
		userRepo.getByIDUser = actor

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "client-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			EstimatedMinutes: 60,
		})

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected authorization error, got %v", err)
		}
	})

	t.Run("returns project repository error", func(t *testing.T) {
		s, userRepo, _, projectRepo, _, _, _ := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.getByIDUser = actor

		projectRepo.getErr = errRepository

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "staff-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			EstimatedMinutes: 60,
		})

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when actor is not project member", func(t *testing.T) {
		s, userRepo, _, projectRepo, _, _, _ := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.getByIDUser = actor

		projectRepo.project = newTestProject(t, "started", "admin-1", "user-1")

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "staff-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			EstimatedMinutes: 60,
		})

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected project membership error, got %v", err)
		}
	})

	t.Run("returns error when project does not accept tasks", func(t *testing.T) {
		s, userRepo, _, projectRepo, _, _, _ := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.getByIDUser = actor

		projectRepo.project = newTestProject(
			t,
			domainproject.StatusCompleted,
			"admin-1",
			"staff-1",
		)

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "staff-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			EstimatedMinutes: 60,
		})

		if !errors.Is(err, domainproject.ErrProjectNotAcceptingTasks) {
			t.Fatalf("expected project-not-accepting error, got %v", err)
		}
	})

	t.Run("returns assignee validation error", func(t *testing.T) {
		s, userRepo, _, projectRepo, _, _, _ := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.getByIDUser = actor
		userRepo.existsAll = false

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
		)

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "staff-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			EstimatedMinutes: 60,
			AssigneeIDs:      []string{"missing-user"},
		})

		if !errors.Is(err, domainiam.ErrUserNotFound) {
			t.Fatalf("expected user-not-found error, got %v", err)
		}
	})

	t.Run("returns assignee repository error", func(t *testing.T) {
		s, userRepo, _, projectRepo, _, _, _ := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.getByIDUser = actor
		userRepo.existsAllErr = errRepository

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
		)

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "staff-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			EstimatedMinutes: 60,
			AssigneeIDs:      []string{"user-1"},
		})

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns ID generation error", func(t *testing.T) {
		s, userRepo, _, projectRepo, _, idGen, _ := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.getByIDUser = actor
		userRepo.existsAll = true

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
		)

		idGen.err = errIDGeneration

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "staff-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			EstimatedMinutes: 60,
		})

		if !errors.Is(err, errIDGeneration) {
			t.Fatalf("expected ID generation error, got %v", err)
		}
	})

	t.Run("returns task repository error", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.getByIDUser = actor
		userRepo.existsAll = true

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
		)

		taskRepo.addErr = errRepository

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "staff-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			Description:      "Description",
			EstimatedMinutes: 120,
		})

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("creates task", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, idGen, _ := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.getByIDUser = actor
		userRepo.existsAll = true

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
		)

		idGen.id = "task-42"

		err := s.Create(context.Background(), CreateTaskParams{
			ActorID:          "staff-1",
			ProjectID:        "project-1",
			Priority:         "medium",
			Status:           "backlog",
			Name:             "Test Task",
			Description:      "Description",
			EstimatedMinutes: 120,
			AssigneeIDs:      []string{"user-1"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if taskRepo.addCalls != 1 {
			t.Fatalf("expected one Add call, got %d", taskRepo.addCalls)
		}

		if taskRepo.added == nil {
			t.Fatal("expected task to be added")
		}

		if taskRepo.added.ID().String() != "task-42" {
			t.Fatalf("expected task ID task-42, got %s", taskRepo.added.ID())
		}

		if taskRepo.added.ProjectID().String() != "project-1" {
			t.Fatalf("expected project ID project-1, got %s", taskRepo.added.ProjectID())
		}
	})
}

func TestServiceMoveTaskToBacklog(t *testing.T) {
	t.Run("returns authorization error", func(t *testing.T) {
		s, userRepo, _, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"client-1",
			"client@example.com",
			domainiam.RoleClient,
			true,
		)

		err := s.MoveTaskToBacklog(
			context.Background(),
			"task-1",
			"client-1",
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected authorization error, got %v", err)
		}
	})

	t.Run("returns task repository error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.getErr = errRepository

		err := s.MoveTaskToBacklog(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when actor is not assigned", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusInProgress,
			"user-1",
		)

		err := s.MoveTaskToBacklog(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, domaintask.ErrTaskNotAssignedToUser) {
			t.Fatalf("expected not-assigned error, got %v", err)
		}
	})

	t.Run("returns transition error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)

		err := s.MoveTaskToBacklog(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if err == nil {
			t.Fatal("expected transition error")
		}

		if taskRepo.saveCalls != 0 {
			t.Fatalf("expected no Save call, got %d", taskRepo.saveCalls)
		}
	})

	t.Run("returns save error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusInProgress,
			"staff-1",
		)
		taskRepo.saveErr = errRepository

		err := s.MoveTaskToBacklog(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected save error, got %v", err)
		}
	})

	t.Run("moves task and publishes events", func(t *testing.T) {
		s, userRepo, taskRepo, _, bus, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusInProgress,
			"staff-1",
		)

		err := s.MoveTaskToBacklog(
			context.Background(),
			"task-1",
			"staff-1",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if taskRepo.saveCalls != 1 {
			t.Fatalf("expected one Save call, got %d", taskRepo.saveCalls)
		}

		if taskRepo.saved.Status() != domaintask.StatusBacklog {
			t.Fatalf("expected backlog status, got %s", taskRepo.saved.Status())
		}

		if bus.publishCalls == 0 {
			t.Fatal("expected event to be published")
		}
	})
}

func TestServiceMoveTaskToInProgress(t *testing.T) {
	t.Run("returns authorization error", func(t *testing.T) {
		s, userRepo, _, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"client-1",
			"client@example.com",
			domainiam.RoleClient,
			true,
		)

		err := s.MoveTaskToInProgress(
			context.Background(),
			"task-1",
			"client-1",
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected authorization error, got %v", err)
		}
	})

	t.Run("returns task repository error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.getErr = errRepository

		err := s.MoveTaskToInProgress(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when actor is not assigned", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"user-1",
		)

		err := s.MoveTaskToInProgress(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, domaintask.ErrTaskNotAssignedToUser) {
			t.Fatalf("expected not-assigned error, got %v", err)
		}
	})

	t.Run("returns transition error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusInProgress,
			"staff-1",
		)

		err := s.MoveTaskToInProgress(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if err == nil {
			t.Fatal("expected transition error")
		}
	})

	t.Run("returns save error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)
		taskRepo.saveErr = errRepository

		err := s.MoveTaskToInProgress(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected save error, got %v", err)
		}
	})

	t.Run("moves task and publishes events", func(t *testing.T) {
		s, userRepo, taskRepo, _, bus, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)

		err := s.MoveTaskToInProgress(
			context.Background(),
			"task-1",
			"staff-1",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if taskRepo.saveCalls != 1 {
			t.Fatalf("expected one Save call, got %d", taskRepo.saveCalls)
		}

		if taskRepo.saved.Status() != domaintask.StatusInProgress {
			t.Fatalf("expected in-progress status, got %s", taskRepo.saved.Status())
		}

		if bus.publishCalls == 0 {
			t.Fatal("expected event to be published")
		}
	})
}

func TestServiceMoveTaskToCompleted(t *testing.T) {
	t.Run("returns authorization error", func(t *testing.T) {
		s, userRepo, _, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"client-1",
			"client@example.com",
			domainiam.RoleClient,
			true,
		)

		err := s.MoveTaskToCompleted(
			context.Background(),
			"task-1",
			"client-1",
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected authorization error, got %v", err)
		}
	})

	t.Run("returns task repository error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.getErr = errRepository

		err := s.MoveTaskToCompleted(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when actor is not assigned", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusInProgress,
			"user-1",
		)

		err := s.MoveTaskToCompleted(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, domaintask.ErrTaskNotAssignedToUser) {
			t.Fatalf("expected not-assigned error, got %v", err)
		}
	})

	t.Run("returns transition error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusCompleted,
			"staff-1",
		)

		err := s.MoveTaskToCompleted(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if err == nil {
			t.Fatal("expected transition error")
		}
	})

	t.Run("returns save error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusInProgress,
			"staff-1",
		)
		taskRepo.saveErr = errRepository

		err := s.MoveTaskToCompleted(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected save error, got %v", err)
		}
	})

	t.Run("moves task and publishes events", func(t *testing.T) {
		s, userRepo, taskRepo, _, bus, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.task = newTestTask(
			t,
			domaintask.StatusInProgress,
			"staff-1",
		)

		err := s.MoveTaskToCompleted(
			context.Background(),
			"task-1",
			"staff-1",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if taskRepo.saveCalls != 1 {
			t.Fatalf("expected one Save call, got %d", taskRepo.saveCalls)
		}

		if taskRepo.saved.Status() != domaintask.StatusCompleted {
			t.Fatalf("expected completed status, got %s", taskRepo.saved.Status())
		}

		if bus.publishCalls == 0 {
			t.Fatal("expected event to be published")
		}
	})
}

func TestServiceListTasks(t *testing.T) {
	t.Run("returns project repository error", func(t *testing.T) {
		s, _, _, projectRepo, _, _, _ := newTestService()

		projectRepo.getErr = errRepository

		_, err := s.ListTasks(context.Background(), ListTaskParams{
			ActorID:   "staff-1",
			ProjectID: "project-1",
			Limit:     20,
			Offset:    0,
		})

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when actor is not project member", func(t *testing.T) {
		s, _, _, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"user-1",
		)

		_, err := s.ListTasks(context.Background(), ListTaskParams{
			ActorID:   "staff-1",
			ProjectID: "project-1",
			Limit:     20,
			Offset:    0,
		})

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected membership error, got %v", err)
		}
	})

	t.Run("returns task list repository error", func(t *testing.T) {
		s, _, taskRepo, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
		)
		taskRepo.listErr = errRepository

		_, err := s.ListTasks(context.Background(), ListTaskParams{
			ActorID:   "staff-1",
			ProjectID: "project-1",
			Limit:     20,
			Offset:    0,
		})

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns task count repository error", func(t *testing.T) {
		s, _, taskRepo, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
		)
		taskRepo.list = []*domaintask.Task{}
		taskRepo.countErr = errRepository

		_, err := s.ListTasks(context.Background(), ListTaskParams{
			ActorID:   "staff-1",
			ProjectID: "project-1",
			Limit:     20,
			Offset:    0,
		})

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns user repository error", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
		)

		taskRepo.list = []*domaintask.Task{
			newTestTask(t, domaintask.StatusBacklog, "user-1"),
		}
		taskRepo.count = 1

		userRepo.listByIDsErr = errRepository

		_, err := s.ListTasks(context.Background(), ListTaskParams{
			ActorID:   "staff-1",
			ProjectID: "project-1",
			Limit:     20,
			Offset:    0,
		})

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when an assignee is missing from user result", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
		)

		taskRepo.list = []*domaintask.Task{
			newTestTask(t, domaintask.StatusBacklog, "user-1"),
		}
		taskRepo.count = 1
		userRepo.listByIDsUsers = nil

		_, err := s.ListTasks(context.Background(), ListTaskParams{
			ActorID:   "staff-1",
			ProjectID: "project-1",
			Limit:     20,
			Offset:    0,
		})

		if !errors.Is(err, domaintask.ErrTaskAssigneeNotFound) {
			t.Fatalf("expected assignee-not-found error, got %v", err)
		}
	})

	t.Run("lists tasks and maps DTOs", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
		)

		taskRepo.list = []*domaintask.Task{
			newTestTask(t, domaintask.StatusBacklog, "user-1"),
		}
		taskRepo.count = 7

		userRepo.listByIDsUsers = []*domainiam.User{
			newTestUser(
				t,
				"user-1",
				"user@example.com",
				domainiam.RoleStaff,
				true,
			),
		}

		keyword := "authentication"
		status := domaintask.StatusBacklog.String()
		priority := "medium"

		result, err := s.ListTasks(context.Background(), ListTaskParams{
			ActorID:   "staff-1",
			ProjectID: "project-1",
			Keyword:   &keyword,
			Status:    &status,
			Priority:  &priority,
			Limit:     20,
			Offset:    0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected collection")
		}

		if result.TotalCount != 7 {
			t.Fatalf("expected total count 7, got %d", result.TotalCount)
		}

		if len(result.Items) != 1 {
			t.Fatalf("expected one task, got %d", len(result.Items))
		}

		dto := result.Items[0]

		if dto.ID != "task-1" {
			t.Fatalf("expected task-1, got %s", dto.ID)
		}

		if dto.ProjectID != "project-1" {
			t.Fatalf("expected project-1, got %s", dto.ProjectID)
		}

		if dto.Priority != "medium" {
			t.Fatalf("expected medium priority, got %s", dto.Priority)
		}

		if dto.Status != "backlog" {
			t.Fatalf("expected backlog status, got %s", dto.Status)
		}

		if dto.Name != "Test Task" {
			t.Fatalf("expected Test Task, got %s", dto.Name)
		}

		if dto.Description != "Test description" {
			t.Fatalf("unexpected description: %s", dto.Description)
		}

		if dto.EstimatedMinutes != 120 {
			t.Fatalf("expected 120 minutes, got %d", dto.EstimatedMinutes)
		}

		if len(dto.Assignees) != 1 {
			t.Fatalf("expected one assignee, got %d", len(dto.Assignees))
		}

		assignee := dto.Assignees[0]

		if assignee.ID != "user-1" {
			t.Fatalf("expected user-1, got %s", assignee.ID)
		}

		if assignee.FirstName != "Test" {
			t.Fatalf("expected Test, got %s", assignee.FirstName)
		}

		if assignee.LastName != "User" {
			t.Fatalf("expected User, got %s", assignee.LastName)
		}

		if assignee.Role != domainiam.RoleStaff.String() {
			t.Fatalf("unexpected role: %s", assignee.Role)
		}

		if taskRepo.listCalls != 1 {
			t.Fatalf("expected one List call, got %d", taskRepo.listCalls)
		}

		if taskRepo.countCalls != 1 {
			t.Fatalf("expected one Count call, got %d", taskRepo.countCalls)
		}
	})
}

func TestServiceGetCompleteCount(t *testing.T) {
	t.Run("returns project repository error", func(t *testing.T) {
		s, _, _, projectRepo, _, _, _ := newTestService()

		projectRepo.getErr = errRepository

		_, err := s.GetCompleteCount(
			context.Background(),
			"project-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when actor is not project member", func(t *testing.T) {
		s, _, _, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"user-1",
		)

		_, err := s.GetCompleteCount(
			context.Background(),
			"project-1",
			"staff-1",
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected membership error, got %v", err)
		}
	})

	t.Run("returns metric repository error", func(t *testing.T) {
		s, _, taskRepo, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
		)
		taskRepo.listCompletedPerDayErr = errRepository

		_, err := s.GetCompleteCount(
			context.Background(),
			"project-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns metrics", func(t *testing.T) {
		s, _, taskRepo, projectRepo, _, _, _ := newTestService()

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
		)

		key1 := time.Now().Format("2006-01-02")
		key2 := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

		metric1 := domaintask.NewMetric(key1, 3)
		metric2 := domaintask.NewMetric(key2, 7)

		taskRepo.metrics = []domaintask.Metric{
			metric1,
			metric2,
		}

		result, err := s.GetCompleteCount(
			context.Background(),
			"project-1",
			"staff-1",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []MetricDTO{
			{
				Key:   key1,
				Value: 3,
			},
			{
				Key:   key2,
				Value: 7,
			},
		}

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("unexpected result:\nwant: %#v\ngot:  %#v", expected, result)
		}
	})
}

func TestServiceListAssignees(t *testing.T) {
	t.Run("returns task repository error", func(t *testing.T) {
		s, _, taskRepo, _, _, _, _ := newTestService()

		taskRepo.getErr = errRepository

		_, err := s.ListAssignees(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns project repository error", func(t *testing.T) {
		s, _, taskRepo, projectRepo, _, _, _ := newTestService()

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"user-1",
		)
		projectRepo.getErr = errRepository

		_, err := s.ListAssignees(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when actor is not project member", func(t *testing.T) {
		s, _, taskRepo, projectRepo, _, _, _ := newTestService()

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"user-1",
		)

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"user-1",
		)

		_, err := s.ListAssignees(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected membership error, got %v", err)
		}
	})

	t.Run("returns user repository error", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"user-1",
		)

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
		)

		userRepo.listByIDsErr = errRepository

		_, err := s.ListAssignees(
			context.Background(),
			"task-1",
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns assignees", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"user-1",
		)

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
		)

		userRepo.listByIDsUsers = []*domainiam.User{
			newTestUser(
				t,
				"user-1",
				"user@example.com",
				domainiam.RoleStaff,
				true,
			),
		}

		result, err := s.ListAssignees(
			context.Background(),
			"task-1",
			"staff-1",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected one assignee, got %d", len(result))
		}

		if result[0].ID != "user-1" {
			t.Fatalf("expected user-1, got %s", result[0].ID)
		}

		if result[0].FirstName != "Test" {
			t.Fatalf("expected Test, got %s", result[0].FirstName)
		}

		if result[0].LastName != "User" {
			t.Fatalf("expected User, got %s", result[0].LastName)
		}

		if result[0].Role != domainiam.RoleStaff.String() {
			t.Fatalf("unexpected role: %s", result[0].Role)
		}
	})
}

func TestServiceReplaceAssignees(t *testing.T) {
	t.Run("returns authorization error", func(t *testing.T) {
		s, userRepo, _, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"client-1",
			"client@example.com",
			domainiam.RoleClient,
			true,
		)

		err := s.ReplaceAssignees(
			context.Background(),
			"task-1",
			[]string{"user-1"},
			"client-1",
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected authorization error, got %v", err)
		}
	})

	t.Run("returns task repository error", func(t *testing.T) {
		s, userRepo, taskRepo, _, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		taskRepo.getErr = errRepository

		err := s.ReplaceAssignees(
			context.Background(),
			"task-1",
			[]string{"user-1"},
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns project repository error", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)

		projectRepo.getErr = errRepository

		err := s.ReplaceAssignees(
			context.Background(),
			"task-1",
			[]string{"user-1"},
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns error when actor is not project member", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"user-1",
		)

		err := s.ReplaceAssignees(
			context.Background(),
			"task-1",
			[]string{"user-1"},
			"staff-1",
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected membership error, got %v", err)
		}
	})

	t.Run("returns assignee validation error", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.existsAll = false

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
		)

		err := s.ReplaceAssignees(
			context.Background(),
			"task-1",
			[]string{"missing-user"},
			"staff-1",
		)

		if !errors.Is(err, domainiam.ErrUserNotFound) {
			t.Fatalf("expected user-not-found error, got %v", err)
		}
	})

	t.Run("returns save error", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.existsAll = true

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
		)

		taskRepo.saveErr = errRepository

		err := s.ReplaceAssignees(
			context.Background(),
			"task-1",
			[]string{"user-1"},
			"staff-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected save error, got %v", err)
		}
	})

	t.Run("replaces assignees", func(t *testing.T) {
		s, userRepo, taskRepo, projectRepo, _, _, _ := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)
		userRepo.existsAll = true

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)

		projectRepo.project = newTestProject(
			t,
			"started",
			"admin-1",
			"staff-1",
			"user-1",
			"user-2",
		)

		err := s.ReplaceAssignees(
			context.Background(),
			"task-1",
			[]string{"user-1", "user-2"},
			"staff-1",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if taskRepo.saveCalls != 1 {
			t.Fatalf("expected one Save call, got %d", taskRepo.saveCalls)
		}

		assignees := taskRepo.saved.Assignees()

		expected := []domainiam.UserID{
			newTestUserID(t, "user-1"),
			newTestUserID(t, "user-2"),
		}

		if !reflect.DeepEqual(assignees, expected) {
			t.Fatalf(
				"unexpected assignees:\nwant: %#v\ngot:  %#v",
				expected,
				assignees,
			)
		}
	})
}

func TestServicePublishEvents(t *testing.T) {
	t.Run("logs warning when event publishing fails", func(t *testing.T) {
		s, userRepo, taskRepo, _, bus, _, logger := newTestService()

		userRepo.getByIDUser = newTestUser(
			t,
			"staff-1",
			"staff@example.com",
			domainiam.RoleStaff,
			true,
		)

		taskRepo.task = newTestTask(
			t,
			domaintask.StatusBacklog,
			"staff-1",
		)

		bus.publishErr = errEventPublish

		err := s.MoveTaskToInProgress(
			context.Background(),
			"task-1",
			"staff-1",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if bus.publishCalls == 0 {
			t.Fatal("expected Publish to be called")
		}

		if logger.warnCalls == 0 {
			t.Fatal("expected warning to be logged")
		}
	})
}
