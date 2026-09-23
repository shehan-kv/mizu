package iam

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"mizu/internal/application/authz"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/filestore"
	"mizu/internal/application/logger"
	"mizu/internal/application/session"
	"mizu/internal/application/uow"
	"mizu/internal/domain/common"
	domainiam "mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

var (
	errUnexpectedCall = errors.New("unexpected call")
	errRepository     = errors.New("repository failure")
	errUOW            = errors.New("unit of work failure")
	errIDGenerator    = errors.New("id generator failure")
	errHasher         = errors.New("password hasher failure")
	errSessionStore   = errors.New("session store failure")
	errFileStore      = errors.New("file store failure")
	errEventBus       = errors.New("event bus failure")
)

type fakeUserRepository struct {
	users map[domainiam.UserID]*domainiam.User

	hasAdministratorFn func(context.Context) (bool, error)
	getByIDFn          func(context.Context, domainiam.UserID) (*domainiam.User, error)
	getByEmailFn       func(context.Context, domainiam.Email) (*domainiam.User, error)
	listFn             func(context.Context, domainiam.UserFilter, common.Page) ([]*domainiam.User, error)
	countFn            func(context.Context, domainiam.UserFilter) (int, error)

	isAnyAdministratorFn func(context.Context, []domainiam.UserID) (bool, error)

	addFn    func(context.Context, *domainiam.User) error
	saveFn   func(context.Context, *domainiam.User) error
	removeFn func(context.Context, *domainiam.User) error

	addCalls    int
	saveCalls   int
	removeCalls int
}

var _ domainiam.UserRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) Add(
	ctx context.Context,
	user *domainiam.User,
) error {
	f.addCalls++

	if f.addFn != nil {
		return f.addFn(ctx, user)
	}

	if f.users == nil {
		f.users = make(map[domainiam.UserID]*domainiam.User)
	}

	f.users[user.ID()] = user
	return nil
}

func (f *fakeUserRepository) Exists(
	context.Context,
	domainiam.UserID,
) (bool, error) {
	panic("unexpected call to Exists")
}

func (f *fakeUserRepository) ExistsAll(
	context.Context,
	[]domainiam.UserID,
) (bool, error) {
	panic("unexpected call to ExistsAll")
}

func (f *fakeUserRepository) GetByID(
	ctx context.Context,
	id domainiam.UserID,
) (*domainiam.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	if f.users != nil {
		user, ok := f.users[id]
		if ok {
			return user, nil
		}
	}

	return nil, domainiam.ErrUserNotFound
}

func (f *fakeUserRepository) GetByEmail(
	ctx context.Context,
	email domainiam.Email,
) (*domainiam.User, error) {
	if f.getByEmailFn != nil {
		return f.getByEmailFn(ctx, email)
	}

	if f.users != nil {
		for _, user := range f.users {
			if user.Email() == email {
				return user, nil
			}
		}
	}

	return nil, domainiam.ErrUserNotFound
}

func (f *fakeUserRepository) List(
	ctx context.Context,
	filter domainiam.UserFilter,
	page common.Page,
) ([]*domainiam.User, error) {
	if f.listFn != nil {
		return f.listFn(ctx, filter, page)
	}

	return nil, nil
}

func (f *fakeUserRepository) ListByIDs(
	context.Context,
	[]domainiam.UserID,
	domainiam.UserFilter,
) ([]*domainiam.User, error) {
	panic("unexpected call to ListByIDs")
}

func (f *fakeUserRepository) IsAnyAdministrator(
	ctx context.Context,
	ids []domainiam.UserID,
) (bool, error) {
	if f.isAnyAdministratorFn != nil {
		return f.isAnyAdministratorFn(ctx, ids)
	}

	for _, id := range ids {
		user, ok := f.users[id]
		if ok && user.IsAdministrator() {
			return true, nil
		}
	}

	return false, nil
}

func (f *fakeUserRepository) HasAdministrator(
	ctx context.Context,
) (bool, error) {
	if f.hasAdministratorFn != nil {
		return f.hasAdministratorFn(ctx)
	}

	for _, user := range f.users {
		if user.IsAdministrator() {
			return true, nil
		}
	}

	return false, nil
}

func (f *fakeUserRepository) Count(
	ctx context.Context,
	filter domainiam.UserFilter,
) (int, error) {
	if f.countFn != nil {
		return f.countFn(ctx, filter)
	}

	return 0, nil
}

func (f *fakeUserRepository) Save(
	ctx context.Context,
	user *domainiam.User,
) error {
	f.saveCalls++

	if f.saveFn != nil {
		return f.saveFn(ctx, user)
	}

	if f.users == nil {
		f.users = make(map[domainiam.UserID]*domainiam.User)
	}

	f.users[user.ID()] = user
	return nil
}

func (f *fakeUserRepository) Remove(
	ctx context.Context,
	user *domainiam.User,
) error {
	f.removeCalls++

	if f.removeFn != nil {
		return f.removeFn(ctx, user)
	}

	delete(f.users, user.ID())
	return nil
}

type fakeCredentialRepository struct {
	credential *domainiam.Credential

	getByUserFn func(context.Context, domainiam.UserID) (*domainiam.Credential, error)
	addFn       func(context.Context, *domainiam.Credential) error
	saveFn      func(context.Context, *domainiam.Credential) error

	addCalls  int
	saveCalls int
}

var _ domainiam.CredentialRepository = (*fakeCredentialRepository)(nil)

func (f *fakeCredentialRepository) Add(
	ctx context.Context,
	credential *domainiam.Credential,
) error {
	f.addCalls++

	if f.addFn != nil {
		return f.addFn(ctx, credential)
	}

	f.credential = credential
	return nil
}

func (f *fakeCredentialRepository) GetByUser(
	ctx context.Context,
	userID domainiam.UserID,
) (*domainiam.Credential, error) {
	if f.getByUserFn != nil {
		return f.getByUserFn(ctx, userID)
	}

	if f.credential == nil {
		return nil, domainiam.ErrCredentialNotFound
	}

	return f.credential, nil
}

func (f *fakeCredentialRepository) Save(
	ctx context.Context,
	credential *domainiam.Credential,
) error {
	f.saveCalls++

	if f.saveFn != nil {
		return f.saveFn(ctx, credential)
	}

	f.credential = credential
	return nil
}

type fakeRecoveryRepository struct {
	recovery *domainiam.Recovery

	getByTokenFn   func(context.Context, domainiam.RecoveryToken) (*domainiam.Recovery, error)
	getByUserFn    func(context.Context, domainiam.UserID) (*domainiam.Recovery, error)
	addFn          func(context.Context, *domainiam.Recovery) error
	saveFn         func(context.Context, *domainiam.Recovery) error
	removeFn       func(context.Context, *domainiam.Recovery) error
	removeByUserFn func(context.Context, domainiam.UserID) error

	addCalls          int
	saveCalls         int
	removeCalls       int
	removeByUserCalls int
}

var _ domainiam.RecoveryRepository = (*fakeRecoveryRepository)(nil)

func (f *fakeRecoveryRepository) Add(
	ctx context.Context,
	recovery *domainiam.Recovery,
) error {
	f.addCalls++

	if f.addFn != nil {
		return f.addFn(ctx, recovery)
	}

	f.recovery = recovery
	return nil
}

func (f *fakeRecoveryRepository) GetByToken(
	ctx context.Context,
	token domainiam.RecoveryToken,
) (*domainiam.Recovery, error) {
	if f.getByTokenFn != nil {
		return f.getByTokenFn(ctx, token)
	}

	if f.recovery == nil {
		return nil, domainiam.ErrRecoveryNotFound
	}

	return f.recovery, nil
}

func (f *fakeRecoveryRepository) GetByUser(
	ctx context.Context,
	userID domainiam.UserID,
) (*domainiam.Recovery, error) {
	if f.getByUserFn != nil {
		return f.getByUserFn(ctx, userID)
	}

	if f.recovery == nil {
		return nil, domainiam.ErrRecoveryNotFound
	}

	return f.recovery, nil
}

func (f *fakeRecoveryRepository) Save(
	ctx context.Context,
	recovery *domainiam.Recovery,
) error {
	f.saveCalls++

	if f.saveFn != nil {
		return f.saveFn(ctx, recovery)
	}

	f.recovery = recovery
	return nil
}

func (f *fakeRecoveryRepository) Remove(
	ctx context.Context,
	recovery *domainiam.Recovery,
) error {
	f.removeCalls++

	if f.removeFn != nil {
		return f.removeFn(ctx, recovery)
	}

	if f.recovery == recovery {
		f.recovery = nil
	}

	return nil
}

func (f *fakeRecoveryRepository) RemoveByUser(
	ctx context.Context,
	userID domainiam.UserID,
) error {
	f.removeByUserCalls++

	if f.removeByUserFn != nil {
		return f.removeByUserFn(ctx, userID)
	}

	if f.recovery != nil && f.recovery.UserID() == userID {
		f.recovery = nil
	}

	return nil
}

type fakeVerificationRepository struct {
	verification *domainiam.Verification

	getFn            func(context.Context, domainiam.VerificationID) (*domainiam.Verification, error)
	addFn            func(context.Context, *domainiam.Verification) error
	removeFn         func(context.Context, *domainiam.Verification) error
	removeByUserIDFn func(context.Context, domainiam.UserID) error

	addCalls          int
	removeCalls       int
	removeByUserCalls int
}

var _ domainiam.VerificationRepository = (*fakeVerificationRepository)(nil)

func (f *fakeVerificationRepository) Add(
	ctx context.Context,
	verification *domainiam.Verification,
) error {
	f.addCalls++

	if f.addFn != nil {
		return f.addFn(ctx, verification)
	}

	f.verification = verification
	return nil
}

func (f *fakeVerificationRepository) Get(
	ctx context.Context,
	id domainiam.VerificationID,
) (*domainiam.Verification, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	if f.verification == nil {
		return nil, domainiam.ErrVerificationNotFound
	}

	return f.verification, nil
}

func (f *fakeVerificationRepository) Remove(
	ctx context.Context,
	verification *domainiam.Verification,
) error {
	f.removeCalls++

	if f.removeFn != nil {
		return f.removeFn(ctx, verification)
	}

	if f.verification == verification {
		f.verification = nil
	}

	return nil
}

func (f *fakeVerificationRepository) RemoveByUserID(
	ctx context.Context,
	userID domainiam.UserID,
) error {
	f.removeByUserCalls++

	if f.removeByUserIDFn != nil {
		return f.removeByUserIDFn(ctx, userID)
	}

	if f.verification != nil && f.verification.UserID() == userID {
		f.verification = nil
	}

	return nil
}

type fakeProjectRepository struct {
	projects []*project.Project

	listByIDsFn func(context.Context, []project.ProjectID) ([]*project.Project, error)
	saveAllFn   func(context.Context, []*project.Project) error

	saveAllCalls int
}

var _ project.Repository = (*fakeProjectRepository)(nil)

func (f *fakeProjectRepository) Add(
	context.Context,
	*project.Project,
) error {
	panic("unexpected call to ProjectRepository.Add")
}

func (f *fakeProjectRepository) Get(
	context.Context,
	project.ProjectID,
) (*project.Project, error) {
	panic("unexpected call to ProjectRepository.Get")
}

func (f *fakeProjectRepository) List(
	context.Context,
	project.Filter,
	common.Page,
) ([]*project.Project, error) {
	panic("unexpected call to ProjectRepository.List")
}

func (f *fakeProjectRepository) ListByIDs(
	ctx context.Context,
	ids []project.ProjectID,
) ([]*project.Project, error) {
	if f.listByIDsFn != nil {
		return f.listByIDsFn(ctx, ids)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	return f.projects, nil
}

func (f *fakeProjectRepository) ListByMember(
	context.Context,
	domainiam.UserID,
) ([]*project.Project, error) {
	panic("unexpected call to ProjectRepository.ListByMember")
}

func (f *fakeProjectRepository) ListCreatedPerDay(
	context.Context,
	domainiam.UserID,
) ([]project.Metric, error) {
	panic("unexpected call to ProjectRepository.ListCreatedPerDay")
}

func (f *fakeProjectRepository) Count(
	context.Context,
	project.Filter,
) (int, error) {
	panic("unexpected call to ProjectRepository.Count")
}

func (f *fakeProjectRepository) Save(
	context.Context,
	*project.Project,
) error {
	panic("unexpected call to ProjectRepository.Save")
}

func (f *fakeProjectRepository) SaveAll(
	ctx context.Context,
	projects []*project.Project,
) error {
	f.saveAllCalls++

	if f.saveAllFn != nil {
		return f.saveAllFn(ctx, projects)
	}

	f.projects = projects
	return nil
}

func (f *fakeProjectRepository) Remove(
	context.Context,
	*project.Project,
) error {
	panic("unexpected call to ProjectRepository.Remove")
}

type fakePasswordHasher struct {
	hashFn   func(domainiam.PlainPassword) (string, error)
	verifyFn func(string, domainiam.PlainPassword) bool

	hashCalls int
}

var _ domainiam.PasswordHasher = (*fakePasswordHasher)(nil)

func (f *fakePasswordHasher) Hash(
	password domainiam.PlainPassword,
) (string, error) {
	f.hashCalls++

	if f.hashFn != nil {
		return f.hashFn(password)
	}

	return "hashed-password", nil
}

func (f *fakePasswordHasher) Verify(
	hash string,
	password domainiam.PlainPassword,
) bool {
	if f.verifyFn != nil {
		return f.verifyFn(hash, password)
	}

	return true
}

type fakeIDGenerator struct {
	ids []string
	err error

	calls int
}

var _ common.IDGenerator = (*fakeIDGenerator)(nil)

func (f *fakeIDGenerator) Generate() (string, error) {
	f.calls++

	if f.err != nil {
		return "", f.err
	}

	if len(f.ids) == 0 {
		return "generated-id", nil
	}

	id := f.ids[0]
	f.ids = f.ids[1:]

	return id, nil
}

type fakeUnitOfWork struct {
	executeFn func(context.Context, func(context.Context) error) error

	calls int
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

type fakeSessionStore struct {
	sessions map[session.SessionID]*session.Session

	addFn    func(context.Context, *session.Session) error
	deleteFn func(context.Context, session.SessionID) error
	getFn    func(context.Context, session.SessionID) (*session.Session, error)
	listFn   func(context.Context, domainiam.UserID) ([]*session.Session, error)

	addCalls    int
	deleteCalls int
}

var _ session.Store = (*fakeSessionStore)(nil)

func (f *fakeSessionStore) Add(
	ctx context.Context,
	s *session.Session,
) error {
	f.addCalls++

	if f.addFn != nil {
		return f.addFn(ctx, s)
	}

	if f.sessions == nil {
		f.sessions = make(map[session.SessionID]*session.Session)
	}

	f.sessions[s.ID()] = s
	return nil
}

func (f *fakeSessionStore) Get(
	ctx context.Context,
	id session.SessionID,
) (*session.Session, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	if s, ok := f.sessions[id]; ok {
		return s, nil
	}

	return nil, errors.New("session not found")
}

func (f *fakeSessionStore) ListByUser(
	ctx context.Context,
	userID domainiam.UserID,
) ([]*session.Session, error) {
	if f.listFn != nil {
		return f.listFn(ctx, userID)
	}

	return nil, nil
}

func (f *fakeSessionStore) Delete(
	ctx context.Context,
	id session.SessionID,
) error {
	f.deleteCalls++

	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}

	delete(f.sessions, id)
	return nil
}

type fakeLogger struct {
	infoCalls  int
	warnCalls  int
	errorCalls int
	fatalCalls int
}

var _ logger.Logger = (*fakeLogger)(nil)

func (f *fakeLogger) Info(string, ...any) {
	f.infoCalls++
}

func (f *fakeLogger) Warn(string, ...any) {
	f.warnCalls++
}

func (f *fakeLogger) Error(string, ...any) {
	f.errorCalls++
}

func (f *fakeLogger) Fatal(string, ...any) {
	f.fatalCalls++
}

type fakeInternalBus struct {
	publishFn func(context.Context, common.Event) error

	events []common.Event
	calls  int
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

type fakeFileStore struct {
	saveFn   func(context.Context, string, io.Reader) error
	openFn   func(context.Context, string) (io.ReadCloser, error)
	deleteFn func(context.Context, string) error

	saveCalls   int
	openCalls   int
	deleteCalls int

	savedKey   string
	deletedKey string
}

var _ filestore.Store = (*fakeFileStore)(nil)

func (f *fakeFileStore) Save(
	ctx context.Context,
	key string,
	reader io.Reader,
) error {
	f.saveCalls++
	f.savedKey = key

	if f.saveFn != nil {
		return f.saveFn(ctx, key, reader)
	}

	return nil
}

func (f *fakeFileStore) Open(
	ctx context.Context,
	key string,
) (io.ReadCloser, error) {
	f.openCalls++

	if f.openFn != nil {
		return f.openFn(ctx, key)
	}

	return io.NopCloser(strings.NewReader("file")), nil
}

func (f *fakeFileStore) Delete(
	ctx context.Context,
	key string,
) error {
	f.deleteCalls++
	f.deletedKey = key

	if f.deleteFn != nil {
		return f.deleteFn(ctx, key)
	}

	return nil
}

type testServiceDeps struct {
	userRepo         *fakeUserRepository
	credRepo         *fakeCredentialRepository
	recoveryRepo     *fakeRecoveryRepository
	verificationRepo *fakeVerificationRepository
	projectRepo      *fakeProjectRepository
	uow              *fakeUnitOfWork
	hasher           *fakePasswordHasher
	idGen            *fakeIDGenerator
	sessionStore     *fakeSessionStore
	logger           *fakeLogger
	bus              *fakeInternalBus
	fileStore        *fakeFileStore
}

func newTestService() (*Service, *testServiceDeps) {
	userRepo := &fakeUserRepository{
		users: make(map[domainiam.UserID]*domainiam.User),
	}

	credRepo := &fakeCredentialRepository{}
	recoveryRepo := &fakeRecoveryRepository{}
	verificationRepo := &fakeVerificationRepository{}
	projectRepo := &fakeProjectRepository{}
	uow := &fakeUnitOfWork{}
	hasher := &fakePasswordHasher{}
	idGen := &fakeIDGenerator{}
	sessionStore := &fakeSessionStore{
		sessions: make(map[session.SessionID]*session.Session),
	}
	logger := &fakeLogger{}
	bus := &fakeInternalBus{}
	fileStore := &fakeFileStore{}

	domainIAMService := domainiam.NewService(
		hasher,
		credRepo,
	)

	authzService := authz.NewService(userRepo)

	projectService := project.NewService(
		userRepo,
		projectRepo,
	)

	service := NewService(
		userRepo,
		credRepo,
		recoveryRepo,
		verificationRepo,
		projectRepo,
		uow,
		domainIAMService,
		authzService,
		projectService,
		bus,
		fileStore,
		idGen,
		sessionStore,
		logger,
		nil,
		hasher,
	)

	return service, &testServiceDeps{
		userRepo:         userRepo,
		credRepo:         credRepo,
		recoveryRepo:     recoveryRepo,
		verificationRepo: verificationRepo,
		projectRepo:      projectRepo,
		uow:              uow,
		hasher:           hasher,
		idGen:            idGen,
		sessionStore:     sessionStore,
		logger:           logger,
		bus:              bus,
		fileStore:        fileStore,
	}
}

func newTestUser(
	t *testing.T,
	id string,
	role domainiam.Role,
	active bool,
) *domainiam.User {
	t.Helper()

	now := time.Now()

	userID, err := domainiam.NewUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	name, err := domainiam.NewName("Test", "User")
	if err != nil {
		t.Fatal(err)
	}

	email, err := domainiam.NewEmail(id + "@example.com")
	if err != nil {
		t.Fatal(err)
	}

	return domainiam.NewSystemUser(
		userID,
		name,
		email,
		nil,
		role,
		active,
		now,
	)
}

func newTestAdmin(t *testing.T, id string) *domainiam.User {
	t.Helper()

	return newTestUser(
		t,
		id,
		domainiam.RoleAdministrator,
		true,
	)
}

func newTestUserID(t *testing.T, value string) domainiam.UserID {
	t.Helper()

	id, err := domainiam.NewUserID(value)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func newTestEmail(t *testing.T, value string) domainiam.Email {
	t.Helper()

	email, err := domainiam.NewEmail(value)
	if err != nil {
		t.Fatal(err)
	}

	return email
}

func newTestVerification(
	t *testing.T,
	userID domainiam.UserID,
	id string,
) *domainiam.Verification {
	t.Helper()

	verificationID, err := domainiam.NewVerificationID(id)
	if err != nil {
		t.Fatal(err)
	}

	return domainiam.NewVerification(
		verificationID,
		userID,
		time.Now(),
	)
}

func newTestRecovery(
	t *testing.T,
	userID domainiam.UserID,
	token string,
) *domainiam.Recovery {
	t.Helper()

	recoveryToken, err := domainiam.NewRecoveryToken(token)
	if err != nil {
		t.Fatal(err)
	}

	return domainiam.NewRecovery(
		userID,
		recoveryToken,
		time.Now(),
	)
}

func TestServiceEnsureDefaultAdminExists(t *testing.T) {
	t.Run("returns nil when administrator already exists", func(t *testing.T) {
		s, deps := newTestService()

		deps.userRepo.hasAdministratorFn = func(context.Context) (bool, error) {
			return true, nil
		}

		err := s.EnsureDefaultAdminExists(context.Background())

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.idGen.calls != 0 {
			t.Fatalf("expected ID generator not to be called, got %d calls", deps.idGen.calls)
		}

		if deps.uow.calls != 0 {
			t.Fatalf("expected UOW not to be called, got %d calls", deps.uow.calls)
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.userRepo.hasAdministratorFn = func(context.Context) (bool, error) {
			return false, errRepository
		}

		err := s.EnsureDefaultAdminExists(context.Background())

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("creates default administrator and credentials", func(t *testing.T) {
		s, deps := newTestService()

		deps.userRepo.hasAdministratorFn = func(context.Context) (bool, error) {
			return false, nil
		}

		deps.idGen.ids = []string{"admin-id"}

		err := s.EnsureDefaultAdminExists(context.Background())

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.uow.calls != 1 {
			t.Fatalf("expected one UOW execution, got %d", deps.uow.calls)
		}

		if deps.userRepo.addCalls != 1 {
			t.Fatalf("expected one user add, got %d", deps.userRepo.addCalls)
		}

		if deps.credRepo.addCalls != 1 {
			t.Fatalf("expected one credential add, got %d", deps.credRepo.addCalls)
		}

		user, err := deps.userRepo.GetByID(
			context.Background(),
			newTestUserID(t, "admin-id"),
		)
		if err != nil {
			t.Fatalf("expected created user, got %v", err)
		}

		if !user.IsAdministrator() {
			t.Fatal("expected created user to be administrator")
		}

		if deps.credRepo.credential == nil {
			t.Fatal("expected credential to be created")
		}
	})
}

func TestServiceSignIn(t *testing.T) {
	t.Run("returns validation error for invalid email", func(t *testing.T) {
		s, _ := newTestService()

		_, err := s.SignIn(
			context.Background(),
			"invalid",
			"password",
			nil,
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("returns validation error for invalid password", func(t *testing.T) {
		s, _ := newTestService()

		_, err := s.SignIn(
			context.Background(),
			"user@example.com",
			"",
			nil,
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.userRepo.getByEmailFn = func(
			context.Context,
			domainiam.Email,
		) (*domainiam.User, error) {
			return nil, errRepository
		}

		_, err := s.SignIn(
			context.Background(),
			"user@example.com",
			"password",
			nil,
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("creates session", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")

		deps.userRepo.getByEmailFn = func(
			context.Context,
			domainiam.Email,
		) (*domainiam.User, error) {
			return user, nil
		}

		deps.credRepo.credential = domainiam.NewCredential(
			user.ID(),
			"hashed-password",
		)

		deps.idGen.ids = []string{"session-1"}

		before := time.Now()

		result, err := s.SignIn(
			context.Background(),
			user.Email().String(),
			"password",
			nil,
		)

		after := time.Now()

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result.ID != "session-1" {
			t.Fatalf("expected session ID session-1, got %s", result.ID)
		}

		if result.UserID != user.ID().String() {
			t.Fatalf("expected user ID %s, got %s",
				user.ID(),
				result.UserID,
			)
		}

		if result.Role != user.Role().String() {
			t.Fatalf("expected role %s, got %s",
				user.Role(),
				result.Role,
			)
		}

		if result.IssuedAt.Before(before) || result.IssuedAt.After(after) {
			t.Fatal("session issued time is outside expected range")
		}

		if deps.sessionStore.addCalls != 1 {
			t.Fatalf("expected one session add, got %d", deps.sessionStore.addCalls)
		}

		if deps.userRepo.saveCalls != 1 {
			t.Fatalf("expected sign-in to save user, got %d", deps.userRepo.saveCalls)
		}
	})

	t.Run("revokes external session before creating new session", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")

		deps.userRepo.getByEmailFn = func(
			context.Context,
			domainiam.Email,
		) (*domainiam.User, error) {
			return user, nil
		}

		deps.credRepo.credential = domainiam.NewCredential(
			user.ID(),
			"hashed-password",
		)

		deps.idGen.ids = []string{"session-2"}

		_, err := s.SignIn(
			context.Background(),
			user.Email().String(),
			"password",
			func() *string {
				value := "old-session"
				return &value
			}(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.sessionStore.deleteCalls != 1 {
			t.Fatalf(
				"expected one old session deletion, got %d",
				deps.sessionStore.deleteCalls,
			)
		}

		if deps.sessionStore.addCalls != 1 {
			t.Fatalf("expected one session add, got %d", deps.sessionStore.addCalls)
		}
	})

	t.Run("continues when old session deletion fails", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")

		deps.userRepo.getByEmailFn = func(
			context.Context,
			domainiam.Email,
		) (*domainiam.User, error) {
			return user, nil
		}

		deps.credRepo.credential = domainiam.NewCredential(
			user.ID(),
			"hashed-password",
		)

		deps.idGen.ids = []string{"session-3"}

		deps.sessionStore.deleteFn = func(
			context.Context,
			session.SessionID,
		) error {
			return errSessionStore
		}

		oldSession := "old-session"

		_, err := s.SignIn(
			context.Background(),
			user.Email().String(),
			"password",
			&oldSession,
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.logger.errorCalls != 1 {
			t.Fatalf(
				"expected one logged error, got %d",
				deps.logger.errorCalls,
			)
		}

		if deps.sessionStore.addCalls != 1 {
			t.Fatalf("expected session to still be added")
		}
	})

	t.Run("continues when recording sign-in fails", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")

		deps.userRepo.getByEmailFn = func(
			context.Context,
			domainiam.Email,
		) (*domainiam.User, error) {
			return user, nil
		}

		deps.credRepo.credential = domainiam.NewCredential(
			user.ID(),
			"hashed-password",
		)

		deps.idGen.ids = []string{"session-4"}

		deps.userRepo.saveFn = func(
			context.Context,
			*domainiam.User,
		) error {
			return errRepository
		}

		_, err := s.SignIn(
			context.Background(),
			user.Email().String(),
			"password",
			nil,
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.logger.errorCalls != 1 {
			t.Fatalf("expected one logged error, got %d",
				deps.logger.errorCalls)
		}
	})
}

func TestServiceSignOut(t *testing.T) {
	t.Run("ignores invalid session ID", func(t *testing.T) {
		s, deps := newTestService()

		err := s.SignOut(context.Background(), "")

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.sessionStore.deleteCalls != 0 {
			t.Fatal("expected invalid session ID not to reach store")
		}
	})

	t.Run("deletes session", func(t *testing.T) {
		s, deps := newTestService()

		err := s.SignOut(context.Background(), "session-1")

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.sessionStore.deleteCalls != 1 {
			t.Fatalf("expected one delete, got %d",
				deps.sessionStore.deleteCalls)
		}
	})

	t.Run("propagates store error", func(t *testing.T) {
		s, deps := newTestService()

		deps.sessionStore.deleteFn = func(
			context.Context,
			session.SessionID,
		) error {
			return errSessionStore
		}

		err := s.SignOut(context.Background(), "session-1")

		if !errors.Is(err, errSessionStore) {
			t.Fatalf("expected session-store error, got %v", err)
		}
	})
}

func TestServiceCreateUser(t *testing.T) {
	t.Run("requires administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "not-admin", domainiam.RoleClient, true)
		deps.userRepo.users[actor.ID()] = actor

		err := s.CreateUser(context.Background(), CreateUserParams{
			ActorID:   actor.ID().String(),
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Role:      domainiam.RoleClient.String(),
		})

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("creates user and verification", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		deps.userRepo.users[admin.ID()] = admin

		deps.idGen.ids = []string{
			"user-1",
			"verification-1",
		}

		title := "Developer"

		err := s.CreateUser(
			context.Background(),
			CreateUserParams{
				ActorID:   admin.ID().String(),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Title:     &title,
				Role:      domainiam.RoleClient.String(),
				IsActive:  true,
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.userRepo.addCalls != 1 {
			t.Fatalf("expected one user add, got %d", deps.userRepo.addCalls)
		}

		if deps.verificationRepo.addCalls != 1 {
			t.Fatalf(
				"expected one verification add, got %d",
				deps.verificationRepo.addCalls,
			)
		}

		if deps.projectRepo.saveAllCalls != 1 {
			t.Fatalf(
				"expected one project SaveAll, got %d",
				deps.projectRepo.saveAllCalls,
			)
		}

		if deps.uow.calls != 1 {
			t.Fatalf("expected one UOW call, got %d", deps.uow.calls)
		}

		if deps.bus.calls != 2 {
			t.Fatalf(
				"expected two published events, got %d",
				deps.bus.calls,
			)
		}
	})

	t.Run("propagates project assignment error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		deps.userRepo.users[admin.ID()] = admin

		deps.idGen.ids = []string{
			"user-1",
			"verification-1",
		}

		deps.projectRepo.listByIDsFn = func(
			context.Context,
			[]project.ProjectID,
		) ([]*project.Project, error) {
			return nil, errRepository
		}

		err := s.CreateUser(
			context.Background(),
			CreateUserParams{
				ActorID:   admin.ID().String(),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Role:      domainiam.RoleClient.String(),
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}

		if deps.uow.calls != 0 {
			t.Fatal("expected UOW not to run")
		}
	})

	t.Run("does not publish events when UOW fails", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		deps.userRepo.users[admin.ID()] = admin

		deps.idGen.ids = []string{
			"user-1",
			"verification-1",
		}

		deps.uow.executeFn = func(
			context.Context,
			func(context.Context) error,
		) error {
			return errUOW
		}

		err := s.CreateUser(
			context.Background(),
			CreateUserParams{
				ActorID:   admin.ID().String(),
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Role:      domainiam.RoleClient.String(),
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

func TestServiceRegenerateVerification(t *testing.T) {
	t.Run("returns error for verified user", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		target := newTestUser(
			t,
			"target-1",
			domainiam.RoleClient,
			true,
		)

		target.Verify(time.Now())

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[target.ID()] = target

		err := s.RegenerateVerification(
			context.Background(),
			target.ID().String(),
			admin.ID().String(),
		)

		if !errors.Is(err, domainiam.ErrUserAlreadyVerified) {
			t.Fatalf("expected already-verified error, got %v", err)
		}

		if deps.verificationRepo.removeByUserCalls != 0 {
			t.Fatalf("expected no verification removal, got %d",
				deps.verificationRepo.removeByUserCalls)
		}

		if deps.verificationRepo.addCalls != 0 {
			t.Fatalf("expected no verification add, got %d",
				deps.verificationRepo.addCalls)
		}

		if deps.bus.calls != 0 {
			t.Fatalf("expected no events, got %d", deps.bus.calls)
		}
	})

	t.Run("replaces verification", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")

		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[user.ID()] = user

		deps.idGen.ids = []string{"verification-2"}

		err := s.RegenerateVerification(
			context.Background(),
			user.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.verificationRepo.removeByUserCalls != 1 {
			t.Fatalf(
				"expected one old verification removal, got %d",
				deps.verificationRepo.removeByUserCalls,
			)
		}

		if deps.verificationRepo.addCalls != 1 {
			t.Fatalf(
				"expected one verification add, got %d",
				deps.verificationRepo.addCalls,
			)
		}

		if deps.bus.calls != 1 {
			t.Fatalf("expected one event, got %d", deps.bus.calls)
		}
	})
}

func TestServiceVerifyAccount(t *testing.T) {
	t.Run("returns error for already verified user", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		user.Verify(time.Now())

		verification := newTestVerification(
			t,
			user.ID(),
			"verification-1",
		)

		deps.verificationRepo.verification = verification
		deps.userRepo.users[user.ID()] = user

		err := s.VerifyAccount(
			context.Background(),
			"verification-1",
			"password",
		)

		if !errors.Is(err, domainiam.ErrUserAlreadyVerified) {
			t.Fatalf("expected already-verified error, got %v", err)
		}

		if deps.userRepo.saveCalls != 0 {
			t.Fatalf("expected no user save, got %d", deps.userRepo.saveCalls)
		}

		if deps.credRepo.addCalls != 0 {
			t.Fatalf("expected no credential add, got %d", deps.credRepo.addCalls)
		}

		if deps.verificationRepo.removeCalls != 0 {
			t.Fatalf(
				"expected no verification removal, got %d",
				deps.verificationRepo.removeCalls,
			)
		}

		if deps.bus.calls != 0 {
			t.Fatalf("expected no events, got %d", deps.bus.calls)
		}
	})

	t.Run("verifies user and creates credential", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		verification := newTestVerification(
			t,
			user.ID(),
			"verification-1",
		)

		deps.verificationRepo.verification = verification
		deps.userRepo.users[user.ID()] = user

		err := s.VerifyAccount(
			context.Background(),
			"verification-1",
			"password",
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if !user.IsVerified() {
			t.Fatal("expected user to be verified")
		}

		if deps.userRepo.saveCalls != 1 {
			t.Fatalf("expected one user save, got %d", deps.userRepo.saveCalls)
		}

		if deps.credRepo.addCalls != 1 {
			t.Fatalf(
				"expected one credential add, got %d",
				deps.credRepo.addCalls,
			)
		}

		if deps.verificationRepo.removeCalls != 1 {
			t.Fatalf(
				"expected one verification removal, got %d",
				deps.verificationRepo.removeCalls,
			)
		}

		if deps.bus.calls != 1 {
			t.Fatalf("expected one event, got %d", deps.bus.calls)
		}
	})
}

func TestServiceConfirmRecovery(t *testing.T) {
	t.Run("adds credential when credential does not exist", func(t *testing.T) {
		s, deps := newTestService()

		userID := newTestUserID(t, "user-1")
		recovery := newTestRecovery(
			t,
			userID,
			"recovery-token",
		)

		deps.recoveryRepo.recovery = recovery
		deps.credRepo.getByUserFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.Credential, error) {
			return nil, domainiam.ErrCredentialNotFound
		}

		err := s.ConfirmRecovery(
			context.Background(),
			"recovery-token",
			"new-password",
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.credRepo.addCalls != 1 {
			t.Fatalf("expected one credential add, got %d",
				deps.credRepo.addCalls)
		}

		if deps.credRepo.saveCalls != 0 {
			t.Fatal("expected existing credential not to be saved")
		}

		if deps.recoveryRepo.removeCalls != 1 {
			t.Fatalf("expected recovery removal, got %d",
				deps.recoveryRepo.removeCalls)
		}
	})

	t.Run("updates existing credential", func(t *testing.T) {
		s, deps := newTestService()

		userID := newTestUserID(t, "user-1")

		recovery := newTestRecovery(
			t,
			userID,
			"recovery-token",
		)

		credential := domainiam.NewCredential(
			userID,
			"old-hash",
		)

		deps.recoveryRepo.recovery = recovery
		deps.credRepo.credential = credential

		err := s.ConfirmRecovery(
			context.Background(),
			"recovery-token",
			"new-password",
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.credRepo.saveCalls != 1 {
			t.Fatalf(
				"expected one credential save, got %d",
				deps.credRepo.saveCalls,
			)
		}

		if deps.credRepo.addCalls != 0 {
			t.Fatal("expected credential not to be added")
		}

		if deps.recoveryRepo.removeCalls != 1 {
			t.Fatalf(
				"expected recovery removal, got %d",
				deps.recoveryRepo.removeCalls,
			)
		}
	})

	t.Run("does not treat arbitrary credential errors as not found", func(t *testing.T) {
		s, deps := newTestService()

		userID := newTestUserID(t, "user-1")

		recovery := newTestRecovery(
			t,
			userID,
			"recovery-token",
		)

		deps.recoveryRepo.recovery = recovery

		deps.credRepo.getByUserFn = func(
			context.Context,
			domainiam.UserID,
		) (*domainiam.Credential, error) {
			return nil, errRepository
		}

		err := s.ConfirmRecovery(
			context.Background(),
			"recovery-token",
			"new-password",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}

		if deps.credRepo.addCalls != 0 {
			t.Fatal("expected credential not to be added")
		}
	})
}

func TestServiceVerificationExists(t *testing.T) {
	t.Run("returns repository result", func(t *testing.T) {
		s, deps := newTestService()

		deps.verificationRepo.getFn = func(
			context.Context,
			domainiam.VerificationID,
		) (*domainiam.Verification, error) {
			return nil, errRepository
		}

		err := s.VerificationExists(
			context.Background(),
			"verification-1",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns nil when verification exists", func(t *testing.T) {
		s, deps := newTestService()

		userID := newTestUserID(t, "user-1")
		deps.verificationRepo.verification =
			newTestVerification(t, userID, "verification-1")

		err := s.VerificationExists(
			context.Background(),
			"verification-1",
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestServiceRecoveryExists(t *testing.T) {
	t.Run("returns repository result", func(t *testing.T) {
		s, deps := newTestService()

		deps.recoveryRepo.getByTokenFn = func(
			context.Context,
			domainiam.RecoveryToken,
		) (*domainiam.Recovery, error) {
			return nil, errRepository
		}

		err := s.RecoveryExists(
			context.Background(),
			"recovery-token",
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("returns nil when recovery exists", func(t *testing.T) {
		s, deps := newTestService()

		userID := newTestUserID(t, "user-1")
		deps.recoveryRepo.recovery =
			newTestRecovery(t, userID, "recovery-token")

		err := s.RecoveryExists(
			context.Background(),
			"recovery-token",
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestServiceRegenerateRecovery(t *testing.T) {
	t.Run("rejects inactive user", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			false,
		)

		deps.userRepo.users[user.ID()] = user

		err := s.RegenerateRecovery(
			context.Background(),
			user.Email().String(),
		)

		if !errors.Is(err, domainiam.ErrUserInactive) {
			t.Fatalf("expected inactive error, got %v", err)
		}
	})

	t.Run("replaces recovery token", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user
		deps.idGen.ids = []string{"new-recovery-token"}

		err := s.RegenerateRecovery(
			context.Background(),
			user.Email().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.recoveryRepo.removeByUserCalls != 1 {
			t.Fatalf(
				"expected one old recovery removal, got %d",
				deps.recoveryRepo.removeByUserCalls,
			)
		}

		if deps.recoveryRepo.addCalls != 1 {
			t.Fatalf(
				"expected one recovery add, got %d",
				deps.recoveryRepo.addCalls,
			)
		}

		if deps.bus.calls != 1 {
			t.Fatalf("expected one event, got %d", deps.bus.calls)
		}
	})
}

func TestServiceActivate(t *testing.T) {
	t.Run("returns nil when already active", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		user := newTestAdmin(t, "user-1")

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[user.ID()] = user

		err := s.Activate(
			context.Background(),
			user.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.userRepo.saveCalls != 0 {
			t.Fatal("expected no save for already-active user")
		}
	})

	t.Run("activates inactive user", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			false,
		)

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[user.ID()] = user

		err := s.Activate(
			context.Background(),
			user.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if !user.IsActive() {
			t.Fatal("expected user to become active")
		}

		if deps.userRepo.saveCalls != 1 {
			t.Fatalf("expected one save, got %d", deps.userRepo.saveCalls)
		}
	})
}

func TestServiceDeactivate(t *testing.T) {
	t.Run("cannot deactivate self", func(t *testing.T) {
		s, _ := newTestService()

		err := s.Deactivate(
			context.Background(),
			"user-1",
			"user-1",
		)

		if !errors.Is(err, domainiam.ErrUserCannotDeleteSelf) {
			t.Fatalf("expected self-deactivation error, got %v", err)
		}
	})

	t.Run("returns nil when already inactive", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			false,
		)

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[user.ID()] = user

		err := s.Deactivate(
			context.Background(),
			user.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.userRepo.saveCalls != 0 {
			t.Fatal("expected no save for already-inactive user")
		}
	})

	t.Run("deactivates active user", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		user := newTestAdmin(t, "user-1")

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[user.ID()] = user

		err := s.Deactivate(
			context.Background(),
			user.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if user.IsActive() {
			t.Fatal("expected user to become inactive")
		}

		if deps.userRepo.saveCalls != 1 {
			t.Fatalf("expected one save, got %d", deps.userRepo.saveCalls)
		}
	})
}

func TestServiceUpdate(t *testing.T) {
	t.Run("updates own profile", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user

		title := "Senior Developer"

		err := s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   user.ID().String(),
				UserID:    user.ID().String(),
				FirstName: "Updated",
				LastName:  "User",
				Email:     "updated@example.com",
				Title:     &title,
				Role:      domainiam.RoleClient.String(),
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if user.FirstName() != "Updated" {
			t.Fatalf("expected updated first name, got %s",
				user.FirstName())
		}

		if user.LastName() != "User" {
			t.Fatalf("expected updated last name, got %s",
				user.LastName())
		}

		if user.Email().String() != "updated@example.com" {
			t.Fatalf("expected updated email, got %s",
				user.Email())
		}

		if deps.userRepo.saveCalls != 1 {
			t.Fatalf("expected one save, got %d", deps.userRepo.saveCalls)
		}

		if deps.bus.calls == 0 {
			t.Fatal("expected user events to be published")
		}
	})

	t.Run("administrator can update another user's role", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[user.ID()] = user

		err := s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   admin.ID().String(),
				UserID:    user.ID().String(),
				FirstName: "Updated",
				LastName:  "User",
				Email:     "updated@example.com",
				Role:      domainiam.RoleStaff.String(),
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if !user.IsStaff() {
			t.Fatal("expected role to be changed to staff")
		}
	})

	t.Run("self cannot change role through Update", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user

		err := s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   user.ID().String(),
				UserID:    user.ID().String(),
				FirstName: "Updated",
				LastName:  "User",
				Email:     "updated@example.com",
				Role:      domainiam.RoleClient.String(),
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if !user.IsAdministrator() {
			t.Fatal("expected self-update not to change role")
		}
	})

	t.Run("replaces profile image", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user

		deps.idGen.ids = []string{"image-1"}

		image := &ProfileImageParams{
			FileName: "avatar.png",
			MimeType: "image/png",
			Size:     4,
			Reader:   bytes.NewReader([]byte("data")),
		}

		err := s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   user.ID().String(),
				UserID:    user.ID().String(),
				FirstName: "Updated",
				LastName:  "User",
				Email:     "updated@example.com",
				Role:      user.Role().String(),
				Image:     image,
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.fileStore.saveCalls != 1 {
			t.Fatalf("expected one image save, got %d",
				deps.fileStore.saveCalls)
		}

		if deps.fileStore.deleteCalls != 0 {
			t.Fatal("expected no old image deletion")
		}

		if deps.userRepo.saveCalls != 1 {
			t.Fatalf("expected one user save, got %d",
				deps.userRepo.saveCalls)
		}
	})

	t.Run("deletes previous profile image", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user

		// First image.
		deps.idGen.ids = []string{"old-image"}

		err := s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   user.ID().String(),
				UserID:    user.ID().String(),
				FirstName: "Test",
				LastName:  "User",
				Email:     user.Email().String(),
				Role:      user.Role().String(),
				Image: &ProfileImageParams{
					FileName: "old.png",
					MimeType: "image/png",
					Size:     4,
					Reader:   bytes.NewReader([]byte("old")),
				},
			},
		)

		if err != nil {
			t.Fatalf("first update failed: %v", err)
		}

		// Second image.
		deps.idGen.ids = []string{"new-image"}

		err = s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   user.ID().String(),
				UserID:    user.ID().String(),
				FirstName: "Test",
				LastName:  "User",
				Email:     user.Email().String(),
				Role:      user.Role().String(),
				Image: &ProfileImageParams{
					FileName: "new.png",
					MimeType: "image/png",
					Size:     4,
					Reader:   bytes.NewReader([]byte("new")),
				},
			},
		)

		if err != nil {
			t.Fatalf("second update failed: %v", err)
		}

		if deps.fileStore.deleteCalls != 1 {
			t.Fatalf(
				"expected one previous image deletion, got %d",
				deps.fileStore.deleteCalls,
			)
		}

		if deps.fileStore.deletedKey != "old-image" {
			t.Fatalf(
				"expected old-image to be deleted, got %s",
				deps.fileStore.deletedKey,
			)
		}
	})

	t.Run("does not save user when image deletion fails", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user

		deps.idGen.ids = []string{"old-image"}

		err := s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   user.ID().String(),
				UserID:    user.ID().String(),
				FirstName: "Test",
				LastName:  "User",
				Email:     user.Email().String(),
				Role:      user.Role().String(),
				Image: &ProfileImageParams{
					FileName: "old.png",
					MimeType: "image/png",
					Size:     4,
					Reader:   bytes.NewReader([]byte("old")),
				},
			},
		)

		if err != nil {
			t.Fatalf("first update failed: %v", err)
		}

		deps.idGen.ids = []string{"new-image"}

		deps.fileStore.deleteFn = func(
			context.Context,
			string,
		) error {
			return errFileStore
		}

		saveCallsBefore := deps.userRepo.saveCalls

		err = s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   user.ID().String(),
				UserID:    user.ID().String(),
				FirstName: "Test",
				LastName:  "User",
				Email:     user.Email().String(),
				Role:      user.Role().String(),
				Image: &ProfileImageParams{
					FileName: "new.png",
					MimeType: "image/png",
					Size:     4,
					Reader:   bytes.NewReader([]byte("new")),
				},
			},
		)

		if !errors.Is(err, errFileStore) {
			t.Fatalf("expected file-store error, got %v", err)
		}

		if deps.userRepo.saveCalls != saveCallsBefore {
			t.Fatal("expected user not to be saved")
		}
	})
}

func TestServiceDelete(t *testing.T) {
	t.Run("cannot delete self", func(t *testing.T) {
		s, _ := newTestService()

		err := s.Delete(
			context.Background(),
			"user-1",
			"user-1",
		)

		if !errors.Is(err, domainiam.ErrUserCannotDeleteSelf) {
			t.Fatalf("expected self-delete error, got %v", err)
		}
	})

	t.Run("administrator deletes user", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[user.ID()] = user

		err := s.Delete(
			context.Background(),
			user.ID().String(),
			admin.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if deps.userRepo.removeCalls != 1 {
			t.Fatalf("expected one remove, got %d",
				deps.userRepo.removeCalls)
		}
	})
}

func TestServiceGetUser(t *testing.T) {
	t.Run("returns user DTO", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user

		result, err := s.GetUser(
			context.Background(),
			user.ID().String(),
			user.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result == nil {
			t.Fatal("expected DTO")
		}

		if result.ID != user.ID().String() {
			t.Fatalf("expected ID %s, got %s",
				user.ID(),
				result.ID)
		}

		if result.Email != user.Email().String() {
			t.Fatalf("expected email %s, got %s",
				user.Email(),
				result.Email)
		}

		if result.Role != user.Role().String() {
			t.Fatalf("expected role %s, got %s",
				user.Role(),
				result.Role)
		}

		if result.HasImage {
			t.Fatal("expected HasImage=false")
		}
	})

	t.Run("administrator can retrieve another user", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		deps.userRepo.users[admin.ID()] = admin
		deps.userRepo.users[user.ID()] = user

		result, err := s.GetUser(
			context.Background(),
			admin.ID().String(),
			user.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result.ID != user.ID().String() {
			t.Fatalf("expected target user, got %s", result.ID)
		}
	})

	t.Run("forbidden for unrelated non-administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(
			t,
			"actor-1",
			domainiam.RoleClient,
			true,
		)

		target := newTestUser(
			t,
			"target-1",
			domainiam.RoleClient,
			true,
		)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[target.ID()] = target

		_, err := s.GetUser(
			context.Background(),
			actor.ID().String(),
			target.ID().String(),
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})
}

func TestServiceGetProfileImage(t *testing.T) {
	t.Run("returns image-not-found when user has no image", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user

		_, err := s.GetProfileImage(
			context.Background(),
			user.ID().String(),
		)

		if !errors.Is(err, domainiam.ErrUserImageNotFound) {
			t.Fatalf("expected image-not-found error, got %v", err)
		}

		if deps.fileStore.openCalls != 0 {
			t.Fatal("expected file store not to be opened")
		}
	})

	t.Run("opens and returns image", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestAdmin(t, "user-1")
		deps.userRepo.users[user.ID()] = user

		deps.idGen.ids = []string{"image-1"}

		err := s.Update(
			context.Background(),
			UpdateUserParams{
				ActorID:   user.ID().String(),
				UserID:    user.ID().String(),
				FirstName: user.FirstName(),
				LastName:  user.LastName(),
				Email:     user.Email().String(),
				Role:      user.Role().String(),
				Image: &ProfileImageParams{
					FileName: "avatar.png",
					MimeType: "image/png",
					Size:     4,
					Reader:   bytes.NewReader([]byte("data")),
				},
			},
		)

		if err != nil {
			t.Fatalf("failed to create image: %v", err)
		}

		reader := io.NopCloser(strings.NewReader("image-data"))

		deps.fileStore.openFn = func(
			context.Context,
			string,
		) (io.ReadCloser, error) {
			return reader, nil
		}

		result, err := s.GetProfileImage(
			context.Background(),
			user.ID().String(),
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result == nil {
			t.Fatal("expected image DTO")
		}

		if result.Name != "image-1" {
			t.Fatalf("expected image name image-1, got %s",
				result.Name)
		}

		if result.MimeType != "image/png" {
			t.Fatalf(
				"expected MIME type image/png, got %s",
				result.MimeType,
			)
		}

		if result.Reader != reader {
			t.Fatal("expected returned reader to be file-store reader")
		}

		if deps.fileStore.openCalls != 1 {
			t.Fatalf("expected one file open, got %d",
				deps.fileStore.openCalls)
		}
	})
}

func TestServiceListUsers(t *testing.T) {
	t.Run("requires administrator", func(t *testing.T) {
		s, deps := newTestService()

		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		deps.userRepo.users[user.ID()] = user

		_, err := s.ListUsers(
			context.Background(),
			ListUsersParams{
				ActorID: user.ID().String(),
				Limit:   10,
			},
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
	})

	t.Run("returns users and total count", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		user := newTestUser(
			t,
			"user-1",
			domainiam.RoleClient,
			true,
		)

		deps.userRepo.users[admin.ID()] = admin

		deps.userRepo.listFn = func(
			ctx context.Context,
			filter domainiam.UserFilter,
			page common.Page,
		) ([]*domainiam.User, error) {
			return []*domainiam.User{
				user,
			}, nil
		}

		deps.userRepo.countFn = func(
			ctx context.Context,
			filter domainiam.UserFilter,
		) (int, error) {
			return 5, nil
		}

		keyword := "test"
		active := true
		verified := false
		role := domainiam.RoleClient.String()

		result, err := s.ListUsers(
			context.Background(),
			ListUsersParams{
				ActorID:    admin.ID().String(),
				Keyword:    &keyword,
				Role:       &role,
				IsActive:   &active,
				IsVerified: &verified,
				Limit:      10,
				Offset:     0,
			},
		)

		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if result == nil {
			t.Fatal("expected collection")
		}

		if len(result.Items) != 1 {
			t.Fatalf(
				"expected one item, got %d",
				len(result.Items),
			)
		}

		if result.TotalCount != 5 {
			t.Fatalf(
				"expected total count 5, got %d",
				result.TotalCount,
			)
		}

		if result.Items[0].ID != user.ID().String() {
			t.Fatalf(
				"expected user ID %s, got %s",
				user.ID(),
				result.Items[0].ID,
			)
		}
	})

	t.Run("propagates list error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		deps.userRepo.users[admin.ID()] = admin

		deps.userRepo.listFn = func(
			context.Context,
			domainiam.UserFilter,
			common.Page,
		) ([]*domainiam.User, error) {
			return nil, errRepository
		}

		_, err := s.ListUsers(
			context.Background(),
			ListUsersParams{
				ActorID: admin.ID().String(),
				Limit:   10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("propagates count error", func(t *testing.T) {
		s, deps := newTestService()

		admin := newTestAdmin(t, "admin-1")
		deps.userRepo.users[admin.ID()] = admin

		deps.userRepo.listFn = func(
			context.Context,
			domainiam.UserFilter,
			common.Page,
		) ([]*domainiam.User, error) {
			return nil, nil
		}

		deps.userRepo.countFn = func(
			context.Context,
			domainiam.UserFilter,
		) (int, error) {
			return 0, errRepository
		}

		_, err := s.ListUsers(
			context.Background(),
			ListUsersParams{
				ActorID: admin.ID().String(),
				Limit:   10,
			},
		)

		if !errors.Is(err, errRepository) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestServicePublishEvents(t *testing.T) {
	t.Run("publishes all events", func(t *testing.T) {
		s, deps := newTestService()

		// Generate two real domain events by creating a user and changing it.
		user := newTestAdmin(t, "user-1")
		now := time.Now()

		user.ChangeTitle(func() *string {
			title := "Developer"
			return &title
		}(), now)

		user.ChangeEmail(
			newTestEmail(t, "new@example.com"),
			now,
		)

		events := user.PullEvents()

		if len(events) == 0 {
			t.Fatal("expected test user to contain domain events")
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

	t.Run("logs warning but does not return event-bus error", func(t *testing.T) {
		s, deps := newTestService()

		deps.bus.publishFn = func(
			context.Context,
			common.Event,
		) error {
			return errEventBus
		}

		user := newTestAdmin(t, "user-1")

		user.ChangeTitle(func() *string {
			title := "Developer"
			return &title
		}(), time.Now())

		events := user.PullEvents()

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
