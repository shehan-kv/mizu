package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/application/uow"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	domaincontract "mizu/internal/domain/contract"
	domainiam "mizu/internal/domain/iam"
	domainmessage "mizu/internal/domain/message"
	domainproject "mizu/internal/domain/project"
	domaintask "mizu/internal/domain/task"
)

var (
	errRepository = errors.New("repository failed")
	errUOW        = errors.New("unit of work failed")
	errIDGen      = errors.New("id generation failed")
	errEventBus   = errors.New("event bus failed")
)

type fakeUserRepository struct {
	existsFn    func(context.Context, domainiam.UserID) (bool, error)
	existsAllFn func(context.Context, []domainiam.UserID) (bool, error)
	getByIDFn   func(context.Context, domainiam.UserID) (*domainiam.User, error)
	listByIDsFn func(context.Context, []domainiam.UserID, domainiam.UserFilter) ([]*domainiam.User, error)
	isAdminFn   func(context.Context, []domainiam.UserID) (bool, error)

	existsCalls    int
	existsAllCalls int
	getByIDCalls   int
	listByIDsCalls int
	isAdminCalls   int

	existsResult    bool
	existsErr       error
	existsAllResult bool
	existsAllErr    error
}

var _ domainiam.UserRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) Add(context.Context, *domainiam.User) error {
	panic("unexpected call to UserRepository.Add")
}

func (f *fakeUserRepository) Exists(
	ctx context.Context,
	id domainiam.UserID,
) (bool, error) {
	f.existsCalls++

	if f.existsFn != nil {
		return f.existsFn(ctx, id)
	}

	return f.existsResult, f.existsErr
}

func (f *fakeUserRepository) ExistsAll(
	ctx context.Context,
	ids []domainiam.UserID,
) (bool, error) {
	f.existsAllCalls++

	if f.existsAllFn != nil {
		return f.existsAllFn(ctx, ids)
	}

	return f.existsAllResult, f.existsAllErr
}

func (f *fakeUserRepository) GetByID(
	ctx context.Context,
	id domainiam.UserID,
) (*domainiam.User, error) {
	f.getByIDCalls++

	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	return nil, nil
}

func (f *fakeUserRepository) GetByEmail(
	context.Context,
	domainiam.Email,
) (*domainiam.User, error) {
	panic("unexpected call to UserRepository.GetByEmail")
}

func (f *fakeUserRepository) List(
	context.Context,
	domainiam.UserFilter,
	common.Page,
) ([]*domainiam.User, error) {
	panic("unexpected call to UserRepository.List")
}

func (f *fakeUserRepository) ListByIDs(
	ctx context.Context,
	ids []domainiam.UserID,
	filter domainiam.UserFilter,
) ([]*domainiam.User, error) {
	f.listByIDsCalls++

	if f.listByIDsFn != nil {
		return f.listByIDsFn(ctx, ids, filter)
	}

	return nil, nil
}

func (f *fakeUserRepository) IsAnyAdministrator(
	ctx context.Context,
	ids []domainiam.UserID,
) (bool, error) {
	f.isAdminCalls++

	if f.isAdminFn != nil {
		return f.isAdminFn(ctx, ids)
	}

	return false, nil
}

func (f *fakeUserRepository) HasAdministrator(
	context.Context,
) (bool, error) {
	panic("unexpected call to UserRepository.HasAdministrator")
}

func (f *fakeUserRepository) Count(
	context.Context,
	domainiam.UserFilter,
) (int, error) {
	panic("unexpected call to UserRepository.Count")
}

func (f *fakeUserRepository) Save(
	context.Context,
	*domainiam.User,
) error {
	panic("unexpected call to UserRepository.Save")
}

func (f *fakeUserRepository) Remove(
	context.Context,
	*domainiam.User,
) error {
	panic("unexpected call to UserRepository.Remove")
}

type fakeProjectRepository struct {
	addFn          func(context.Context, *domainproject.Project) error
	getFn          func(context.Context, domainproject.ProjectID) (*domainproject.Project, error)
	listFn         func(context.Context, domainproject.Filter, common.Page) ([]*domainproject.Project, error)
	listByIDsFn    func(context.Context, []domainproject.ProjectID) ([]*domainproject.Project, error)
	listByMemberFn func(context.Context, domainiam.UserID) ([]*domainproject.Project, error)
	listCreatedFn  func(context.Context, domainiam.UserID) ([]domainproject.Metric, error)
	countFn        func(context.Context, domainproject.Filter) (int, error)
	saveFn         func(context.Context, *domainproject.Project) error
	saveAllFn      func(context.Context, []*domainproject.Project) error
	removeFn       func(context.Context, *domainproject.Project) error

	addCalls          int
	getCalls          int
	listCalls         int
	listByIDsCalls    int
	listByMemberCalls int
	listCreatedCalls  int
	countCalls        int
	saveCalls         int
	saveAllCalls      int
	removeCalls       int

	addedProject   *domainproject.Project
	savedProject   *domainproject.Project
	savedProjects  []*domainproject.Project
	removedProject *domainproject.Project

	project   *domainproject.Project
	projects  []*domainproject.Project
	listCount int
}

var _ domainproject.Repository = (*fakeProjectRepository)(nil)

func (f *fakeProjectRepository) Add(
	ctx context.Context,
	p *domainproject.Project,
) error {
	f.addCalls++
	f.addedProject = p

	if f.addFn != nil {
		return f.addFn(ctx, p)
	}

	return nil
}

func (f *fakeProjectRepository) Get(
	ctx context.Context,
	id domainproject.ProjectID,
) (*domainproject.Project, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	return f.project, nil
}

func (f *fakeProjectRepository) List(
	ctx context.Context,
	filter domainproject.Filter,
	page common.Page,
) ([]*domainproject.Project, error) {
	f.listCalls++

	if f.listFn != nil {
		return f.listFn(ctx, filter, page)
	}

	return f.projects, nil
}

func (f *fakeProjectRepository) ListByIDs(
	ctx context.Context,
	ids []domainproject.ProjectID,
) ([]*domainproject.Project, error) {
	f.listByIDsCalls++

	if f.listByIDsFn != nil {
		return f.listByIDsFn(ctx, ids)
	}

	return f.projects, nil
}

func (f *fakeProjectRepository) ListByMember(
	ctx context.Context,
	id domainiam.UserID,
) ([]*domainproject.Project, error) {
	f.listByMemberCalls++

	if f.listByMemberFn != nil {
		return f.listByMemberFn(ctx, id)
	}

	return f.projects, nil
}

func (f *fakeProjectRepository) ListCreatedPerDay(
	ctx context.Context,
	id domainiam.UserID,
) ([]domainproject.Metric, error) {
	f.listCreatedCalls++

	if f.listCreatedFn != nil {
		return f.listCreatedFn(ctx, id)
	}

	return nil, nil
}

func (f *fakeProjectRepository) Count(
	ctx context.Context,
	filter domainproject.Filter,
) (int, error) {
	f.countCalls++

	if f.countFn != nil {
		return f.countFn(ctx, filter)
	}

	return f.listCount, nil
}

func (f *fakeProjectRepository) Save(
	ctx context.Context,
	p *domainproject.Project,
) error {
	f.saveCalls++
	f.savedProject = p

	if f.saveFn != nil {
		return f.saveFn(ctx, p)
	}

	return nil
}

func (f *fakeProjectRepository) SaveAll(
	ctx context.Context,
	projects []*domainproject.Project,
) error {
	f.saveAllCalls++
	f.savedProjects = projects

	if f.saveAllFn != nil {
		return f.saveAllFn(ctx, projects)
	}

	return nil
}

func (f *fakeProjectRepository) Remove(
	ctx context.Context,
	p *domainproject.Project,
) error {
	f.removeCalls++
	f.removedProject = p

	if f.removeFn != nil {
		return f.removeFn(ctx, p)
	}

	return nil
}

type fakeTaskRepository struct {
	getStatsFn  func(context.Context, domainproject.ProjectID) (domaintask.Stats, error)
	listStatsFn func(context.Context, []domainproject.ProjectID) (map[domainproject.ProjectID]domaintask.Stats, error)

	getStatsCalls  int
	listStatsCalls int

	stats          domaintask.Stats
	statsByProject map[domainproject.ProjectID]domaintask.Stats
}

var _ domaintask.Repository = (*fakeTaskRepository)(nil)

func (f *fakeTaskRepository) Add(
	context.Context,
	*domaintask.Task,
) error {
	panic("unexpected call to TaskRepository.Add")
}

func (f *fakeTaskRepository) Get(
	context.Context,
	domaintask.TaskID,
) (*domaintask.Task, error) {
	panic("unexpected call to TaskRepository.Get")
}

func (f *fakeTaskRepository) GetStatsByProject(
	ctx context.Context,
	id domainproject.ProjectID,
) (domaintask.Stats, error) {
	f.getStatsCalls++

	if f.getStatsFn != nil {
		return f.getStatsFn(ctx, id)
	}

	return f.stats, nil
}

func (f *fakeTaskRepository) List(
	context.Context,
	domaintask.TaskFilter,
	common.Page,
) ([]*domaintask.Task, error) {
	panic("unexpected call to TaskRepository.List")
}

func (f *fakeTaskRepository) ListCompletedPerDay(
	context.Context,
	domainproject.ProjectID,
) ([]domaintask.Metric, error) {
	panic("unexpected call to TaskRepository.ListCompletedPerDay")
}

func (f *fakeTaskRepository) ListStatsByProjects(
	ctx context.Context,
	ids []domainproject.ProjectID,
) (map[domainproject.ProjectID]domaintask.Stats, error) {
	f.listStatsCalls++

	if f.listStatsFn != nil {
		return f.listStatsFn(ctx, ids)
	}

	return f.statsByProject, nil
}

func (f *fakeTaskRepository) Count(
	context.Context,
	domaintask.TaskFilter,
) (int, error) {
	panic("unexpected call to TaskRepository.Count")
}

func (f *fakeTaskRepository) Save(
	context.Context,
	*domaintask.Task,
) error {
	panic("unexpected call to TaskRepository.Save")
}

func (f *fakeTaskRepository) Remove(
	context.Context,
	*domaintask.Task,
) error {
	panic("unexpected call to TaskRepository.Remove")
}

type fakeBillingRepository struct {
	getStatsFn  func(context.Context, domainproject.ProjectID) (billing.Stats, error)
	listStatsFn func(context.Context, []domainproject.ProjectID) (map[domainproject.ProjectID]billing.Stats, error)

	getStatsCalls  int
	listStatsCalls int

	stats          billing.Stats
	statsByProject map[domainproject.ProjectID]billing.Stats
}

var _ billing.Repository = (*fakeBillingRepository)(nil)

func (f *fakeBillingRepository) Add(
	context.Context,
	*billing.Invoice,
) error {
	panic("unexpected call to BillingRepository.Add")
}

func (f *fakeBillingRepository) Get(
	context.Context,
	billing.InvoiceID,
) (*billing.Invoice, error) {
	panic("unexpected call to BillingRepository.Get")
}

func (f *fakeBillingRepository) GetCurrencyByCode(
	context.Context,
	billing.CurrencyCode,
) (billing.Currency, error) {
	panic("unexpected call to BillingRepository.GetCurrencyByCode")
}

func (f *fakeBillingRepository) GetBillingOverviewByMember(
	context.Context,
	domainiam.UserID,
) (billing.BillingOverview, error) {
	panic("unexpected call to BillingRepository.GetBillingOverviewByMember")
}

func (f *fakeBillingRepository) GetStatsByProject(
	ctx context.Context,
	id domainproject.ProjectID,
) (billing.Stats, error) {
	f.getStatsCalls++

	if f.getStatsFn != nil {
		return f.getStatsFn(ctx, id)
	}

	return f.stats, nil
}

func (f *fakeBillingRepository) ListByMember(
	context.Context,
	billing.FilterByMember,
	common.Page,
) ([]*billing.Invoice, error) {
	panic("unexpected call to BillingRepository.ListByMember")
}

func (f *fakeBillingRepository) ListByProject(
	context.Context,
	billing.FilterByProject,
	common.Page,
) ([]*billing.Invoice, error) {
	panic("unexpected call to BillingRepository.ListByProject")
}

func (f *fakeBillingRepository) ListMonthlyPaidCountByProject(
	context.Context,
	domainproject.ProjectID,
) ([]billing.Metric, error) {
	panic("unexpected call to BillingRepository.ListMonthlyPaidCountByProject")
}

func (f *fakeBillingRepository) ListMonthlyPaidCountByMember(
	context.Context,
	domainiam.UserID,
) ([]billing.Metric, error) {
	panic("unexpected call to BillingRepository.ListMonthlyPaidCountByMember")
}

func (f *fakeBillingRepository) ListStatsByProjects(
	ctx context.Context,
	ids []domainproject.ProjectID,
) (map[domainproject.ProjectID]billing.Stats, error) {
	f.listStatsCalls++

	if f.listStatsFn != nil {
		return f.listStatsFn(ctx, ids)
	}

	return f.statsByProject, nil
}

func (f *fakeBillingRepository) CountByProject(
	context.Context,
	billing.FilterByProject,
) (int, error) {
	panic("unexpected call to BillingRepository.CountByProject")
}

func (f *fakeBillingRepository) CountByMember(
	context.Context,
	billing.FilterByMember,
) (int, error) {
	panic("unexpected call to BillingRepository.CountByMember")
}

func (f *fakeBillingRepository) Save(
	context.Context,
	*billing.Invoice,
) error {
	panic("unexpected call to BillingRepository.Save")
}

type fakeContractRepository struct {
	getStatsFn    func(context.Context, domainproject.ProjectID) (domaincontract.Stats, error)
	getStatsCalls int
	stats         domaincontract.Stats
}

var _ domaincontract.Repository = (*fakeContractRepository)(nil)

func (f *fakeContractRepository) Add(
	context.Context,
	*domaincontract.Contract,
) error {
	panic("unexpected call to ContractRepository.Add")
}

func (f *fakeContractRepository) Get(
	context.Context,
	domaincontract.ContractID,
) (*domaincontract.Contract, error) {
	panic("unexpected call to ContractRepository.Get")
}

func (f *fakeContractRepository) GetStatsByProject(
	ctx context.Context,
	id domainproject.ProjectID,
) (domaincontract.Stats, error) {
	f.getStatsCalls++

	if f.getStatsFn != nil {
		return f.getStatsFn(ctx, id)
	}

	return f.stats, nil
}

func (f *fakeContractRepository) ListByProject(
	context.Context,
	domaincontract.FilterByProject,
	common.Page,
) ([]*domaincontract.Contract, error) {
	panic("unexpected call to ContractRepository.ListByProject")
}

func (f *fakeContractRepository) ListBySignatory(
	context.Context,
	domaincontract.FilterBySignatory,
	common.Page,
) ([]*domaincontract.Contract, error) {
	panic("unexpected call to ContractRepository.ListBySignatory")
}

func (f *fakeContractRepository) ListStatsByProjects(
	context.Context,
	[]domainproject.ProjectID,
) (map[domainproject.ProjectID]domaincontract.Stats, error) {
	panic("unexpected call to ContractRepository.ListStatsByProjects")
}

func (f *fakeContractRepository) CountByProject(
	context.Context,
	domaincontract.FilterByProject,
) (int, error) {
	panic("unexpected call to ContractRepository.CountByProject")
}

func (f *fakeContractRepository) CountBySignatory(
	context.Context,
	domaincontract.FilterBySignatory,
) (int, error) {
	panic("unexpected call to ContractRepository.CountBySignatory")
}

func (f *fakeContractRepository) Save(
	context.Context,
	*domaincontract.Contract,
) error {
	panic("unexpected call to ContractRepository.Save")
}

type fakeFileRepository struct {
	getStatsFn    func(context.Context, domainproject.ProjectID) (domainmessage.Stats, error)
	getStatsCalls int
	stats         domainmessage.Stats
}

var _ domainmessage.FileRepository = (*fakeFileRepository)(nil)

func (f *fakeFileRepository) Add(
	context.Context,
	*domainmessage.File,
) error {
	panic("unexpected call to FileRepository.Add")
}

func (f *fakeFileRepository) Get(
	context.Context,
	domainmessage.FileID,
) (*domainmessage.File, error) {
	panic("unexpected call to FileRepository.Get")
}

func (f *fakeFileRepository) GetStatsByProject(
	ctx context.Context,
	id domainproject.ProjectID,
) (domainmessage.Stats, error) {
	f.getStatsCalls++

	if f.getStatsFn != nil {
		return f.getStatsFn(ctx, id)
	}

	return f.stats, nil
}

func (f *fakeFileRepository) ListByChannel(
	context.Context,
	domainmessage.ChannelFileFilter,
	common.Page,
) ([]*domainmessage.File, error) {
	panic("unexpected call to FileRepository.ListByChannel")
}

func (f *fakeFileRepository) ListByProject(
	context.Context,
	domainmessage.ProjectFileFilter,
	common.Page,
) ([]*domainmessage.File, error) {
	panic("unexpected call to FileRepository.ListByProject")
}

func (f *fakeFileRepository) CountByChannel(
	context.Context,
	domainmessage.ChannelFileFilter,
) (int, error) {
	panic("unexpected call to FileRepository.CountByChannel")
}

func (f *fakeFileRepository) CountByProject(
	context.Context,
	domainmessage.ProjectFileFilter,
) (int, error) {
	panic("unexpected call to FileRepository.CountByProject")
}

type fakeUnitOfWork struct {
	calls     int
	executeFn func(context.Context, func(context.Context) error) error
}

var _ uow.UnitOfWork = (*fakeUnitOfWork)(nil)

func (f *fakeUnitOfWork) Execute(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	f.calls++

	if f.executeFn != nil {
		return f.executeFn(ctx, fn)
	}

	return fn(ctx)
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

type fakeInternalBus struct {
	calls      int
	events     []common.Event
	publishErr error
}

var _ eventbus.InternalBus = (*fakeInternalBus)(nil)

func (f *fakeInternalBus) Publish(
	_ context.Context,
	event common.Event,
) error {
	f.calls++
	f.events = append(f.events, event)

	return f.publishErr
}

func (f *fakeInternalBus) Subscribe(
	common.EventType,
	eventbus.EventHandler,
) {
	panic("unexpected call to InternalBus.Subscribe")
}

type fakeLogger struct {
	warnCalls int
}

var _ logger.Logger = (*fakeLogger)(nil)

func (f *fakeLogger) Info(string, ...any) {}

func (f *fakeLogger) Warn(string, ...any) {
	f.warnCalls++
}

func (f *fakeLogger) Error(string, ...any) {}

func (f *fakeLogger) Fatal(string, ...any) {
	panic("unexpected call to Logger.Fatal")
}

type testServiceDeps struct {
	userRepo     *fakeUserRepository
	projectRepo  *fakeProjectRepository
	taskRepo     *fakeTaskRepository
	billingRepo  *fakeBillingRepository
	contractRepo *fakeContractRepository
	fileRepo     *fakeFileRepository
	uow          *fakeUnitOfWork
	idGen        *fakeIDGenerator
	bus          *fakeInternalBus
	logger       *fakeLogger
}

func newTestService() (*Service, *testServiceDeps) {
	userRepo := &fakeUserRepository{}
	projectRepo := &fakeProjectRepository{}
	taskRepo := &fakeTaskRepository{}
	billingRepo := &fakeBillingRepository{}
	contractRepo := &fakeContractRepository{}
	fileRepo := &fakeFileRepository{}
	uow := &fakeUnitOfWork{}
	idGen := &fakeIDGenerator{}
	bus := &fakeInternalBus{}
	logger := &fakeLogger{}

	authzSrv := authz.NewService(userRepo)

	projectSrv := domainproject.NewService(
		userRepo,
		projectRepo,
	)

	service := NewService(
		userRepo,
		projectRepo,
		taskRepo,
		billingRepo,
		contractRepo,
		fileRepo,
		uow,
		authzSrv,
		projectSrv,
		bus,
		idGen,
		logger,
	)

	return service, &testServiceDeps{
		userRepo:     userRepo,
		projectRepo:  projectRepo,
		taskRepo:     taskRepo,
		billingRepo:  billingRepo,
		contractRepo: contractRepo,
		fileRepo:     fileRepo,
		uow:          uow,
		idGen:        idGen,
		bus:          bus,
		logger:       logger,
	}
}

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

func newTestUser(
	t *testing.T,
	id string,
	role domainiam.Role,
	active bool,
) *domainiam.User {
	t.Helper()

	userID := newTestUserID(t, id)

	name, err := domainiam.NewName("Test", "User")
	if err != nil {
		t.Fatalf("failed to create name: %v", err)
	}

	email, err := domainiam.NewEmail(id + "@example.com")
	if err != nil {
		t.Fatalf("failed to create email: %v", err)
	}

	now := time.Now()

	return domainiam.NewUser(
		userID,
		name,
		email,
		nil,
		role,
		active,
		userID,
		now,
	)
}

func newTestAdmin(t *testing.T, id string) *domainiam.User {
	return newTestUser(
		t,
		id,
		domainiam.RoleAdministrator,
		true,
	)
}

func newTestProject(
	t *testing.T,
	id string,
	creator domainiam.UserID,
	members ...domainiam.UserID,
) *domainproject.Project {
	t.Helper()

	projectID := newTestProjectID(t, id)

	name, err := domainproject.NewName("Test Project")
	if err != nil {
		t.Fatalf("failed to create project name: %v", err)
	}

	status, err := domainproject.NewStatus("started")
	if err != nil {
		t.Fatalf("failed to create project status: %v", err)
	}

	now := time.Now()

	p, err := domainproject.NewProject(
		projectID,
		name,
		status,
		creator,
		members,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	return p
}

func TestServiceReplaceMemberProjects(t *testing.T) {
	t.Run("requires administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return actor, nil
		}

		_ = member

		err := s.ReplaceMemberProjects(
			context.Background(),
			"member-1",
			"actor-1",
			nil,
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("returns user-not-found when member does not exist", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}
		deps.userRepo.existsResult = false

		err := s.ReplaceMemberProjects(
			context.Background(),
			"member-1",
			"admin-1",
			nil,
		)

		if !errors.Is(err, domainiam.ErrUserNotFound) {
			t.Fatalf("expected user-not-found, got %v", err)
		}

		if deps.projectRepo.listByMemberCalls != 0 {
			t.Fatal("expected projects not to be listed")
		}
	})

	t.Run("replaces member projects and saves affected projects", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		oldProject := newTestProject(
			t,
			"project-1",
			admin.ID(),
			admin.ID(),
			member.ID(),
		)

		newProject := newTestProject(
			t,
			"project-2",
			admin.ID(),
			admin.ID(),
		)

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.userRepo.existsResult = true
		deps.projectRepo.projects = []*domainproject.Project{newProject}

		deps.projectRepo.listByMemberFn = func(
			context.Context,
			domainiam.UserID,
		) ([]*domainproject.Project, error) {
			return []*domainproject.Project{oldProject}, nil
		}

		err := s.ReplaceMemberProjects(
			context.Background(),
			member.ID().String(),
			admin.ID().String(),
			[]string{"project-2"},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.projectRepo.listByMemberCalls != 1 {
			t.Fatalf(
				"expected one ListByMember call, got %d",
				deps.projectRepo.listByMemberCalls,
			)
		}

		if deps.projectRepo.listByIDsCalls != 1 {
			t.Fatalf(
				"expected one ListByIDs call, got %d",
				deps.projectRepo.listByIDsCalls,
			)
		}

		if deps.projectRepo.saveAllCalls != 1 {
			t.Fatalf(
				"expected one SaveAll call, got %d",
				deps.projectRepo.saveAllCalls,
			)
		}

		if len(deps.projectRepo.savedProjects) != 2 {
			t.Fatalf(
				"expected two affected projects, got %d",
				len(deps.projectRepo.savedProjects),
			)
		}

		if oldProject.HasMember(member.ID()) {
			t.Fatal("expected member to be removed from old project")
		}

		if !newProject.HasMember(member.ID()) {
			t.Fatal("expected member to be added to new project")
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}
		deps.userRepo.existsErr = errRepository

		err := s.ReplaceMemberProjects(
			context.Background(),
			"member-1",
			"admin-1",
			nil,
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestServiceCreateProject(t *testing.T) {
	t.Run("requires administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return actor, nil
		}

		err := s.CreateProject(
			context.Background(),
			CreateProjectParams{
				ActorID: "actor-1",
				Name:    "Test Project",
				Status:  "planning",
			},
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}

		if deps.idGen.calls != 0 {
			t.Fatal("expected ID generator not to be called")
		}
	})

	t.Run("creates project", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.userRepo.existsAllResult = true
		deps.idGen.id = "project-1"

		err := s.CreateProject(
			context.Background(),
			CreateProjectParams{
				ActorID: "admin-1",
				Name:    "Test Project",
				Status:  "started",
				Members: []string{member.ID().String()},
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.idGen.calls != 1 {
			t.Fatalf("expected one ID generation, got %d", deps.idGen.calls)
		}

		if deps.userRepo.existsAllCalls != 1 {
			t.Fatalf(
				"expected one ExistsAll call, got %d",
				deps.userRepo.existsAllCalls,
			)
		}

		if deps.uow.calls != 1 {
			t.Fatalf("expected one UOW call, got %d", deps.uow.calls)
		}

		if deps.projectRepo.addCalls != 1 {
			t.Fatalf(
				"expected one project add, got %d",
				deps.projectRepo.addCalls,
			)
		}

		if deps.projectRepo.addedProject == nil {
			t.Fatal("expected added project")
		}

		if deps.projectRepo.addedProject.ID().String() != "project-1" {
			t.Fatalf(
				"expected project-1, got %s",
				deps.projectRepo.addedProject.ID(),
			)
		}

		if !deps.projectRepo.addedProject.HasMember(admin.ID()) {
			t.Fatal("expected creator to be a project member")
		}

		if !deps.projectRepo.addedProject.HasMember(member.ID()) {
			t.Fatal("expected supplied member to be a project member")
		}

		if deps.bus.calls == 0 {
			t.Fatal("expected project events to be published")
		}
	})

	t.Run("returns user-not-found when a member does not exist", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.userRepo.existsAllResult = false
		deps.idGen.id = "project-1"

		err := s.CreateProject(
			context.Background(),
			CreateProjectParams{
				ActorID: "admin-1",
				Name:    "Test Project",
				Status:  "started",
				Members: []string{"missing-user"},
			},
		)

		if !errors.Is(err, domainiam.ErrUserNotFound) {
			t.Fatalf("expected user-not-found, got %v", err)
		}

		if deps.projectRepo.addCalls != 0 {
			t.Fatal("expected project not to be added")
		}

		if deps.uow.calls != 0 {
			t.Fatal("expected UOW not to run")
		}
	})

	t.Run("propagates ID generation error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.idGen.err = errIDGen

		err := s.CreateProject(
			context.Background(),
			CreateProjectParams{
				ActorID: "admin-1",
				Name:    "Test Project",
				Status:  "started",
			},
		)

		if !errors.Is(err, errIDGen) {
			t.Fatalf("expected ID-generation error, got %v", err)
		}
	})

	t.Run("does not publish events when UOW fails", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.userRepo.existsAllResult = true
		deps.idGen.id = "project-1"
		deps.uow.executeFn = func(
			context.Context,
			func(context.Context) error,
		) error {
			return errUOW
		}

		err := s.CreateProject(
			context.Background(),
			CreateProjectParams{
				ActorID: "admin-1",
				Name:    "Test Project",
				Status:  "started",
			},
		)

		if !errors.Is(err, errUOW) {
			t.Fatalf("expected UOW error, got %v", err)
		}

		if deps.bus.calls != 0 {
			t.Fatal("expected no events to be published")
		}
	})
}

func TestServiceListStats(t *testing.T) {
	t.Run("returns project statistics", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.projects = []*domainproject.Project{p}
		deps.projectRepo.listCount = 7

		result, err := s.ListStats(
			context.Background(),
			ListStatsParams{
				ActorID:  admin.ID().String(),
				MemberID: admin.ID().String(),
				Limit:    10,
				Offset:   0,
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result == nil {
			t.Fatal("expected collection")
		}

		if len(result.Items) != 1 {
			t.Fatalf("expected one item, got %d", len(result.Items))
		}

		if result.TotalCount != 7 {
			t.Fatalf(
				"expected total count 7, got %d",
				result.TotalCount,
			)
		}

		if result.Items[0].ID != p.ID().String() {
			t.Fatalf(
				"expected project ID %s, got %s",
				p.ID(),
				result.Items[0].ID,
			)
		}

		if result.Items[0].Name != p.Name().String() {
			t.Fatalf(
				"expected project name %s, got %s",
				p.Name(),
				result.Items[0].Name,
			)
		}

		if result.Items[0].Status != p.Status().String() {
			t.Fatalf(
				"expected status %s, got %s",
				p.Status(),
				result.Items[0].Status,
			)
		}

		if deps.taskRepo.listStatsCalls != 1 {
			t.Fatal("expected task stats to be queried")
		}

		if deps.billingRepo.listStatsCalls != 1 {
			t.Fatal("expected billing stats to be queried")
		}

		if deps.projectRepo.countCalls != 1 {
			t.Fatal("expected project count to be queried")
		}
	})

	t.Run("propagates project repository error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.listFn = func(
			context.Context,
			domainproject.Filter,
			common.Page,
		) ([]*domainproject.Project, error) {
			return nil, errRepository
		}

		_, err := s.ListStats(
			context.Background(),
			ListStatsParams{
				ActorID:  admin.ID().String(),
				MemberID: admin.ID().String(),
				Limit:    10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("propagates task statistics error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.projects = []*domainproject.Project{p}

		deps.taskRepo.listStatsFn = func(
			context.Context,
			[]domainproject.ProjectID,
		) (map[domainproject.ProjectID]domaintask.Stats, error) {
			return nil, errRepository
		}

		_, err := s.ListStats(
			context.Background(),
			ListStatsParams{
				ActorID:  admin.ID().String(),
				MemberID: admin.ID().String(),
				Limit:    10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("propagates billing statistics error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.projects = []*domainproject.Project{p}

		deps.billingRepo.listStatsFn = func(
			context.Context,
			[]domainproject.ProjectID,
		) (map[domainproject.ProjectID]billing.Stats, error) {
			return nil, errRepository
		}

		_, err := s.ListStats(
			context.Background(),
			ListStatsParams{
				ActorID:  admin.ID().String(),
				MemberID: admin.ID().String(),
				Limit:    10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns forbidden for unrelated non-administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		target := newTestUser(t, "target-1", domainiam.RoleClient, true)

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return actor, nil
		}

		_, err := s.ListStats(
			context.Background(),
			ListStatsParams{
				ActorID:  actor.ID().String(),
				MemberID: target.ID().String(),
				Limit:    10,
			},
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})
}

func TestServiceListAllStats(t *testing.T) {
	t.Run("returns statistics for member projects", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.projects = []*domainproject.Project{p}

		result, err := s.ListAllStats(
			context.Background(),
			admin.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected one result, got %d", len(result))
		}

		if result[0].ID != p.ID().String() {
			t.Fatalf("expected project ID %s, got %s", p.ID(), result[0].ID)
		}

		if deps.projectRepo.listByMemberCalls != 1 {
			t.Fatal("expected ListByMember to be called")
		}

		if deps.taskRepo.listStatsCalls != 1 {
			t.Fatal("expected task statistics to be queried")
		}

		if deps.billingRepo.listStatsCalls != 1 {
			t.Fatal("expected billing statistics to be queried")
		}
	})

	t.Run("propagates project repository error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.listByMemberFn = func(
			context.Context,
			domainiam.UserID,
		) ([]*domainproject.Project, error) {
			return nil, errRepository
		}

		_, err := s.ListAllStats(
			context.Background(),
			admin.ID().String(),
			admin.ID().String(),
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestServiceGetProjectOverview(t *testing.T) {
	t.Run("returns project overview", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			admin.ID(),
			admin.ID(),
			member.ID(),
		)

		deps.projectRepo.project = p

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]domainiam.UserID,
			domainiam.UserFilter,
		) ([]*domainiam.User, error) {
			return []*domainiam.User{admin, member}, nil
		}

		result, err := s.GetProjectOverview(
			context.Background(),
			p.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result.ID != p.ID().String() {
			t.Fatalf("expected ID %s, got %s", p.ID(), result.ID)
		}

		if result.Name != p.Name().String() {
			t.Fatalf("expected name %s, got %s", p.Name(), result.Name)
		}

		if result.Status != p.Status().String() {
			t.Fatalf(
				"expected status %s, got %s",
				p.Status(),
				result.Status,
			)
		}

		if len(result.Members) != 2 {
			t.Fatalf(
				"expected two members, got %d",
				len(result.Members),
			)
		}

		if deps.taskRepo.getStatsCalls != 1 {
			t.Fatal("expected task stats query")
		}

		if deps.billingRepo.getStatsCalls != 1 {
			t.Fatal("expected billing stats query")
		}

		if deps.contractRepo.getStatsCalls != 1 {
			t.Fatal("expected contract stats query")
		}

		if deps.fileRepo.getStatsCalls != 1 {
			t.Fatal("expected file stats query")
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			admin.ID(),
			admin.ID(),
		)

		deps.projectRepo.project = p

		_, err := s.GetProjectOverview(
			context.Background(),
			p.ID().String(),
			member.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected not-project-member, got %v", err)
		}

		if deps.taskRepo.getStatsCalls != 0 {
			t.Fatal("expected task stats not to be queried")
		}
	})

	t.Run("propagates task statistics error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.projectRepo.project = p
		deps.taskRepo.getStatsFn = func(
			context.Context,
			domainproject.ProjectID,
		) (domaintask.Stats, error) {
			return domaintask.Stats{}, errRepository
		}

		_, err := s.GetProjectOverview(
			context.Background(),
			p.ID().String(),
			admin.ID().String(),
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestServiceGetCreatedCount(t *testing.T) {
	t.Run("returns project metrics", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestAdmin(t, "admin-1")

		deps.projectRepo.listCreatedFn = func(
			context.Context,
			domainiam.UserID,
		) ([]domainproject.Metric, error) {
			return nil, nil
		}

		result, err := s.GetCreatedCount(
			context.Background(),
			actor.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if deps.projectRepo.listCreatedCalls != 1 {
			t.Fatalf(
				"expected one ListCreatedPerDay call, got %d",
				deps.projectRepo.listCreatedCalls,
			)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.projectRepo.listCreatedFn = func(
			context.Context,
			domainiam.UserID,
		) ([]domainproject.Metric, error) {
			return nil, errRepository
		}

		_, err := s.GetCreatedCount(
			context.Background(),
			"admin-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestServiceListMembers(t *testing.T) {
	t.Run("returns project members", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			admin.ID(),
			admin.ID(),
			member.ID(),
		)

		deps.projectRepo.project = p

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]domainiam.UserID,
			domainiam.UserFilter,
		) ([]*domainiam.User, error) {
			return []*domainiam.User{admin, member}, nil
		}

		result, err := s.ListMembers(
			context.Background(),
			ListMembersParams{
				ActorID:   admin.ID().String(),
				ProjectID: p.ID().String(),
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected two members, got %d", len(result))
		}

		if result[0].ID != admin.ID().String() {
			t.Fatalf(
				"expected first member %s, got %s",
				admin.ID(),
				result[0].ID,
			)
		}

		if result[0].Role != admin.Role().String() {
			t.Fatalf(
				"expected role %s, got %s",
				admin.Role(),
				result[0].Role,
			)
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		outsider := newTestUser(t, "outsider-1", domainiam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			admin.ID(),
			admin.ID(),
		)

		deps.projectRepo.project = p

		_, err := s.ListMembers(
			context.Background(),
			ListMembersParams{
				ActorID:   outsider.ID().String(),
				ProjectID: p.ID().String(),
			},
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected not-project-member, got %v", err)
		}

		if deps.userRepo.listByIDsCalls != 0 {
			t.Fatal("expected users not to be queried")
		}
	})

	t.Run("propagates user repository error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.projectRepo.project = p

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]domainiam.UserID,
			domainiam.UserFilter,
		) ([]*domainiam.User, error) {
			return nil, errRepository
		}

		_, err := s.ListMembers(
			context.Background(),
			ListMembersParams{
				ActorID:   admin.ID().String(),
				ProjectID: p.ID().String(),
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestServiceReplaceMembers(t *testing.T) {
	t.Run("requires administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return actor, nil
		}

		err := s.ReplaceMembers(
			context.Background(),
			"project-1",
			[]string{"member-1"},
			actor.ID().String(),
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("returns not-project-member when administrator is not a project member", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			member.ID(),
			member.ID(),
		)

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		err := s.ReplaceMembers(
			context.Background(),
			p.ID().String(),
			[]string{"member-1"},
			admin.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected not-project-member, got %v", err)
		}

		if deps.projectRepo.saveCalls != 0 {
			t.Fatal("expected project not to be saved")
		}
	})

	t.Run("replaces members and saves project", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			admin.ID(),
			admin.ID(),
		)

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.userRepo.existsAllResult = true

		// EnsureHasAdministrator checks this after replacement.
		deps.userRepo.isAdminFn = func(
			context.Context,
			[]domainiam.UserID,
		) (bool, error) {
			return true, nil
		}

		err := s.ReplaceMembers(
			context.Background(),
			p.ID().String(),
			[]string{member.ID().String()},
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.userRepo.existsAllCalls != 1 {
			t.Fatal("expected ExistsAll to be called")
		}

		if deps.userRepo.isAdminCalls != 1 {
			t.Fatal("expected IsAnyAdministrator to be called")
		}

		if deps.projectRepo.saveCalls != 1 {
			t.Fatalf(
				"expected one save, got %d",
				deps.projectRepo.saveCalls,
			)
		}

		if !p.HasMember(member.ID()) {
			t.Fatal("expected replacement member to be present")
		}

		if p.HasMember(admin.ID()) {
			t.Fatal("expected old member set to be replaced")
		}
	})

	t.Run("returns user-not-found when member does not exist", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.userRepo.existsAllResult = false

		err := s.ReplaceMembers(
			context.Background(),
			p.ID().String(),
			[]string{"missing-user"},
			admin.ID().String(),
		)

		if !errors.Is(err, domainiam.ErrUserNotFound) {
			t.Fatalf("expected user-not-found, got %v", err)
		}

		if deps.projectRepo.saveCalls != 0 {
			t.Fatal("expected project not to be saved")
		}
	})

	t.Run("propagates administrator invariant error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.userRepo.existsAllResult = true
		deps.userRepo.isAdminFn = func(
			context.Context,
			[]domainiam.UserID,
		) (bool, error) {
			return false, nil
		}

		err := s.ReplaceMembers(
			context.Background(),
			p.ID().String(),
			[]string{member.ID().String()},
			admin.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrProjectMustHaveAdministrator) {
			t.Fatalf(
				"expected administrator invariant error, got %v",
				err,
			)
		}

		if deps.projectRepo.saveCalls != 0 {
			t.Fatal("expected project not to be saved")
		}
	})
}

func TestServiceStartProject(t *testing.T) {
	t.Run("starts project", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.project = p

		err := s.StartProject(
			context.Background(),
			p.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.projectRepo.saveCalls != 1 {
			t.Fatalf("expected one save, got %d", deps.projectRepo.saveCalls)
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		outsider := newTestUser(t, "outsider-1", domainiam.RoleClient, true)
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return outsider, nil
		}

		deps.projectRepo.project = p

		_, err := func() (struct{}, error) {
			return struct{}{}, s.StartProject(
				context.Background(),
				p.ID().String(),
				outsider.ID().String(),
			)
		}()

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})
}

func TestServicePauseProject(t *testing.T) {
	t.Run("saves paused project", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.project = p

		err := s.PauseProject(
			context.Background(),
			p.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.projectRepo.saveCalls != 1 {
			t.Fatal("expected one save")
		}
	})
}

func TestServiceCancelProject(t *testing.T) {
	t.Run("saves cancelled project", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.project = p

		err := s.CancelProject(
			context.Background(),
			p.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.projectRepo.saveCalls != 1 {
			t.Fatal("expected one save")
		}
	})
}

func TestServiceCompleteProject(t *testing.T) {
	t.Run("saves completed project", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.project = p

		err := s.CompleteProject(
			context.Background(),
			p.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.projectRepo.saveCalls != 1 {
			t.Fatal("expected one save")
		}
	})
}

func TestServiceDelete(t *testing.T) {
	t.Run("requires administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return actor, nil
		}

		err := s.Delete(
			context.Background(),
			"project-1",
			actor.ID().String(),
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("deletes project", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.project = p

		err := s.Delete(
			context.Background(),
			p.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.projectRepo.removeCalls != 1 {
			t.Fatalf(
				"expected one remove, got %d",
				deps.projectRepo.removeCalls,
			)
		}

		if deps.projectRepo.removedProject != p {
			t.Fatal("expected target project to be removed")
		}
	})

	t.Run("rejects non-member administrator", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			member.ID(),
			member.ID(),
		)

		deps.userRepo.getByIDFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.User, error) {
			return admin, nil
		}

		deps.projectRepo.project = p

		err := s.Delete(
			context.Background(),
			p.ID().String(),
			admin.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected not-project-member, got %v", err)
		}

		if deps.projectRepo.removeCalls != 0 {
			t.Fatal("expected project not to be removed")
		}
	})
}

func TestServicePublishEvents(t *testing.T) {
	t.Run("publishes all events", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		events := p.PullEvents()

		if len(events) == 0 {
			t.Fatal("expected project to contain domain events")
		}

		s.publishEvents(context.Background(), events)

		if deps.bus.calls != len(events) {
			t.Fatalf(
				"expected %d published events, got %d",
				len(events),
				deps.bus.calls,
			)
		}
	})

	t.Run("logs warning when publishing fails", func(t *testing.T) {
		s, deps := newTestService()

		deps.bus.publishErr = errEventBus

		admin := newTestAdmin(t, "admin-1")
		p := newTestProject(t, "project-1", admin.ID(), admin.ID())

		events := p.PullEvents()

		s.publishEvents(context.Background(), events)

		if deps.logger.warnCalls != len(events) {
			t.Fatalf(
				"expected %d warnings, got %d",
				len(events),
				deps.logger.warnCalls,
			)
		}
	})
}
