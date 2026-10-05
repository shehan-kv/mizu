package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/billing"
	domainbilling "mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	domainiam "mizu/internal/domain/iam"
	domainproject "mizu/internal/domain/project"
)

var (
	errBillingRepo   = errors.New("billing repository error")
	errProjectRepo   = errors.New("project repository error")
	errUserRepo      = errors.New("user repository error")
	errIDGenerator   = errors.New("id generator error")
	errMailer        = errors.New("mailer error")
	errEventBus      = errors.New("event bus error")
	errCurrencyRepo  = errors.New("currency repository error")
	errListProjects  = errors.New("list projects error")
	errCountInvoices = errors.New("count invoices error")
)

type fakeBillingUserRepository struct {
	domainiam.UserRepository

	users map[domainiam.UserID]*domainiam.User
	err   error
}

func newFakeBillingUserRepository() *fakeBillingUserRepository {
	return &fakeBillingUserRepository{
		users: make(map[domainiam.UserID]*domainiam.User),
	}
}

func (f *fakeBillingUserRepository) GetByID(
	ctx context.Context,
	id domainiam.UserID,
) (*domainiam.User, error) {
	if f.err != nil {
		return nil, f.err
	}

	user, ok := f.users[id]
	if !ok {
		return nil, errUserRepo
	}

	return user, nil
}

type fakeBillingRepository struct {
	billing.Repository

	invoices map[billing.InvoiceID]*billing.Invoice

	getFn func(
		context.Context,
		billing.InvoiceID,
	) (*billing.Invoice, error)

	addFn func(
		context.Context,
		*billing.Invoice,
	) error

	saveFn func(
		context.Context,
		*billing.Invoice,
	) error

	getCurrencyByCodeFn func(
		context.Context,
		billing.CurrencyCode,
	) (billing.Currency, error)

	listByProjectFn func(
		context.Context,
		billing.FilterByProject,
		common.Page,
	) ([]*billing.Invoice, error)

	countByProjectFn func(
		context.Context,
		billing.FilterByProject,
	) (int, error)

	listByMemberFn func(
		context.Context,
		billing.FilterByMember,
		common.Page,
	) ([]*billing.Invoice, error)

	countByMemberFn func(
		context.Context,
		billing.FilterByMember,
	) (int, error)

	listMonthlyPaidCountByProjectFn func(
		context.Context,
		domainproject.ProjectID,
	) ([]billing.Metric, error)

	listMonthlyPaidCountByMemberFn func(
		context.Context,
		domainiam.UserID,
	) ([]billing.Metric, error)

	getBillingOverviewByMemberFn func(
		context.Context,
		domainiam.UserID,
	) (billing.BillingOverview, error)

	addedInvoice *billing.Invoice
	savedInvoice *billing.Invoice

	getCalls                int
	addCalls                int
	saveCalls               int
	getCurrencyByCodeCalls  int
	listByProjectCalls      int
	countByProjectCalls     int
	listByMemberCalls       int
	countByMemberCalls      int
	listPaidProjectCalls    int
	listPaidMemberCalls     int
	getBillingOverviewCalls int
}

func newFakeBillingRepository() *fakeBillingRepository {
	return &fakeBillingRepository{
		invoices: make(map[billing.InvoiceID]*billing.Invoice),
	}
}

func (f *fakeBillingRepository) Add(
	ctx context.Context,
	invoice *billing.Invoice,
) error {
	f.addCalls++

	if f.addFn != nil {
		return f.addFn(ctx, invoice)
	}

	f.addedInvoice = invoice
	f.invoices[invoice.ID()] = invoice

	return nil
}

func (f *fakeBillingRepository) Get(
	ctx context.Context,
	id billing.InvoiceID,
) (*billing.Invoice, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	invoice, ok := f.invoices[id]
	if !ok {
		return nil, errBillingRepo
	}

	return invoice, nil
}

func (f *fakeBillingRepository) Save(
	ctx context.Context,
	invoice *billing.Invoice,
) error {
	f.saveCalls++

	if f.saveFn != nil {
		return f.saveFn(ctx, invoice)
	}

	f.savedInvoice = invoice
	f.invoices[invoice.ID()] = invoice

	return nil
}

func (f *fakeBillingRepository) GetCurrencyByCode(
	ctx context.Context,
	code billing.CurrencyCode,
) (billing.Currency, error) {
	f.getCurrencyByCodeCalls++

	if f.getCurrencyByCodeFn != nil {
		return f.getCurrencyByCodeFn(ctx, code)
	}

	return newTestBillingCurrency(), nil
}

func (f *fakeBillingRepository) ListByProject(
	ctx context.Context,
	filter billing.FilterByProject,
	page common.Page,
) ([]*billing.Invoice, error) {
	f.listByProjectCalls++

	if f.listByProjectFn != nil {
		return f.listByProjectFn(ctx, filter, page)
	}

	return nil, nil
}

func (f *fakeBillingRepository) CountByProject(
	ctx context.Context,
	filter billing.FilterByProject,
) (int, error) {
	f.countByProjectCalls++

	if f.countByProjectFn != nil {
		return f.countByProjectFn(ctx, filter)
	}

	return 0, nil
}

func (f *fakeBillingRepository) ListByMember(
	ctx context.Context,
	filter billing.FilterByMember,
	page common.Page,
) ([]*billing.Invoice, error) {
	f.listByMemberCalls++

	if f.listByMemberFn != nil {
		return f.listByMemberFn(ctx, filter, page)
	}

	return nil, nil
}

func (f *fakeBillingRepository) CountByMember(
	ctx context.Context,
	filter billing.FilterByMember,
) (int, error) {
	f.countByMemberCalls++

	if f.countByMemberFn != nil {
		return f.countByMemberFn(ctx, filter)
	}

	return 0, nil
}

func (f *fakeBillingRepository) ListMonthlyPaidCountByProject(
	ctx context.Context,
	projectID domainproject.ProjectID,
) ([]billing.Metric, error) {
	f.listPaidProjectCalls++

	if f.listMonthlyPaidCountByProjectFn != nil {
		return f.listMonthlyPaidCountByProjectFn(ctx, projectID)
	}

	return nil, nil
}

func (f *fakeBillingRepository) ListMonthlyPaidCountByMember(
	ctx context.Context,
	memberID domainiam.UserID,
) ([]billing.Metric, error) {
	f.listPaidMemberCalls++

	if f.listMonthlyPaidCountByMemberFn != nil {
		return f.listMonthlyPaidCountByMemberFn(ctx, memberID)
	}

	return nil, nil
}

func (f *fakeBillingRepository) GetBillingOverviewByMember(
	ctx context.Context,
	memberID domainiam.UserID,
) (billing.BillingOverview, error) {
	f.getBillingOverviewCalls++

	if f.getBillingOverviewByMemberFn != nil {
		return f.getBillingOverviewByMemberFn(ctx, memberID)
	}

	return billing.BillingOverview{}, nil
}

type fakeBillingProjectRepository struct {
	domainproject.Repository

	projects map[domainproject.ProjectID]*domainproject.Project

	getFn func(
		context.Context,
		domainproject.ProjectID,
	) (*domainproject.Project, error)

	listByIDsFn func(
		context.Context,
		[]domainproject.ProjectID,
	) ([]*domainproject.Project, error)

	getCalls       int
	listByIDsCalls int
}

func newFakeBillingProjectRepository() *fakeBillingProjectRepository {
	return &fakeBillingProjectRepository{
		projects: make(map[domainproject.ProjectID]*domainproject.Project),
	}
}

func (f *fakeBillingProjectRepository) Get(
	ctx context.Context,
	id domainproject.ProjectID,
) (*domainproject.Project, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	project, ok := f.projects[id]
	if !ok {
		return nil, errProjectRepo
	}

	return project, nil
}

func (f *fakeBillingProjectRepository) ListByIDs(
	ctx context.Context,
	ids []domainproject.ProjectID,
) ([]*domainproject.Project, error) {
	f.listByIDsCalls++

	if f.listByIDsFn != nil {
		return f.listByIDsFn(ctx, ids)
	}

	projects := make([]*domainproject.Project, 0, len(ids))

	for _, id := range ids {
		if p, ok := f.projects[id]; ok {
			projects = append(projects, p)
		}
	}

	return projects, nil
}

type fakeBillingBus struct {
	events []common.Event
	err    error
}

func (f *fakeBillingBus) Publish(
	ctx context.Context,
	event common.Event,
) error {
	if f.err != nil {
		return f.err
	}

	f.events = append(f.events, event)
	return nil
}

func (f *fakeBillingBus) Subscribe(
	eventType common.EventType,
	handler eventbus.EventHandler,
) {
}

type fakeBillingIDGenerator struct {
	id    string
	err   error
	calls int
}

func (f *fakeBillingIDGenerator) Generate() (string, error) {
	f.calls++

	if f.err != nil {
		return "", f.err
	}

	return f.id, nil
}

type fakeBillingLogger struct {
	warnings []string
}

func (f *fakeBillingLogger) Info(msg string, args ...any) {}

func (f *fakeBillingLogger) Warn(msg string, args ...any) {
	f.warnings = append(f.warnings, msg)
}

func (f *fakeBillingLogger) Error(msg string, args ...any) {}

func (f *fakeBillingLogger) Fatal(msg string, args ...any) {}

type fakeBillingMailer struct {
	email *mailer.InvoiceEmail
	err   error
	calls int
}

func (f *fakeBillingMailer) SendRecoveryEmail(
	ctx context.Context,
	email domainiam.Email,
	token domainiam.RecoveryToken,
) error {
	return nil
}

func (f *fakeBillingMailer) SendVerificationEmail(
	ctx context.Context,
	email domainiam.Email,
	verificationID domainiam.VerificationID,
) error {
	return nil
}

func (f *fakeBillingMailer) SendVerifiedEmail(
	ctx context.Context,
	email domainiam.Email,
) error {
	return nil
}

func (f *fakeBillingMailer) SendContractEmail(
	ctx context.Context,
	email mailer.ContractEmail,
) error {
	return nil
}

func (f *fakeBillingMailer) SendInvoiceEmail(
	ctx context.Context,
	email mailer.InvoiceEmail,
) error {
	f.calls++
	f.email = &email

	return f.err
}

func newTestBillingCurrency() billing.Currency {
	name, err := billing.NewCurrencyName("US Dollar")
	if err != nil {
		panic(err)
	}

	symbol, err := billing.NewCurrencySymbol("$")
	if err != nil {
		panic(err)
	}

	code, err := billing.NewCurrencyCode("USD")
	if err != nil {
		panic(err)
	}

	decimals, err := billing.NewCurrencyDecimals(2)
	if err != nil {
		panic(err)
	}

	return billing.NewCurrency(
		name,
		symbol,
		code,
		decimals,
	)
}

func newTestBillingItem(t *testing.T) billing.Item {
	t.Helper()

	qty, err := billing.NewQty("2")
	if err != nil {
		t.Fatalf("failed to create quantity: %v", err)
	}

	unitPrice, err := billing.NewDecimal("100")
	if err != nil {
		t.Fatalf("failed to create unit price: %v", err)
	}

	discountRate, err := billing.NewDecimal("10")
	if err != nil {
		t.Fatalf("failed to create discount rate: %v", err)
	}

	taxRate, err := billing.NewDecimal("5")
	if err != nil {
		t.Fatalf("failed to create tax rate: %v", err)
	}

	item, err := billing.NewItem(
		"Test item",
		qty,
		unitPrice,
		discountRate,
		billing.DiscountTypePercentage,
		taxRate,
		billing.TaxTypePercentage,
	)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	return item
}

func newTestInvoice(t *testing.T, id string, projectID domainproject.ProjectID, isInvoice bool) *billing.Invoice {
	t.Helper()

	invoiceID, err := billing.NewInvoiceID(id)
	if err != nil {
		t.Fatalf("failed to create invoice ID: %v", err)
	}

	now := time.Now()
	currency := newTestBillingCurrency()
	item := newTestBillingItem(t)

	note := "Test invoice"

	invoice, err := billing.NewInvoice(
		invoiceID,
		projectID,
		isInvoice,
		nil,
		currency,
		&note,
		[]billing.Item{item},
		now,
	)
	if err != nil {
		t.Fatalf("failed to create invoice: %v", err)
	}

	return invoice
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

func newTestBillingService(
	userRepo *fakeBillingUserRepository,
	billingRepo *fakeBillingRepository,
	projectRepo *fakeBillingProjectRepository,
	bus *fakeBillingBus,
	idGen *fakeBillingIDGenerator,
	mailerRepo *fakeBillingMailer,
	loggerRepo *fakeBillingLogger,
) *Service {
	authzService := authz.NewService(userRepo)

	return NewService(
		userRepo,
		billingRepo,
		projectRepo,
		authzService,
		bus,
		idGen,
		mailerRepo,
		loggerRepo,
	)
}

func TestService_CreateInvoice(t *testing.T) {
	t.Run("creates invoice", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()
		bus := &fakeBillingBus{}
		idGen := &fakeBillingIDGenerator{id: "invoice-1"}
		mailerRepo := &fakeBillingMailer{}
		loggerRepo := &fakeBillingLogger{}

		actor := newTestUser(t, "actor-1", domainiam.RoleStaff, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			bus,
			idGen,
			mailerRepo,
			loggerRepo,
		)

		note := "Test note"

		err := service.CreateInvoice(context.Background(), CreateInvoiceParams{
			ActorID:      actor.ID().String(),
			ProjectID:    project.ID().String(),
			CurrencyCode: "usd",
			Note:         &note,
			IsInvoice:    true,
			Items: []CreateItemParams{
				{
					Description:  "Development",
					Qty:          "2",
					UnitPrice:    "100",
					DiscountRate: "10",
					DiscountType: "percentage",
					TaxRate:      "5",
					TaxType:      "percentage",
				},
			},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if billingRepo.addedInvoice == nil {
			t.Fatal("expected invoice to be added")
		}

		if billingRepo.addedInvoice.ID().String() != "invoice-1" {
			t.Fatalf(
				"expected invoice ID invoice-1, got %s",
				billingRepo.addedInvoice.ID(),
			)
		}

		if billingRepo.addedInvoice.ProjectID() != project.ID() {
			t.Fatalf("unexpected project ID")
		}

		if !billingRepo.addedInvoice.IsInvoice() {
			t.Fatal("expected invoice to be an invoice")
		}

		if billingRepo.addedInvoice.Status() != billing.StatusPending {
			t.Fatalf("expected pending status, got %s", billingRepo.addedInvoice.Status())
		}

		if len(billingRepo.addedInvoice.Items()) != 1 {
			t.Fatalf("expected 1 item, got %d", len(billingRepo.addedInvoice.Items()))
		}

		if len(bus.events) != 1 {
			t.Fatalf("expected 1 published event, got %d", len(bus.events))
		}
	})

	t.Run("rejects invalid actor ID", func(t *testing.T) {
		service := newTestBillingService(
			newFakeBillingUserRepository(),
			newFakeBillingRepository(),
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.CreateInvoice(context.Background(), CreateInvoiceParams{
			ActorID:      "",
			ProjectID:    "project-1",
			CurrencyCode: "USD",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects non administrator or staff", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.CreateInvoice(context.Background(), CreateInvoiceParams{
			ActorID:      actor.ID().String(),
			ProjectID:    project.ID().String(),
			CurrencyCode: "USD",
		})
		if err == nil {
			t.Fatal("expected authorization error")
		}
	})

	t.Run("rejects non-project-member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestAdmin(t, "actor-1")
		creator := newTestUser(t, "creator-1", domainiam.RoleStaff, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[creator.ID()] = creator

		project := newTestProject(t, "project-1", creator.ID())
		projectRepo.projects[project.ID()] = project

		service := newTestBillingService(
			userRepo,
			newFakeBillingRepository(),
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.CreateInvoice(context.Background(), CreateInvoiceParams{
			ActorID:      actor.ID().String(),
			ProjectID:    project.ID().String(),
			CurrencyCode: "USD",
		})
		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected ErrNotProjectMember, got %v", err)
		}
	})

	t.Run("propagates currency repository error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestAdmin(t, "actor-1")
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		billingRepo.getCurrencyByCodeFn = func(
			ctx context.Context,
			code billing.CurrencyCode,
		) (billing.Currency, error) {
			return billing.Currency{}, errCurrencyRepo
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.CreateInvoice(context.Background(), CreateInvoiceParams{
			ActorID:      actor.ID().String(),
			ProjectID:    project.ID().String(),
			CurrencyCode: "USD",
			Items: []CreateItemParams{
				{
					Description:  "Development",
					Qty:          "1",
					UnitPrice:    "100",
					DiscountRate: "0",
					DiscountType: "fixed",
					TaxRate:      "0",
					TaxType:      "fixed",
				},
			},
		})
		if !errors.Is(err, errCurrencyRepo) {
			t.Fatalf("expected currency error, got %v", err)
		}
	})

	t.Run("propagates ID generator error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestAdmin(t, "actor-1")
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{err: errIDGenerator},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.CreateInvoice(context.Background(), CreateInvoiceParams{
			ActorID:      actor.ID().String(),
			ProjectID:    project.ID().String(),
			CurrencyCode: "USD",
			Items: []CreateItemParams{
				{
					Description:  "Development",
					Qty:          "1",
					UnitPrice:    "100",
					DiscountRate: "0",
					DiscountType: "fixed",
					TaxRate:      "0",
					TaxType:      "fixed",
				},
			},
		})
		if !errors.Is(err, errIDGenerator) {
			t.Fatalf("expected ID generator error, got %v", err)
		}
	})

	t.Run("propagates repository add error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestAdmin(t, "actor-1")
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		billingRepo.addFn = func(
			ctx context.Context,
			invoice *billing.Invoice,
		) error {
			return errBillingRepo
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.CreateInvoice(context.Background(), CreateInvoiceParams{
			ActorID:      actor.ID().String(),
			ProjectID:    project.ID().String(),
			CurrencyCode: "USD",
			Items: []CreateItemParams{
				{
					Description:  "Development",
					Qty:          "1",
					UnitPrice:    "100",
					DiscountRate: "0",
					DiscountType: "fixed",
					TaxRate:      "0",
					TaxType:      "fixed",
				},
			},
		})
		if !errors.Is(err, errBillingRepo) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestService_ListInvoicesByProject(t *testing.T) {
	t.Run("returns invoices", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)

		billingRepo.listByProjectFn = func(
			ctx context.Context,
			filter billing.FilterByProject,
			page common.Page,
		) ([]*billing.Invoice, error) {
			return []*billing.Invoice{invoice}, nil
		}

		billingRepo.countByProjectFn = func(
			ctx context.Context,
			filter billing.FilterByProject,
		) (int, error) {
			return 1, nil
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		result, err := service.ListInvoicesByProject(
			context.Background(),
			ListInvoicesByProjectParams{
				ActorID:   actor.ID().String(),
				ProjectID: project.ID().String(),
				Limit:     10,
				Offset:    0,
			},
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.TotalCount != 1 {
			t.Fatalf("expected total count 1, got %d", result.TotalCount)
		}

		if len(result.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result.Items))
		}

		got := result.Items[0]

		if got.ID != invoice.ID().String() {
			t.Fatalf("expected invoice ID %s, got %s", invoice.ID(), got.ID)
		}

		if got.ProjectID != project.ID().String() {
			t.Fatalf("unexpected project ID")
		}

		if got.ProjectName != project.Name().String() {
			t.Fatalf("unexpected project name")
		}

		if got.CurrencyCode != "USD" {
			t.Fatalf("expected USD, got %s", got.CurrencyCode)
		}

		if got.Status != billing.StatusPending.String() {
			t.Fatalf("unexpected status")
		}
	})

	t.Run("rejects non-project-member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		project := newTestProject(t, "project-1", member.ID())
		projectRepo.projects[project.ID()] = project

		service := newTestBillingService(
			userRepo,
			newFakeBillingRepository(),
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListInvoicesByProject(
			context.Background(),
			ListInvoicesByProjectParams{
				ActorID:   actor.ID().String(),
				ProjectID: project.ID().String(),
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected ErrNotProjectMember, got %v", err)
		}
	})

	t.Run("propagates list error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		billingRepo.listByProjectFn = func(
			ctx context.Context,
			filter billing.FilterByProject,
			page common.Page,
		) ([]*billing.Invoice, error) {
			return nil, errBillingRepo
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListInvoicesByProject(
			context.Background(),
			ListInvoicesByProjectParams{
				ActorID:   actor.ID().String(),
				ProjectID: project.ID().String(),
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errBillingRepo) {
			t.Fatalf("expected billing repository error, got %v", err)
		}
	})

	t.Run("propagates count error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)

		billingRepo.listByProjectFn = func(
			ctx context.Context,
			filter billing.FilterByProject,
			page common.Page,
		) ([]*billing.Invoice, error) {
			return []*billing.Invoice{invoice}, nil
		}

		billingRepo.countByProjectFn = func(
			ctx context.Context,
			filter billing.FilterByProject,
		) (int, error) {
			return 0, errCountInvoices
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListInvoicesByProject(
			context.Background(),
			ListInvoicesByProjectParams{
				ActorID:   actor.ID().String(),
				ProjectID: project.ID().String(),
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errCountInvoices) {
			t.Fatalf("expected count error, got %v", err)
		}
	})
}

func TestService_ListInvoicesByMember(t *testing.T) {
	t.Run("returns invoices for self", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		member := newTestUser(t, "member-1", domainiam.RoleClient, true)
		userRepo.users[member.ID()] = member

		project := newTestProject(t, "project-1", member.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)

		billingRepo.listByMemberFn = func(
			ctx context.Context,
			filter billing.FilterByMember,
			page common.Page,
		) ([]*billing.Invoice, error) {
			return []*billing.Invoice{invoice}, nil
		}

		billingRepo.countByMemberFn = func(
			ctx context.Context,
			filter billing.FilterByMember,
		) (int, error) {
			return 1, nil
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		result, err := service.ListInvoicesByMember(
			context.Background(),
			ListInvoicesByMemberParams{
				ActorID:  member.ID().String(),
				MemberID: member.ID().String(),
				Limit:    10,
				Offset:   0,
			},
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.TotalCount != 1 {
			t.Fatalf("expected count 1, got %d", result.TotalCount)
		}

		if len(result.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result.Items))
		}

		if result.Items[0].ProjectName != project.Name().String() {
			t.Fatalf("unexpected project name")
		}
	})

	t.Run("allows administrator to list another member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[admin.ID()] = admin
		userRepo.users[member.ID()] = member

		billingRepo.listByMemberFn = func(
			ctx context.Context,
			filter billing.FilterByMember,
			page common.Page,
		) ([]*billing.Invoice, error) {
			return nil, nil
		}

		billingRepo.countByMemberFn = func(
			ctx context.Context,
			filter billing.FilterByMember,
		) (int, error) {
			return 0, nil
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListInvoicesByMember(
			context.Background(),
			ListInvoicesByMemberParams{
				ActorID:  admin.ID().String(),
				MemberID: member.ID().String(),
				Limit:    10,
				Offset:   0,
			},
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects unauthorized member access", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		service := newTestBillingService(
			userRepo,
			newFakeBillingRepository(),
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListInvoicesByMember(
			context.Background(),
			ListInvoicesByMemberParams{
				ActorID:  actor.ID().String(),
				MemberID: member.ID().String(),
			},
		)

		if err == nil {
			t.Fatal("expected authorization error")
		}
	})
}

func TestService_GetInvoice(t *testing.T) {
	t.Run("returns invoice DTO", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		dto, err := service.GetInvoice(
			context.Background(),
			actor.ID().String(),
			invoice.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if dto.ID != invoice.ID().String() {
			t.Fatalf("unexpected ID")
		}

		if dto.ProjectID != project.ID().String() {
			t.Fatalf("unexpected project ID")
		}

		if dto.ProjectName != project.Name().String() {
			t.Fatalf("unexpected project name")
		}

		if dto.IsInvoice != invoice.IsInvoice() {
			t.Fatalf("unexpected IsInvoice")
		}

		if dto.Status != invoice.Status().String() {
			t.Fatalf("unexpected status")
		}

		if dto.CurrencyCode != invoice.Currency().Code().String() {
			t.Fatalf("unexpected currency")
		}

		if dto.SubTotal != invoice.SubTotal().String() {
			t.Fatalf("unexpected subtotal")
		}

		if dto.TotalTax != invoice.TotalTax().String() {
			t.Fatalf("unexpected total tax")
		}

		if dto.TotalDiscount != invoice.TotalDiscount().String() {
			t.Fatalf("unexpected total discount")
		}

		if len(dto.Items) != len(invoice.Items()) {
			t.Fatalf("expected %d items, got %d", len(invoice.Items()), len(dto.Items))
		}

		item := invoice.Items()[0]
		dtoItem := dto.Items[0]

		if dtoItem.Description != item.Description() {
			t.Fatalf("unexpected description")
		}

		if dtoItem.Qty != item.Qty().String() {
			t.Fatalf("unexpected quantity")
		}

		if dtoItem.UnitPrice != item.UnitPrice().String() {
			t.Fatalf("unexpected unit price")
		}

		if dtoItem.DiscountRate != item.DiscountRate().String() {
			t.Fatalf("unexpected discount rate")
		}

		if dtoItem.DiscountType != item.DiscountType().String() {
			t.Fatalf("unexpected discount type")
		}

		if dtoItem.TaxRate != item.TaxRate().String() {
			t.Fatalf("unexpected tax rate")
		}

		if dtoItem.TaxType != item.TaxType().String() {
			t.Fatalf("unexpected tax type")
		}

		if dtoItem.LineGross != item.LineGross().String() {
			t.Fatalf("unexpected line gross")
		}

		if dtoItem.LineDiscount != item.LineDiscount().String() {
			t.Fatalf("unexpected line discount")
		}

		if dtoItem.LineNet != item.LineNet().String() {
			t.Fatalf("unexpected line net")
		}

		if dtoItem.LineTax != item.LineTax().String() {
			t.Fatalf("unexpected line tax")
		}

		if dtoItem.LineTotal != item.LineTotal().String() {
			t.Fatalf("unexpected line total")
		}
	})

	t.Run("rejects non-project-member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		project := newTestProject(t, "project-1", member.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.GetInvoice(
			context.Background(),
			actor.ID().String(),
			invoice.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected ErrNotProjectMember, got %v", err)
		}
	})

	t.Run("propagates invoice repository error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		userRepo.users[actor.ID()] = actor

		billingRepo.getFn = func(
			ctx context.Context,
			id billing.InvoiceID,
		) (*billing.Invoice, error) {
			return nil, errBillingRepo
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.GetInvoice(
			context.Background(),
			actor.ID().String(),
			"invoice-1",
		)

		if !errors.Is(err, errBillingRepo) {
			t.Fatalf("expected billing repository error, got %v", err)
		}
	})
}

func TestService_EmailInvoice(t *testing.T) {
	t.Run("sends invoice email", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()
		mailerRepo := &fakeBillingMailer{}

		actor := newTestAdmin(t, "actor-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		project := newTestProject(
			t,
			"project-1",
			actor.ID(),
			member.ID(),
		)
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-12345678", project.ID(), true)
		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			mailerRepo,
			&fakeBillingLogger{},
		)

		err := service.EmailInvoice(
			context.Background(),
			invoice.ID().String(),
			actor.ID().String(),
			member.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if mailerRepo.calls != 1 {
			t.Fatalf("expected 1 mailer call, got %d", mailerRepo.calls)
		}

		if mailerRepo.email == nil {
			t.Fatal("expected email")
		}

		if mailerRepo.email.RecipientEmail != member.Email().String() {
			t.Fatalf("unexpected recipient")
		}

		if mailerRepo.email.InvoiceID != invoice.ID().String() {
			t.Fatalf("unexpected invoice ID")
		}

		if mailerRepo.email.ProjectID != project.ID().String() {
			t.Fatalf("unexpected project ID")
		}

		if mailerRepo.email.Status != invoice.Status().String() {
			t.Fatalf("unexpected status")
		}

		if mailerRepo.email.CurrencyName != invoice.Currency().Name().String() {
			t.Fatalf("unexpected currency name")
		}

		if mailerRepo.email.CurrencyCode != invoice.Currency().Code().String() {
			t.Fatalf("unexpected currency code")
		}

		if mailerRepo.email.SubTotal != invoice.SubTotal().String() {
			t.Fatalf("unexpected subtotal")
		}

		if mailerRepo.email.TotalTax != invoice.TotalTax().String() {
			t.Fatalf("unexpected tax")
		}

		if mailerRepo.email.TotalDiscount != invoice.TotalDiscount().String() {
			t.Fatalf("unexpected discount")
		}

		if len(mailerRepo.email.Items) != 1 {
			t.Fatalf("expected 1 email item, got %d", len(mailerRepo.email.Items))
		}
	})

	t.Run("uses quote subject for quote", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()
		mailerRepo := &fakeBillingMailer{}

		actor := newTestAdmin(t, "actor-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		project := newTestProject(t, "project-1", actor.ID(), member.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "quote-12345678", project.ID(), false)
		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			mailerRepo,
			&fakeBillingLogger{},
		)

		err := service.EmailInvoice(
			context.Background(),
			invoice.ID().String(),
			actor.ID().String(),
			member.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if mailerRepo.email.Subject != "Quote: #12345678" {
			t.Fatalf(
				"expected quote subject, got %q",
				mailerRepo.email.Subject,
			)
		}
	})

	t.Run("rejects actor who is not authorized for member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		project := newTestProject(t, "project-1", actor.ID(), member.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		billingRepo.invoices[invoice.ID()] = invoice

		mailerRepo := &fakeBillingMailer{}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			mailerRepo,
			&fakeBillingLogger{},
		)

		err := service.EmailInvoice(
			context.Background(),
			invoice.ID().String(),
			actor.ID().String(),
			member.ID().String(),
		)
		if err == nil {
			t.Fatal("expected authorization error")
		}

		if mailerRepo.calls != 0 {
			t.Fatal("expected mailer not to be called")
		}
	})

	t.Run("rejects non-project-member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestAdmin(t, "actor-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)
		other := newTestUser(t, "other-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member
		userRepo.users[other.ID()] = other

		project := newTestProject(t, "project-1", actor.ID(), member.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.EmailInvoice(
			context.Background(),
			invoice.ID().String(),
			actor.ID().String(),
			other.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected ErrNotProjectMember, got %v", err)
		}
	})

	t.Run("propagates mailer error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()
		mailerRepo := &fakeBillingMailer{err: errMailer}

		actor := newTestAdmin(t, "actor-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		project := newTestProject(t, "project-1", actor.ID(), member.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			mailerRepo,
			&fakeBillingLogger{},
		)

		err := service.EmailInvoice(
			context.Background(),
			invoice.ID().String(),
			actor.ID().String(),
			member.ID().String(),
		)

		if !errors.Is(err, errMailer) {
			t.Fatalf("expected mailer error, got %v", err)
		}
	})
}

func TestService_AcceptInvoice(t *testing.T) {
	testInvoiceMutation(
		t,
		"AcceptInvoice",
		domainiam.RoleClient,
		func(
			service *Service,
			ctx context.Context,
			actorID string,
			invoiceID string,
		) error {
			return service.AcceptInvoice(ctx, actorID, invoiceID)
		},
		billing.StatusAccepted,
	)
}

func TestService_RejectInvoice(t *testing.T) {
	testInvoiceMutation(
		t,
		"RejectInvoice",
		domainiam.RoleClient,
		func(
			service *Service,
			ctx context.Context,
			actorID string,
			invoiceID string,
		) error {
			return service.RejectInvoice(ctx, actorID, invoiceID)
		},
		billing.StatusRejected,
	)
}

func TestService_PayInvoice(t *testing.T) {
	t.Run("pays accepted invoice", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()
		bus := &fakeBillingBus{}

		actor := newTestAdmin(t, "actor-1")
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		if err := invoice.Accept(time.Now()); err != nil {
			t.Fatalf("failed to prepare invoice: %v", err)
		}
		invoice.PullEvents()

		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			bus,
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.PayInvoice(
			context.Background(),
			actor.ID().String(),
			invoice.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if billingRepo.savedInvoice != invoice {
			t.Fatal("expected invoice to be saved")
		}

		if invoice.Status() != billing.StatusPaid {
			t.Fatalf("expected paid status, got %s", invoice.Status())
		}

		if len(bus.events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(bus.events))
		}
	})
}

func TestService_CancelInvoice(t *testing.T) {
	testInvoiceMutation(
		t,
		"CancelInvoice",
		domainiam.RoleStaff,
		func(
			service *Service,
			ctx context.Context,
			actorID string,
			invoiceID string,
		) error {
			return service.CancelInvoice(ctx, actorID, invoiceID)
		},
		billing.StatusCancelled,
	)
}

func TestService_ConvertToInvoice(t *testing.T) {
	t.Run("converts quote to invoice", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()
		bus := &fakeBillingBus{}

		actor := newTestAdmin(t, "actor-1")
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "quote-1", project.ID(), false)
		invoice.PullEvents()

		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			bus,
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.ConvertToInvoice(
			context.Background(),
			actor.ID().String(),
			invoice.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !invoice.IsInvoice() {
			t.Fatal("expected quote to become invoice")
		}

		if billingRepo.savedInvoice != invoice {
			t.Fatal("expected invoice to be saved")
		}

		if len(bus.events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(bus.events))
		}
	})

	t.Run("rejects already converted invoice", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestAdmin(t, "actor-1")
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		invoice.PullEvents()

		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := service.ConvertToInvoice(
			context.Background(),
			actor.ID().String(),
			invoice.ID().String(),
		)
		if err == nil {
			t.Fatal("expected error")
		}

		if billingRepo.saveCalls != 0 {
			t.Fatal("expected invoice not to be saved")
		}
	})
}

func TestService_ListPaidCountByProject(t *testing.T) {
	t.Run("returns metrics", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		billingRepo.listMonthlyPaidCountByProjectFn = func(
			ctx context.Context,
			projectID domainproject.ProjectID,
		) ([]billing.Metric, error) {
			return []billing.Metric{
				billing.NewMetric("2026-01", 3),
				billing.NewMetric("2026-02", 5),
			}, nil
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		result, err := service.ListPaidCountByProject(
			context.Background(),
			actor.ID().String(),
			project.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 2 metrics, got %d", len(result))
		}

		if result[0].Key != "2026-01" || result[0].Value != int64(3) {
			t.Fatalf("unexpected first metric: %+v", result[0])
		}

		if result[1].Key != "2026-02" || result[1].Value != int64(5) {
			t.Fatalf("unexpected second metric: %+v", result[1])
		}
	})

	t.Run("rejects non-project-member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		project := newTestProject(t, "project-1", member.ID())
		projectRepo.projects[project.ID()] = project

		service := newTestBillingService(
			userRepo,
			newFakeBillingRepository(),
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListPaidCountByProject(
			context.Background(),
			actor.ID().String(),
			project.ID().String(),
		)

		if !errors.Is(err, domainproject.ErrNotProjectMember) {
			t.Fatalf("expected ErrNotProjectMember, got %v", err)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		billingRepo.listMonthlyPaidCountByProjectFn = func(
			ctx context.Context,
			projectID domainproject.ProjectID,
		) ([]billing.Metric, error) {
			return nil, errBillingRepo
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListPaidCountByProject(
			context.Background(),
			actor.ID().String(),
			project.ID().String(),
		)

		if !errors.Is(err, errBillingRepo) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestService_ListPaidCountByMember(t *testing.T) {
	t.Run("returns metrics for self", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()

		member := newTestUser(t, "member-1", domainiam.RoleClient, true)
		userRepo.users[member.ID()] = member

		billingRepo.listMonthlyPaidCountByMemberFn = func(
			ctx context.Context,
			memberID domainiam.UserID,
		) ([]billing.Metric, error) {
			return []billing.Metric{
				billing.NewMetric("2026-01", 4),
				billing.NewMetric("2026-02", 7),
			}, nil
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		result, err := service.ListPaidCountByMember(
			context.Background(),
			member.ID().String(),
			member.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 2 metrics, got %d", len(result))
		}

		if result[0].Key != "2026-01" || result[0].Value != int64(4) {
			t.Fatalf("unexpected first metric: %+v", result[0])
		}

		if result[1].Key != "2026-02" || result[1].Value != int64(7) {
			t.Fatalf("unexpected second metric: %+v", result[1])
		}
	})

	t.Run("allows administrator to access another member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[admin.ID()] = admin
		userRepo.users[member.ID()] = member

		service := newTestBillingService(
			userRepo,
			billingRepo,
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListPaidCountByMember(
			context.Background(),
			admin.ID().String(),
			member.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects unauthorized member access", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		service := newTestBillingService(
			userRepo,
			newFakeBillingRepository(),
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.ListPaidCountByMember(
			context.Background(),
			actor.ID().String(),
			member.ID().String(),
		)

		if err == nil {
			t.Fatal("expected authorization error")
		}
	})
}

func TestService_GetBillingSummaryByMember(t *testing.T) {
	t.Run("maps all overview metrics", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()

		member := newTestUser(t, "member-1", domainiam.RoleClient, true)
		userRepo.users[member.ID()] = member

		usd, err := billing.NewCurrencyCode("USD")
		if err != nil {
			t.Fatalf("failed to create USD code: %v", err)
		}

		eur, err := billing.NewCurrencyCode("EUR")
		if err != nil {
			t.Fatalf("failed to create EUR code: %v", err)
		}

		amount100, err := billing.NewDecimal("100")
		if err != nil {
			t.Fatalf("failed to create amount: %v", err)
		}

		amount200, err := billing.NewDecimal("200")
		if err != nil {
			t.Fatalf("failed to create amount: %v", err)
		}

		overview := billing.NewBillingOverview(
			[]billing.BillingOverviewMetric{
				billing.NewBillingOverviewMetric(usd, amount100, 2),
			},
			[]billing.BillingOverviewMetric{
				billing.NewBillingOverviewMetric(eur, amount200, 3),
			},
			[]billing.BillingOverviewMetric{
				billing.NewBillingOverviewMetric(usd, amount200, 4),
			},
			[]billing.BillingOverviewMetric{
				billing.NewBillingOverviewMetric(eur, amount100, 5),
			},
			[]billing.BillingOverviewMetric{
				billing.NewBillingOverviewMetric(usd, amount100, 6),
			},
			[]billing.BillingOverviewMetric{
				billing.NewBillingOverviewMetric(eur, amount200, 7),
			},
			[]billing.BillingOverviewMetric{
				billing.NewBillingOverviewMetric(usd, amount200, 8),
			},
		)

		billingRepo.getBillingOverviewByMemberFn = func(
			ctx context.Context,
			memberID domainiam.UserID,
		) (billing.BillingOverview, error) {
			return overview, nil
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		result, err := service.GetBillingSummaryByMember(
			context.Background(),
			member.ID().String(),
			member.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(result.InvoicesPaid) != 1 {
			t.Fatalf("expected one paid metric")
		}

		if result.InvoicesPaid[0].CurrencyCode != "USD" ||
			result.InvoicesPaid[0].Amount != "100" ||
			result.InvoicesPaid[0].Count != 2 {
			t.Fatalf("unexpected paid metric: %+v", result.InvoicesPaid[0])
		}

		if len(result.InvoicesPending) != 1 ||
			result.InvoicesPending[0].CurrencyCode != "EUR" ||
			result.InvoicesPending[0].Amount != "200" ||
			result.InvoicesPending[0].Count != 3 {
			t.Fatalf("unexpected pending metric: %+v", result.InvoicesPending)
		}

		if len(result.InvoicesAccepted) != 1 ||
			result.InvoicesAccepted[0].Count != 4 {
			t.Fatalf("unexpected accepted metric: %+v", result.InvoicesAccepted)
		}

		if len(result.InvoicesRejected) != 1 ||
			result.InvoicesRejected[0].Count != 5 {
			t.Fatalf("unexpected rejected metric: %+v", result.InvoicesRejected)
		}

		if len(result.InvoicesCancelled) != 1 ||
			result.InvoicesCancelled[0].Count != 6 {
			t.Fatalf("unexpected cancelled metric: %+v", result.InvoicesCancelled)
		}

		if len(result.QuotesPending) != 1 ||
			result.QuotesPending[0].Count != 7 {
			t.Fatalf("unexpected pending quote metric: %+v", result.QuotesPending)
		}

		if len(result.QuotesRejected) != 1 ||
			result.QuotesRejected[0].Count != 8 {
			t.Fatalf("unexpected rejected quote metric: %+v", result.QuotesRejected)
		}
	})

	t.Run("allows administrator to access another member", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()

		admin := newTestAdmin(t, "admin-1")
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[admin.ID()] = admin
		userRepo.users[member.ID()] = member

		service := newTestBillingService(
			userRepo,
			billingRepo,
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.GetBillingSummaryByMember(
			context.Background(),
			admin.ID().String(),
			member.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects unauthorized member access", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()

		actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
		member := newTestUser(t, "member-1", domainiam.RoleClient, true)

		userRepo.users[actor.ID()] = actor
		userRepo.users[member.ID()] = member

		service := newTestBillingService(
			userRepo,
			newFakeBillingRepository(),
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.GetBillingSummaryByMember(
			context.Background(),
			actor.ID().String(),
			member.ID().String(),
		)

		if err == nil {
			t.Fatal("expected authorization error")
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()

		member := newTestUser(t, "member-1", domainiam.RoleClient, true)
		userRepo.users[member.ID()] = member

		billingRepo.getBillingOverviewByMemberFn = func(
			ctx context.Context,
			memberID domainiam.UserID,
		) (billing.BillingOverview, error) {
			return billing.BillingOverview{}, errBillingRepo
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			newFakeBillingProjectRepository(),
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-1"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		_, err := service.GetBillingSummaryByMember(
			context.Background(),
			member.ID().String(),
			member.ID().String(),
		)

		if !errors.Is(err, errBillingRepo) {
			t.Fatalf("expected billing repository error, got %v", err)
		}
	})
}

func TestService_EventPublishingFailureDoesNotFailMutation(t *testing.T) {
	userRepo := newFakeBillingUserRepository()
	billingRepo := newFakeBillingRepository()
	projectRepo := newFakeBillingProjectRepository()
	bus := &fakeBillingBus{err: errEventBus}

	actor := newTestUser(t, "actor-1", domainiam.RoleClient, true)
	userRepo.users[actor.ID()] = actor

	project := newTestProject(t, "project-1", actor.ID())
	projectRepo.projects[project.ID()] = project

	invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
	invoice.PullEvents()
	billingRepo.invoices[invoice.ID()] = invoice

	logger := &fakeBillingLogger{}

	service := newTestBillingService(
		userRepo,
		billingRepo,
		projectRepo,
		bus,
		&fakeBillingIDGenerator{id: "invoice-2"},
		&fakeBillingMailer{},
		logger,
	)

	err := service.AcceptInvoice(
		context.Background(),
		actor.ID().String(),
		invoice.ID().String(),
	)
	if err != nil {
		t.Fatalf(
			"expected mutation to succeed despite event publishing failure, got %v",
			err,
		)
	}

	if invoice.Status() != domainbilling.StatusAccepted {
		t.Fatalf(
			"expected invoice status to be accepted, got %s",
			invoice.Status(),
		)
	}
}

func testInvoiceMutation(
	t *testing.T,
	name string,
	role domainiam.Role,
	mutate func(
		service *Service,
		ctx context.Context,
		actorID string,
		invoiceID string,
	) error,
	expectedStatus billing.Status,
) {
	t.Helper()

	t.Run(name+" succeeds", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()
		bus := &fakeBillingBus{}

		actor := newTestUser(t, "actor-1", role, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		invoice.PullEvents()

		billingRepo.invoices[invoice.ID()] = invoice

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			bus,
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := mutate(
			service,
			context.Background(),
			actor.ID().String(),
			invoice.ID().String(),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if invoice.Status() != expectedStatus {
			t.Fatalf(
				"expected status %s, got %s",
				expectedStatus,
				invoice.Status(),
			)
		}

		if billingRepo.savedInvoice != invoice {
			t.Fatal("expected invoice to be saved")
		}

		if len(bus.events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(bus.events))
		}
	})

	t.Run(name+" propagates repository get error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", role, true)
		userRepo.users[actor.ID()] = actor

		billingRepo.getFn = func(
			ctx context.Context,
			id billing.InvoiceID,
		) (*billing.Invoice, error) {
			return nil, errBillingRepo
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := mutate(
			service,
			context.Background(),
			actor.ID().String(),
			"invoice-1",
		)
		if !errors.Is(err, errBillingRepo) {
			t.Fatalf("expected billing repository error, got %v", err)
		}
	})

	t.Run(name+" propagates save error", func(t *testing.T) {
		userRepo := newFakeBillingUserRepository()
		billingRepo := newFakeBillingRepository()
		projectRepo := newFakeBillingProjectRepository()

		actor := newTestUser(t, "actor-1", role, true)
		userRepo.users[actor.ID()] = actor

		project := newTestProject(t, "project-1", actor.ID())
		projectRepo.projects[project.ID()] = project

		invoice := newTestInvoice(t, "invoice-1", project.ID(), true)
		invoice.PullEvents()

		billingRepo.invoices[invoice.ID()] = invoice
		billingRepo.saveFn = func(
			ctx context.Context,
			invoice *billing.Invoice,
		) error {
			return errBillingRepo
		}

		service := newTestBillingService(
			userRepo,
			billingRepo,
			projectRepo,
			&fakeBillingBus{},
			&fakeBillingIDGenerator{id: "invoice-2"},
			&fakeBillingMailer{},
			&fakeBillingLogger{},
		)

		err := mutate(
			service,
			context.Background(),
			actor.ID().String(),
			invoice.ID().String(),
		)
		if !errors.Is(err, errBillingRepo) {
			t.Fatalf("expected billing repository error, got %v", err)
		}
	})
}
