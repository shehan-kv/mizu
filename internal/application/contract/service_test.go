package contract

import (
	"context"
	"errors"
	"testing"
	"time"

	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/logger"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	domaincontract "mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	domainiam "mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	domainproject "mizu/internal/domain/project"
)

var (
	errRepository = errors.New("repository failure")
	errIDGen      = errors.New("id generation failure")
	errPublish    = errors.New("publish failure")
	errMail       = errors.New("mail failure")
)

type fakeUserRepository struct {
	getByIDFn   func(context.Context, iam.UserID) (*iam.User, error)
	listByIDsFn func(context.Context, []iam.UserID, iam.UserFilter) ([]*iam.User, error)
	existsAllFn func(context.Context, []iam.UserID) (bool, error)
	isAdminFn   func(context.Context, []iam.UserID) (bool, error)

	getByIDCalls   int
	listByIDsCalls int
	existsAllCalls int
	isAdminCalls   int

	listByIDsIDs    []iam.UserID
	listByIDsFilter iam.UserFilter
}

var _ iam.UserRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) Add(context.Context, *iam.User) error {
	panic("unexpected call to Add")
}

func (f *fakeUserRepository) Exists(context.Context, iam.UserID) (bool, error) {
	panic("unexpected call to Exists")
}

func (f *fakeUserRepository) ExistsAll(
	ctx context.Context,
	ids []iam.UserID,
) (bool, error) {
	f.existsAllCalls++

	if f.existsAllFn != nil {
		return f.existsAllFn(ctx, ids)
	}

	return true, nil
}

func (f *fakeUserRepository) GetByID(
	ctx context.Context,
	id iam.UserID,
) (*iam.User, error) {
	f.getByIDCalls++

	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	return nil, errRepository
}

func (f *fakeUserRepository) GetByEmail(
	context.Context,
	iam.Email,
) (*iam.User, error) {
	panic("unexpected call to GetByEmail")
}

func (f *fakeUserRepository) List(
	context.Context,
	iam.UserFilter,
	common.Page,
) ([]*iam.User, error) {
	panic("unexpected call to List")
}

func (f *fakeUserRepository) ListByIDs(
	ctx context.Context,
	ids []iam.UserID,
	filter iam.UserFilter,
) ([]*iam.User, error) {
	f.listByIDsCalls++
	f.listByIDsIDs = append([]iam.UserID(nil), ids...)
	f.listByIDsFilter = filter

	if f.listByIDsFn != nil {
		return f.listByIDsFn(ctx, ids, filter)
	}

	return nil, nil
}

func (f *fakeUserRepository) IsAnyAdministrator(
	ctx context.Context,
	ids []iam.UserID,
) (bool, error) {
	f.isAdminCalls++

	if f.isAdminFn != nil {
		return f.isAdminFn(ctx, ids)
	}

	return true, nil
}

func (f *fakeUserRepository) HasAdministrator(context.Context) (bool, error) {
	panic("unexpected call to HasAdministrator")
}

func (f *fakeUserRepository) Count(
	context.Context,
	iam.UserFilter,
) (int, error) {
	panic("unexpected call to Count")
}

func (f *fakeUserRepository) Save(
	context.Context,
	*iam.User,
) error {
	panic("unexpected call to Save")
}

func (f *fakeUserRepository) Remove(
	context.Context,
	*iam.User,
) error {
	panic("unexpected call to Remove")
}

type fakeContractRepository struct {
	addFn  func(context.Context, *domaincontract.Contract) error
	getFn  func(context.Context, domaincontract.ContractID) (*domaincontract.Contract, error)
	saveFn func(context.Context, *domaincontract.Contract) error

	listByProjectFn func(
		context.Context,
		domaincontract.FilterByProject,
		common.Page,
	) ([]*domaincontract.Contract, error)

	listBySignatoryFn func(
		context.Context,
		domaincontract.FilterBySignatory,
		common.Page,
	) ([]*domaincontract.Contract, error)

	countByProjectFn func(
		context.Context,
		domaincontract.FilterByProject,
	) (int, error)

	countBySignatoryFn func(
		context.Context,
		domaincontract.FilterBySignatory,
	) (int, error)

	addCalls              int
	getCalls              int
	saveCalls             int
	listByProjectCalls    int
	listBySignatoryCalls  int
	countByProjectCalls   int
	countBySignatoryCalls int

	addedContract *domaincontract.Contract
	savedContract *domaincontract.Contract

	projectFilter   domaincontract.FilterByProject
	signatoryFilter domaincontract.FilterBySignatory
	projectPage     common.Page
	signatoryPage   common.Page
}

var _ domaincontract.Repository = (*fakeContractRepository)(nil)

func (f *fakeContractRepository) Add(
	ctx context.Context,
	c *domaincontract.Contract,
) error {
	f.addCalls++
	f.addedContract = c

	if f.addFn != nil {
		return f.addFn(ctx, c)
	}

	return nil
}

func (f *fakeContractRepository) Get(
	ctx context.Context,
	id domaincontract.ContractID,
) (*domaincontract.Contract, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	return nil, errRepository
}

func (f *fakeContractRepository) GetStatsByProject(
	context.Context,
	domainproject.ProjectID,
) (domaincontract.Stats, error) {
	panic("unexpected call to GetStatsByProject")
}

func (f *fakeContractRepository) ListByProject(
	ctx context.Context,
	filter domaincontract.FilterByProject,
	page common.Page,
) ([]*domaincontract.Contract, error) {
	f.listByProjectCalls++
	f.projectFilter = filter
	f.projectPage = page

	if f.listByProjectFn != nil {
		return f.listByProjectFn(ctx, filter, page)
	}

	return nil, nil
}

func (f *fakeContractRepository) ListBySignatory(
	ctx context.Context,
	filter domaincontract.FilterBySignatory,
	page common.Page,
) ([]*domaincontract.Contract, error) {
	f.listBySignatoryCalls++
	f.signatoryFilter = filter
	f.signatoryPage = page

	if f.listBySignatoryFn != nil {
		return f.listBySignatoryFn(ctx, filter, page)
	}

	return nil, nil
}

func (f *fakeContractRepository) ListStatsByProjects(
	context.Context,
	[]domainproject.ProjectID,
) (map[domainproject.ProjectID]domaincontract.Stats, error) {
	panic("unexpected call to ListStatsByProjects")
}

func (f *fakeContractRepository) CountByProject(
	ctx context.Context,
	filter domaincontract.FilterByProject,
) (int, error) {
	f.countByProjectCalls++

	if f.countByProjectFn != nil {
		return f.countByProjectFn(ctx, filter)
	}

	return 0, nil
}

func (f *fakeContractRepository) CountBySignatory(
	ctx context.Context,
	filter domaincontract.FilterBySignatory,
) (int, error) {
	f.countBySignatoryCalls++

	if f.countBySignatoryFn != nil {
		return f.countBySignatoryFn(ctx, filter)
	}

	return 0, nil
}

func (f *fakeContractRepository) Save(
	ctx context.Context,
	c *domaincontract.Contract,
) error {
	f.saveCalls++
	f.savedContract = c

	if f.saveFn != nil {
		return f.saveFn(ctx, c)
	}

	return nil
}

type fakeProjectRepository struct {
	getFn func(
		context.Context,
		domainproject.ProjectID,
	) (*domainproject.Project, error)

	getCalls int
	project  *domainproject.Project
}

var _ domainproject.Repository = (*fakeProjectRepository)(nil)

func (f *fakeProjectRepository) Add(
	context.Context,
	*domainproject.Project,
) error {
	panic("unexpected call to Add")
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
	context.Context,
	domainproject.Filter,
	common.Page,
) ([]*domainproject.Project, error) {
	panic("unexpected call to List")
}

func (f *fakeProjectRepository) ListByIDs(
	context.Context,
	[]domainproject.ProjectID,
) ([]*domainproject.Project, error) {
	panic("unexpected call to ListByIDs")
}

func (f *fakeProjectRepository) ListByMember(
	context.Context,
	iam.UserID,
) ([]*domainproject.Project, error) {
	panic("unexpected call to ListByMember")
}

func (f *fakeProjectRepository) ListCreatedPerDay(
	context.Context,
	iam.UserID,
) ([]domainproject.Metric, error) {
	panic("unexpected call to ListCreatedPerDay")
}

func (f *fakeProjectRepository) Count(
	context.Context,
	domainproject.Filter,
) (int, error) {
	panic("unexpected call to Count")
}

func (f *fakeProjectRepository) Save(
	context.Context,
	*domainproject.Project,
) error {
	panic("unexpected call to Save")
}

func (f *fakeProjectRepository) SaveAll(
	context.Context,
	[]*domainproject.Project,
) error {
	panic("unexpected call to SaveAll")
}

func (f *fakeProjectRepository) Remove(
	context.Context,
	*domainproject.Project,
) error {
	panic("unexpected call to Remove")
}

type fakeInternalBus struct {
	publishFn func(context.Context, common.Event) error

	calls  int
	events []common.Event
}

var _ eventbus.InternalBus = (*fakeInternalBus)(nil)

func (f *fakeInternalBus) Publish(
	ctx context.Context,
	event common.Event,
) error {
	f.calls++
	f.events = append(f.events, event)

	if f.publishFn != nil {
		return f.publishFn(ctx, event)
	}

	return nil
}

func (f *fakeInternalBus) Subscribe(
	common.EventType,
	eventbus.EventHandler,
) {
	panic("unexpected call to Subscribe")
}

type fakeIDGenerator struct {
	id    string
	err   error
	calls int
}

var _ common.IDGenerator = (*fakeIDGenerator)(nil)

func (f *fakeIDGenerator) Generate() (string, error) {
	f.calls++

	if f.err != nil {
		return "", f.err
	}

	return f.id, nil
}

type fakeMailer struct {
	sendContractEmailFn func(
		context.Context,
		mailer.ContractEmail,
	) error

	sendContractEmailCalls int
	contractEmail          mailer.ContractEmail
}

var _ mailer.Mailer = (*fakeMailer)(nil)

func (f *fakeMailer) SendRecoveryEmail(
	context.Context,
	iam.Email,
	iam.RecoveryToken,
) error {
	panic("unexpected call to SendRecoveryEmail")
}

func (f *fakeMailer) SendVerificationEmail(
	context.Context,
	iam.Email,
	iam.VerificationID,
) error {
	panic("unexpected call to SendVerificationEmail")
}

func (f *fakeMailer) SendVerifiedEmail(
	context.Context,
	iam.Email,
) error {
	panic("unexpected call to SendVerifiedEmail")
}

func (f *fakeMailer) SendContractEmail(
	ctx context.Context,
	email mailer.ContractEmail,
) error {
	f.sendContractEmailCalls++
	f.contractEmail = email

	if f.sendContractEmailFn != nil {
		return f.sendContractEmailFn(ctx, email)
	}

	return nil
}

func (f *fakeMailer) SendInvoiceEmail(
	context.Context,
	mailer.InvoiceEmail,
) error {
	panic("unexpected call to SendInvoiceEmail")
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

func (f *fakeLogger) Fatal(string, ...any) {}

type testServiceDeps struct {
	userRepo     *fakeUserRepository
	contractRepo *fakeContractRepository
	projectRepo  *fakeProjectRepository
	bus          *fakeInternalBus
	idGen        *fakeIDGenerator
	mailer       *fakeMailer
	logger       *fakeLogger
}

func newTestService() (*Service, *testServiceDeps) {
	userRepo := &fakeUserRepository{}
	contractRepo := &fakeContractRepository{}
	projectRepo := &fakeProjectRepository{}
	bus := &fakeInternalBus{}
	idGen := &fakeIDGenerator{id: "contract-1"}
	mailer := &fakeMailer{}
	logger := &fakeLogger{}

	authzSrv := authz.NewService(userRepo)
	contractSrv := domaincontract.NewService()

	s := NewService(
		userRepo,
		contractRepo,
		projectRepo,
		contractSrv,
		authzSrv,
		bus,
		idGen,
		mailer,
		logger,
	)

	return s, &testServiceDeps{
		userRepo:     userRepo,
		contractRepo: contractRepo,
		projectRepo:  projectRepo,
		bus:          bus,
		idGen:        idGen,
		mailer:       mailer,
		logger:       logger,
	}
}

func newTestContract(
	t *testing.T,
	id string,
	projectID string,
	signatories []iam.UserID,
) *domaincontract.Contract {
	t.Helper()

	contractID, err := domaincontract.NewContractID(id)
	if err != nil {
		t.Fatalf("failed to create contract ID: %v", err)
	}

	projectIDValue, err := project.NewProjectID(projectID)
	if err != nil {
		t.Fatalf("failed to create project ID: %v", err)
	}

	name, err := domaincontract.NewName("Test Contract")
	if err != nil {
		t.Fatalf("failed to create contract name: %v", err)
	}

	terms, err := domaincontract.NewTerms(
		"These are valid contract terms containing more than fifty characters.",
	)
	if err != nil {
		t.Fatalf("failed to create contract terms: %v", err)
	}

	domainSignatories := make([]domaincontract.Signatory, len(signatories))

	now := time.Now()

	for i, userID := range signatories {
		domainSignatories[i] = domaincontract.NewSignatory(
			userID,
			now,
		)
	}

	c, err := domaincontract.NewContract(
		contractID,
		projectIDValue,
		name,
		terms,
		domainSignatories,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create contract: %v", err)
	}

	c.PullEvents()

	return c
}

func newTestUserID(t *testing.T, id string) domainiam.UserID {
	t.Helper()

	userID, err := domainiam.NewUserID(id)
	if err != nil {
		t.Fatalf("failed to create user ID: %v", err)
	}

	return userID
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

func newTestProjectID(t *testing.T, id string) domainproject.ProjectID {
	t.Helper()

	projectID, err := domainproject.NewProjectID(id)
	if err != nil {
		t.Fatalf("failed to create project ID: %v", err)
	}

	return projectID
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

func TestServiceSign(t *testing.T) {
	t.Run("returns repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return nil, errRepository
		}

		err := s.Sign(
			context.Background(),
			"user-1",
			"contract-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("signs and saves contract", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUserID(t, "user-1")
		other := newTestUserID(t, "user-2")

		c := newTestContract(
			t,
			"contract-1",
			"project-1",
			[]iam.UserID{actor, other},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		err := s.Sign(
			context.Background(),
			actor.String(),
			c.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.contractRepo.saveCalls != 1 {
			t.Fatalf(
				"expected one save, got %d",
				deps.contractRepo.saveCalls,
			)
		}

		if deps.bus.calls != 1 {
			t.Fatalf(
				"expected one published event, got %d",
				deps.bus.calls,
			)
		}
	})

	t.Run("propagates save error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUserID(t, "user-1")
		other := newTestUserID(t, "user-2")

		c := newTestContract(
			t,
			"contract-1",
			"project-1",
			[]iam.UserID{actor, other},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.contractRepo.saveFn = func(
			context.Context,
			*domaincontract.Contract,
		) error {
			return errRepository
		}

		err := s.Sign(
			context.Background(),
			actor.String(),
			c.ID().String(),
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}

		if deps.bus.calls != 0 {
			t.Fatal("expected no events to be published after save failure")
		}
	})
}

func TestServiceReject(t *testing.T) {
	t.Run("rejects and saves contract", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUserID(t, "user-1")
		other := newTestUserID(t, "user-2")

		c := newTestContract(
			t,
			"contract-1",
			"project-1",
			[]iam.UserID{actor, other},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		err := s.Reject(
			context.Background(),
			actor.String(),
			c.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.contractRepo.saveCalls != 1 {
			t.Fatalf(
				"expected one save, got %d",
				deps.contractRepo.saveCalls,
			)
		}

		if deps.bus.calls != 1 {
			t.Fatalf(
				"expected one published event, got %d",
				deps.bus.calls,
			)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return nil, errRepository
		}

		err := s.Reject(
			context.Background(),
			"user-1",
			"contract-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestServiceGetContract(t *testing.T) {
	t.Run("returns not-project-member", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)
		member := newTestUser(t, "member-1", iam.RoleClient, true)
		other := newTestUser(t, "other-1", iam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			member.ID(),
			member.ID(),
			other.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{member.ID(), other.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		_, err := s.GetContract(
			context.Background(),
			actor.ID().String(),
			c.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf(
				"expected not-project-member, got %v",
				err,
			)
		}
	})

	t.Run("returns signatory-not-found when user data is missing", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)
		other := newTestUser(t, "other-1", iam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
			other.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), other.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return nil, nil
		}

		_, err := s.GetContract(
			context.Background(),
			actor.ID().String(),
			c.ID().String(),
		)

		if !errors.Is(err, domaincontract.ErrContractSignatoryNotFound) {
			t.Fatalf(
				"expected signatory-not-found, got %v",
				err,
			)
		}
	})
	t.Run("maps contract and signatories", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)
		staff := newTestUser(t, "staff-1", iam.RoleStaff, true)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
			staff.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), staff.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return []*iam.User{actor, staff}, nil
		}

		result, err := s.GetContract(
			context.Background(),
			actor.ID().String(),
			c.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result == nil {
			t.Fatal("expected DTO")
		}

		if result.ID != c.ID().String() {
			t.Fatalf(
				"expected ID %s, got %s",
				c.ID(),
				result.ID,
			)
		}

		if result.ProjectID != p.ID().String() {
			t.Fatalf(
				"expected project ID %s, got %s",
				p.ID(),
				result.ProjectID,
			)
		}

		if result.Name != c.Name().String() {
			t.Fatalf(
				"expected name %s, got %s",
				c.Name(),
				result.Name,
			)
		}

		if result.Terms != c.Terms().String() {
			t.Fatalf(
				"expected terms %s, got %s",
				c.Terms(),
				result.Terms,
			)
		}

		if len(result.Signatories) != 2 {
			t.Fatalf(
				"expected two signatories, got %d",
				len(result.Signatories),
			)
		}

		if result.MemberSignatoryStatus == "" {
			t.Fatal("expected member signatory status")
		}
	})
}

func TestServiceListOverviewByProject(t *testing.T) {
	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)
		member := newTestUser(t, "member-1", iam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			member.ID(),
			member.ID(),
		)

		deps.projectRepo.project = p

		_, err := s.ListOverviewByProject(
			context.Background(),
			ListByProjectParams{
				ActorID:   actor.ID().String(),
				ProjectID: p.ID().String(),
				Limit:     10,
			},
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf(
				"expected not-project-member, got %v",
				err,
			)
		}
	})

	t.Run("propagates list error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
		)

		deps.projectRepo.project = p

		deps.contractRepo.listByProjectFn = func(
			context.Context,
			domaincontract.FilterByProject,
			common.Page,
		) ([]*domaincontract.Contract, error) {
			return nil, errRepository
		}

		_, err := s.ListOverviewByProject(
			context.Background(),
			ListByProjectParams{
				ActorID:   actor.ID().String(),
				ProjectID: p.ID().String(),
				Limit:     10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("propagates count error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
		)

		deps.projectRepo.project = p

		deps.contractRepo.countByProjectFn = func(
			context.Context,
			domaincontract.FilterByProject,
		) (int, error) {
			return 0, errRepository
		}

		_, err := s.ListOverviewByProject(
			context.Background(),
			ListByProjectParams{
				ActorID:   actor.ID().String(),
				ProjectID: p.ID().String(),
				Limit:     10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("maps contracts and signatories", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)
		staff := newTestUser(t, "staff-1", iam.RoleStaff, true)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
			staff.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), staff.ID()},
		)

		deps.projectRepo.project = p

		deps.contractRepo.listByProjectFn = func(
			context.Context,
			domaincontract.FilterByProject,
			common.Page,
		) ([]*domaincontract.Contract, error) {
			return []*domaincontract.Contract{c}, nil
		}

		deps.contractRepo.countByProjectFn = func(
			context.Context,
			domaincontract.FilterByProject,
		) (int, error) {
			return 5, nil
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return []*iam.User{actor, staff}, nil
		}

		result, err := s.ListOverviewByProject(
			context.Background(),
			ListByProjectParams{
				ActorID:   actor.ID().String(),
				ProjectID: p.ID().String(),
				Limit:     10,
				Offset:    0,
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result.TotalCount != 5 {
			t.Fatalf(
				"expected total count 5, got %d",
				result.TotalCount,
			)
		}

		if len(result.Items) != 1 {
			t.Fatalf(
				"expected one item, got %d",
				len(result.Items),
			)
		}

		if len(result.Items[0].Signatories) != 2 {
			t.Fatalf(
				"expected two signatories, got %d",
				len(result.Items[0].Signatories),
			)
		}

		if result.Items[0].MemberSignatoryStatus == "" {
			t.Fatal("expected member signatory status")
		}
	})
}

func TestServiceListOverviewByMember(t *testing.T) {
	t.Run("allows member to list own contracts", func(t *testing.T) {
		s, deps := newTestService()

		member := newTestUser(t, "member-1", iam.RoleClient, true)

		deps.contractRepo.listBySignatoryFn = func(
			context.Context,
			domaincontract.FilterBySignatory,
			common.Page,
		) ([]*domaincontract.Contract, error) {
			return nil, nil
		}

		deps.contractRepo.countBySignatoryFn = func(
			context.Context,
			domaincontract.FilterBySignatory,
		) (int, error) {
			return 3, nil
		}

		result, err := s.ListOverviewByMember(
			context.Background(),
			ListByMemberParams{
				ActorID:  member.ID().String(),
				MemberID: member.ID().String(),
				Limit:    10,
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result.TotalCount != 3 {
			t.Fatalf(
				"expected total count 3, got %d",
				result.TotalCount,
			)
		}
	})

	t.Run("rejects another member for non-administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)
		target := newTestUser(t, "target-1", iam.RoleClient, true)

		deps.userRepo.getByIDFn = func(
			context.Context,
			iam.UserID,
		) (*iam.User, error) {
			return actor, nil
		}

		_, err := s.ListOverviewByMember(
			context.Background(),
			ListByMemberParams{
				ActorID:  actor.ID().String(),
				MemberID: target.ID().String(),
				Limit:    10,
			},
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("propagates list error", func(t *testing.T) {
		s, deps := newTestService()

		member := newTestUser(t, "member-1", iam.RoleClient, true)

		deps.contractRepo.listBySignatoryFn = func(
			context.Context,
			domaincontract.FilterBySignatory,
			common.Page,
		) ([]*domaincontract.Contract, error) {
			return nil, errRepository
		}

		_, err := s.ListOverviewByMember(
			context.Background(),
			ListByMemberParams{
				ActorID:  member.ID().String(),
				MemberID: member.ID().String(),
				Limit:    10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("propagates count error", func(t *testing.T) {
		s, deps := newTestService()

		member := newTestUser(t, "member-1", iam.RoleClient, true)

		deps.contractRepo.countBySignatoryFn = func(
			context.Context,
			domaincontract.FilterBySignatory,
		) (int, error) {
			return 0, errRepository
		}

		_, err := s.ListOverviewByMember(
			context.Background(),
			ListByMemberParams{
				ActorID:  member.ID().String(),
				MemberID: member.ID().String(),
				Limit:    10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestServiceListSignatories(t *testing.T) {
	t.Run("rejects non-signatory", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUserID(t, "actor-1")
		other := newTestUserID(t, "other-1")
		another := newTestUserID(t, "another-1")

		c := newTestContract(
			t,
			"contract-1",
			"project-1",
			[]iam.UserID{other, another},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		_, err := s.ListSignatories(
			context.Background(),
			c.ID().String(),
			actor.String(),
		)

		if !errors.Is(err, domaincontract.ErrContractUserNotSignatory) {
			t.Fatalf(
				"expected not-signatory, got %v",
				err,
			)
		}
	})

	t.Run("propagates user repository error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUserID(t, "actor-1")
		other := newTestUserID(t, "other-1")

		c := newTestContract(
			t,
			"contract-1",
			"project-1",
			[]iam.UserID{actor, other},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return nil, errRepository
		}

		_, err := s.ListSignatories(
			context.Background(),
			c.ID().String(),
			actor.String(),
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("maps available signatories", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient, true)
		other := newTestUser(t, "other-1", iam.RoleStaff, true)

		c := newTestContract(
			t,
			"contract-1",
			"project-1",
			[]iam.UserID{actor.ID(), other.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return []*iam.User{actor, other}, nil
		}

		result, err := s.ListSignatories(
			context.Background(),
			c.ID().String(),
			actor.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if len(result) != 2 {
			t.Fatalf(
				"expected two signatories, got %d",
				len(result),
			)
		}

		if result[0].ID != actor.ID().String() {
			t.Fatalf(
				"expected actor ID %s, got %s",
				actor.ID(),
				result[0].ID,
			)
		}
	})
}

func TestServiceReplaceSignatories(t *testing.T) {
	t.Run("requires administrator or staff", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"actor-1",
			iam.RoleClient,
			true,
		)

		deps.userRepo.getByIDFn = func(
			context.Context,
			iam.UserID,
		) (*iam.User, error) {
			return actor, nil
		}

		err := s.ReplaceSignatories(
			context.Background(),
			ReplaceSignatoriesParams{
				ActorID:     actor.ID().String(),
				ContractID:  "contract-1",
				Signatories: []string{"user-1"},
			},
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("rejects non-project-member actor", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			iam.RoleStaff,
			true,
		)

		member := newTestUser(
			t,
			"member-1",
			iam.RoleClient,
			true,
		)

		other := newTestUser(
			t,
			"other-1",
			iam.RoleClient,
			true,
		)

		p := newTestProject(
			t,
			"project-1",
			member.ID(),
			member.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{member.ID(), other.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			context.Context,
			iam.UserID,
		) (*iam.User, error) {
			return actor, nil
		}

		err := s.ReplaceSignatories(
			context.Background(),
			ReplaceSignatoriesParams{
				ActorID:     actor.ID().String(),
				ContractID:  c.ID().String(),
				Signatories: []string{member.ID().String()},
			},
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf(
				"expected not-project-member, got %v",
				err,
			)
		}
	})

	t.Run("propagates user lookup error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			iam.RoleStaff,
			true,
		)

		client := newTestUser(
			t,
			"client-1",
			iam.RoleClient,
			true,
		)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), client.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			context.Context,
			iam.UserID,
		) (*iam.User, error) {
			return actor, nil
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return nil, errRepository
		}

		err := s.ReplaceSignatories(
			context.Background(),
			ReplaceSignatoriesParams{
				ActorID:     actor.ID().String(),
				ContractID:  c.ID().String(),
				Signatories: []string{actor.ID().String()},
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("replaces signatories and saves", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"staff-1",
			iam.RoleStaff,
			true,
		)

		client := newTestUser(
			t,
			"client-1",
			iam.RoleClient,
			true,
		)

		newStaff := newTestUser(
			t,
			"staff-2",
			iam.RoleStaff,
			true,
		)

		newClient := newTestUser(
			t,
			"client-2",
			iam.RoleClient,
			true,
		)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
			client.ID(),
			newStaff.ID(),
			newClient.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), client.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			context.Context,
			iam.UserID,
		) (*iam.User, error) {
			return actor, nil
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return []*iam.User{newStaff, newClient}, nil
		}

		err := s.ReplaceSignatories(
			context.Background(),
			ReplaceSignatoriesParams{
				ActorID:    actor.ID().String(),
				ContractID: c.ID().String(),
				Signatories: []string{
					newStaff.ID().String(),
					newClient.ID().String(),
				},
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.contractRepo.saveCalls != 1 {
			t.Fatalf(
				"expected one save, got %d",
				deps.contractRepo.saveCalls,
			)
		}

		signatories := c.Signatories()

		if len(signatories) != 2 {
			t.Fatalf(
				"expected two signatories, got %d",
				len(signatories),
			)
		}

		if signatories[0].UserID() != newStaff.ID() {
			t.Fatalf(
				"expected new staff signatory, got %s",
				signatories[0].UserID(),
			)
		}

		if signatories[1].UserID() != newClient.ID() {
			t.Fatalf(
				"expected new client signatory, got %s",
				signatories[1].UserID(),
			)
		}
	})

}

func TestServiceEmailContract(t *testing.T) {
	t.Run("rejects unauthorized actor", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"actor-1",
			iam.RoleClient,
			true,
		)

		member := newTestUser(
			t,
			"member-1",
			iam.RoleClient,
			true,
		)

		deps.userRepo.getByIDFn = func(
			context.Context,
			iam.UserID,
		) (*iam.User, error) {
			return actor, nil
		}

		err := s.EmailContract(
			context.Background(),
			"contract-1",
			actor.ID().String(),
			member.ID().String(),
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("rejects non-project-member", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"actor-1",
			iam.RoleStaff,
			true,
		)

		member := newTestUser(
			t,
			"member-1",
			iam.RoleClient,
			true,
		)

		other := newTestUser(
			t,
			"other-1",
			iam.RoleClient,
			true,
		)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
			other.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), other.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			context.Context,
			iam.UserID,
		) (*iam.User, error) {
			if actor.ID() == iam.UserID(actor.ID()) {
				return actor, nil
			}
			return member, nil
		}

		err := s.EmailContract(
			context.Background(),
			c.ID().String(),
			actor.ID().String(),
			member.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf(
				"expected not-project-member, got %v",
				err,
			)
		}
	})

	t.Run("returns signatory-not-found", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"actor-1",
			iam.RoleStaff,
			true,
		)

		member := newTestUser(
			t,
			"member-1",
			iam.RoleClient,
			true,
		)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
			member.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), member.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			ctx context.Context,
			id iam.UserID,
		) (*iam.User, error) {
			if id == actor.ID() {
				return actor, nil
			}

			return member, nil
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return nil, nil
		}

		err := s.EmailContract(
			context.Background(),
			c.ID().String(),
			actor.ID().String(),
			member.ID().String(),
		)

		if !errors.Is(err, domaincontract.ErrContractSignatoryNotFound) {
			t.Fatalf(
				"expected signatory-not-found, got %v",
				err,
			)
		}
	})

	t.Run("sends contract email", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"actor-1",
			iam.RoleStaff,
			true,
		)

		member := newTestUser(
			t,
			"member-1",
			iam.RoleClient,
			true,
		)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
			member.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), member.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			ctx context.Context,
			id iam.UserID,
		) (*iam.User, error) {
			if id == actor.ID() {
				return actor, nil
			}

			return member, nil
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return []*iam.User{actor, member}, nil
		}

		err := s.EmailContract(
			context.Background(),
			c.ID().String(),
			actor.ID().String(),
			member.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.mailer.sendContractEmailCalls != 1 {
			t.Fatalf(
				"expected one email, got %d",
				deps.mailer.sendContractEmailCalls,
			)
		}

		email := deps.mailer.contractEmail

		if email.Subject != "Contract" {
			t.Fatalf(
				"expected subject Contract, got %s",
				email.Subject,
			)
		}

		if email.RecipientEmail != member.Email().String() {
			t.Fatalf(
				"expected recipient %s, got %s",
				member.Email(),
				email.RecipientEmail,
			)
		}

		if email.ContractID != c.ID().String() {
			t.Fatalf(
				"expected contract ID %s, got %s",
				c.ID(),
				email.ContractID,
			)
		}

		if email.ProjectID != p.ID().String() {
			t.Fatalf(
				"expected project ID %s, got %s",
				p.ID(),
				email.ProjectID,
			)
		}

		if len(email.Signatories) != 2 {
			t.Fatalf(
				"expected two signatories, got %d",
				len(email.Signatories),
			)
		}
	})

	t.Run("propagates mailer error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"actor-1",
			iam.RoleStaff,
			true,
		)

		member := newTestUser(
			t,
			"member-1",
			iam.RoleClient,
			true,
		)

		p := newTestProject(
			t,
			"project-1",
			actor.ID(),
			actor.ID(),
			member.ID(),
		)

		c := newTestContract(
			t,
			"contract-1",
			p.ID().String(),
			[]iam.UserID{actor.ID(), member.ID()},
		)

		deps.contractRepo.getFn = func(
			context.Context,
			domaincontract.ContractID,
		) (*domaincontract.Contract, error) {
			return c, nil
		}

		deps.projectRepo.project = p

		deps.userRepo.getByIDFn = func(
			ctx context.Context,
			id iam.UserID,
		) (*iam.User, error) {
			if id == actor.ID() {
				return actor, nil
			}

			return member, nil
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return []*iam.User{actor, member}, nil
		}

		deps.mailer.sendContractEmailFn = func(
			context.Context,
			mailer.ContractEmail,
		) error {
			return errMail
		}

		err := s.EmailContract(
			context.Background(),
			c.ID().String(),
			actor.ID().String(),
			member.ID().String(),
		)

		if !errors.Is(err, errMail) {
			t.Fatalf(
				"expected mailer error, got %v",
				err,
			)
		}
	})
}
