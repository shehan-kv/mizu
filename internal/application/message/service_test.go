package message

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
	"mizu/internal/application/message/integration"
	"mizu/internal/application/uow"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	domainmessage "mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

var (
	errUserRepo    = errors.New("user repository error")
	errChannelRepo = errors.New("channel repository error")
	errMessageRepo = errors.New("message repository error")
	errFileRepo    = errors.New("file repository error")
	errProjectRepo = errors.New("project repository error")
	errIDGenerator = errors.New("id generator error")
	errFileStore   = errors.New("file store error")
	errEventBus    = errors.New("event bus error")
	errUOW         = errors.New("unit of work error")
)

type fakeUserRepository struct {
	iam.UserRepository

	users map[iam.UserID]*iam.User

	existsAllFn func(context.Context, []iam.UserID) (bool, error)
	listByIDsFn func(
		context.Context,
		[]iam.UserID,
		iam.UserFilter,
	) ([]*iam.User, error)
	getByIDFn func(context.Context, iam.UserID) (*iam.User, error)

	existsAllCalls int
	listByIDsCalls int
	getByIDCalls   int

	lastExistsAllIDs []iam.UserID
	lastListByIDs    []iam.UserID
}

var _ iam.UserRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) GetByID(
	ctx context.Context,
	id iam.UserID,
) (*iam.User, error) {
	f.getByIDCalls++

	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	if user, ok := f.users[id]; ok {
		return user, nil
	}

	return nil, iam.ErrUserNotFound
}

func (f *fakeUserRepository) ExistsAll(
	ctx context.Context,
	ids []iam.UserID,
) (bool, error) {
	f.existsAllCalls++
	f.lastExistsAllIDs = append([]iam.UserID(nil), ids...)

	if f.existsAllFn != nil {
		return f.existsAllFn(ctx, ids)
	}

	for _, id := range ids {
		if _, ok := f.users[id]; !ok {
			return false, nil
		}
	}

	return true, nil
}

func (f *fakeUserRepository) ListByIDs(
	ctx context.Context,
	ids []iam.UserID,
	filter iam.UserFilter,
) ([]*iam.User, error) {
	f.listByIDsCalls++
	f.lastListByIDs = append([]iam.UserID(nil), ids...)

	if f.listByIDsFn != nil {
		return f.listByIDsFn(ctx, ids, filter)
	}

	users := make([]*iam.User, 0, len(ids))

	for _, id := range ids {
		if user, ok := f.users[id]; ok {
			users = append(users, user)
		}
	}

	return users, nil
}

type fakeChannelRepository struct {
	domainmessage.ChannelRepository

	channel *domainmessage.Channel

	listByMemberFn func(
		context.Context,
		iam.UserID,
	) ([]*domainmessage.Channel, error)

	getFn func(
		context.Context,
		domainmessage.ChannelID,
	) (*domainmessage.Channel, error)

	addFn func(
		context.Context,
		*domainmessage.Channel,
	) error

	saveFn func(
		context.Context,
		*domainmessage.Channel,
	) error

	getCalls          int
	addCalls          int
	saveCalls         int
	listByMemberCalls int

	lastSaved *domainmessage.Channel
	lastAdded *domainmessage.Channel
}

var _ domainmessage.ChannelRepository = (*fakeChannelRepository)(nil)

func (f *fakeChannelRepository) Get(
	ctx context.Context,
	id domainmessage.ChannelID,
) (*domainmessage.Channel, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	if f.channel == nil {
		return nil, domainmessage.ErrChannelNotFound
	}

	return f.channel, nil
}

func (f *fakeChannelRepository) Add(
	ctx context.Context,
	channel *domainmessage.Channel,
) error {
	f.addCalls++
	f.lastAdded = channel

	if f.addFn != nil {
		return f.addFn(ctx, channel)
	}

	f.channel = channel
	return nil
}

func (f *fakeChannelRepository) Save(
	ctx context.Context,
	channel *domainmessage.Channel,
) error {
	f.saveCalls++
	f.lastSaved = channel

	if f.saveFn != nil {
		return f.saveFn(ctx, channel)
	}

	f.channel = channel
	return nil
}

func (f *fakeChannelRepository) ListByMember(
	ctx context.Context,
	memberID iam.UserID,
) ([]*domainmessage.Channel, error) {
	f.listByMemberCalls++

	if f.listByMemberFn != nil {
		return f.listByMemberFn(ctx, memberID)
	}

	if f.channel == nil {
		return nil, nil
	}

	return []*domainmessage.Channel{f.channel}, nil
}

type fakeMessageRepository struct {
	domainmessage.MessageRepository

	messages []*domainmessage.Message

	addFn func(
		context.Context,
		*domainmessage.Message,
	) error

	listByChannelFn func(
		context.Context,
		domainmessage.ChannelID,
		*domainmessage.MessageID,
		int,
	) ([]*domainmessage.Message, error)

	addCalls           int
	listByChannelCalls int

	lastAdded *domainmessage.Message

	lastListChannelID domainmessage.ChannelID
	lastBeforeID      *domainmessage.MessageID
	lastLimit         int
}

var _ domainmessage.MessageRepository = (*fakeMessageRepository)(nil)

func (f *fakeMessageRepository) Add(
	ctx context.Context,
	m *domainmessage.Message,
) error {
	f.addCalls++
	f.lastAdded = m

	if f.addFn != nil {
		return f.addFn(ctx, m)
	}

	f.messages = append(f.messages, m)
	return nil
}

func (f *fakeMessageRepository) ListByChannel(
	ctx context.Context,
	channelID domainmessage.ChannelID,
	before *domainmessage.MessageID,
	limit int,
) ([]*domainmessage.Message, error) {
	f.listByChannelCalls++
	f.lastListChannelID = channelID
	f.lastLimit = limit

	if before != nil {
		id := *before
		f.lastBeforeID = &id
	}

	if f.listByChannelFn != nil {
		return f.listByChannelFn(ctx, channelID, before, limit)
	}

	return f.messages, nil
}

type fakeFileRepository struct {
	domainmessage.FileRepository

	file  *domainmessage.File
	files []*domainmessage.File

	addFn func(
		context.Context,
		*domainmessage.File,
	) error

	getFn func(
		context.Context,
		domainmessage.FileID,
	) (*domainmessage.File, error)

	listByChannelFn func(
		context.Context,
		domainmessage.ChannelFileFilter,
		common.Page,
	) ([]*domainmessage.File, error)

	listByProjectFn func(
		context.Context,
		domainmessage.ProjectFileFilter,
		common.Page,
	) ([]*domainmessage.File, error)

	countByChannelFn func(
		context.Context,
		domainmessage.ChannelFileFilter,
	) (int, error)

	countByProjectFn func(
		context.Context,
		domainmessage.ProjectFileFilter,
	) (int, error)

	addCalls            int
	getCalls            int
	listByChannelCalls  int
	listByProjectCalls  int
	countByChannelCalls int
	countByProjectCalls int

	lastAdded *domainmessage.File

	lastChannelFilter domainmessage.ChannelFileFilter
	lastProjectFilter domainmessage.ProjectFileFilter
	lastChannelPage   common.Page
	lastProjectPage   common.Page
}

var _ domainmessage.FileRepository = (*fakeFileRepository)(nil)

func (f *fakeFileRepository) Add(
	ctx context.Context,
	file *domainmessage.File,
) error {
	f.addCalls++
	f.lastAdded = file

	if f.addFn != nil {
		return f.addFn(ctx, file)
	}

	f.file = file
	f.files = append(f.files, file)
	return nil
}

func (f *fakeFileRepository) Get(
	ctx context.Context,
	id domainmessage.FileID,
) (*domainmessage.File, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	if f.file == nil {
		return nil, errors.New("file not found")
	}

	return f.file, nil
}

func (f *fakeFileRepository) ListByChannel(
	ctx context.Context,
	filter domainmessage.ChannelFileFilter,
	page common.Page,
) ([]*domainmessage.File, error) {
	f.listByChannelCalls++
	f.lastChannelFilter = filter
	f.lastChannelPage = page

	if f.listByChannelFn != nil {
		return f.listByChannelFn(ctx, filter, page)
	}

	return f.files, nil
}

func (f *fakeFileRepository) ListByProject(
	ctx context.Context,
	filter domainmessage.ProjectFileFilter,
	page common.Page,
) ([]*domainmessage.File, error) {
	f.listByProjectCalls++
	f.lastProjectFilter = filter
	f.lastProjectPage = page

	if f.listByProjectFn != nil {
		return f.listByProjectFn(ctx, filter, page)
	}

	return f.files, nil
}

func (f *fakeFileRepository) CountByChannel(
	ctx context.Context,
	filter domainmessage.ChannelFileFilter,
) (int, error) {
	f.countByChannelCalls++

	if f.countByChannelFn != nil {
		return f.countByChannelFn(ctx, filter)
	}

	return len(f.files), nil
}

func (f *fakeFileRepository) CountByProject(
	ctx context.Context,
	filter domainmessage.ProjectFileFilter,
) (int, error) {
	f.countByProjectCalls++

	if f.countByProjectFn != nil {
		return f.countByProjectFn(ctx, filter)
	}

	return len(f.files), nil
}

type fakeProjectRepository struct {
	project.Repository

	project *project.Project

	getFn func(
		context.Context,
		project.ProjectID,
	) (*project.Project, error)

	getCalls int
}

var _ project.Repository = (*fakeProjectRepository)(nil)

func (f *fakeProjectRepository) Get(
	ctx context.Context,
	id project.ProjectID,
) (*project.Project, error) {
	f.getCalls++

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	if f.project == nil {
		return nil, errProjectRepo
	}

	return f.project, nil
}

type fakeIDGenerator struct {
	common.IDGenerator

	id  string
	err error

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

type fakeUnitOfWork struct {
	uow.UnitOfWork

	executeFn func(
		context.Context,
		func(context.Context) error,
	) error

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

type fakeInternalBus struct {
	eventbus.InternalBus

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

type fakeExternalBus struct {
	eventbus.ExternalBus

	publishFn func(context.Context, eventbus.Event) error

	events []eventbus.Event
	calls  int
}

var _ eventbus.ExternalBus = (*fakeExternalBus)(nil)

func (f *fakeExternalBus) Publish(
	ctx context.Context,
	event eventbus.Event,
) error {
	f.calls++
	f.events = append(f.events, event)

	if f.publishFn != nil {
		return f.publishFn(ctx, event)
	}

	return nil
}

type fakeFileStore struct {
	filestore.Store

	saveFn func(context.Context, string, io.Reader) error
	openFn func(context.Context, string) (io.ReadCloser, error)

	saveCalls int
	openCalls int

	savedKey  string
	openedKey string
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
	f.openedKey = key

	if f.openFn != nil {
		return f.openFn(ctx, key)
	}

	return io.NopCloser(strings.NewReader("file")), nil
}

type fakeLogger struct {
	logger.Logger

	warnCalls  int
	errorCalls int
}

var _ logger.Logger = (*fakeLogger)(nil)

func (f *fakeLogger) Warn(string, ...any) {
	f.warnCalls++
}

func (f *fakeLogger) Error(string, ...any) {
	f.errorCalls++
}

type testServiceDeps struct {
	userRepo    *fakeUserRepository
	channelRepo *fakeChannelRepository
	messageRepo *fakeMessageRepository
	fileRepo    *fakeFileRepository
	projectRepo *fakeProjectRepository

	idGen *fakeIDGenerator

	internalBus *fakeInternalBus
	externalBus *fakeExternalBus

	fileStore *fakeFileStore
	uow       *fakeUnitOfWork
	log       *fakeLogger
}

func newTestService() (*Service, *testServiceDeps) {
	userRepo := &fakeUserRepository{
		users: make(map[iam.UserID]*iam.User),
	}

	channelRepo := &fakeChannelRepository{}
	messageRepo := &fakeMessageRepository{}
	fileRepo := &fakeFileRepository{}
	projectRepo := &fakeProjectRepository{}

	idGen := &fakeIDGenerator{
		id: "generated-id",
	}

	internalBus := &fakeInternalBus{}
	externalBus := &fakeExternalBus{}

	fileStore := &fakeFileStore{}
	uow := &fakeUnitOfWork{}
	log := &fakeLogger{}

	authzService := authz.NewService(userRepo)
	messageService := domainmessage.NewService()

	service := NewService(
		userRepo,
		channelRepo,
		messageRepo,
		fileRepo,
		projectRepo,
		idGen,
		internalBus,
		externalBus,
		authzService,
		messageService,
		fileStore,
		uow,
		log,
	)

	return service, &testServiceDeps{
		userRepo:    userRepo,
		channelRepo: channelRepo,
		messageRepo: messageRepo,
		fileRepo:    fileRepo,
		projectRepo: projectRepo,
		idGen:       idGen,
		internalBus: internalBus,
		externalBus: externalBus,
		fileStore:   fileStore,
		uow:         uow,
		log:         log,
	}
}

func newTestUser(
	t *testing.T,
	id string,
	role iam.Role,
) *iam.User {
	t.Helper()

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	name, err := iam.NewName("Test", "User")
	if err != nil {
		t.Fatal(err)
	}

	email, err := iam.NewEmail(id + "@example.com")
	if err != nil {
		t.Fatal(err)
	}

	return iam.NewUser(
		userID,
		name,
		email,
		nil,
		role,
		true,
		userID,
		time.Now(),
	)
}

func newTestProject(
	t *testing.T,
	id string,
	creator string,
	members ...string,
) *project.Project {
	t.Helper()

	projectID, err := project.NewProjectID(id)
	if err != nil {
		t.Fatal(err)
	}

	creatorID, err := iam.NewUserID(creator)
	if err != nil {
		t.Fatal(err)
	}

	memberIDs := make([]iam.UserID, 0, len(members)+1)

	for _, member := range members {
		memberID, err := iam.NewUserID(member)
		if err != nil {
			t.Fatal(err)
		}

		memberIDs = append(memberIDs, memberID)
	}

	name, err := project.NewName("Test Project")
	if err != nil {
		t.Fatal(err)
	}

	p, err := project.NewProject(
		projectID,
		name,
		project.StatusStarted,
		creatorID,
		memberIDs,
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func newTestChannel(
	t *testing.T,
	id string,
	members ...string,
) *domainmessage.Channel {
	t.Helper()

	channelID, err := domainmessage.NewChannelID(id)
	if err != nil {
		t.Fatal(err)
	}

	memberIDs := make([]iam.UserID, 0, len(members))

	for _, member := range members {
		memberID, err := iam.NewUserID(member)
		if err != nil {
			t.Fatal(err)
		}

		memberIDs = append(memberIDs, memberID)
	}

	name, err := domainmessage.NewChannelName("Test Channel")
	if err != nil {
		t.Fatal(err)
	}

	channel, err := domainmessage.NewChannel(
		channelID,
		nil,
		name,
		memberIDs,
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}

	return channel
}

func newTestMessage(
	t *testing.T,
	id string,
	channelID string,
	senderID string,
	content string,
) *domainmessage.Message {
	t.Helper()

	msgID, err := domainmessage.NewMessageID(id)
	if err != nil {
		t.Fatal(err)
	}

	cID, err := domainmessage.NewChannelID(channelID)
	if err != nil {
		t.Fatal(err)
	}

	sID, err := iam.NewUserID(senderID)
	if err != nil {
		t.Fatal(err)
	}

	messageContent, err := domainmessage.NewContent(content)
	if err != nil {
		t.Fatal(err)
	}

	m, err := domainmessage.NewUserMessage(
		msgID,
		cID,
		sID,
		messageContent,
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}

	return m
}

func newTestFile(
	t *testing.T,
	id string,
	channelID string,
	userID string,
) *domainmessage.File {
	t.Helper()

	fileID, err := domainmessage.NewFileID(id)
	if err != nil {
		t.Fatal(err)
	}

	cID, err := domainmessage.NewChannelID(channelID)
	if err != nil {
		t.Fatal(err)
	}

	uID, err := iam.NewUserID(userID)
	if err != nil {
		t.Fatal(err)
	}

	originalName, err := domainmessage.NewFileName("document.pdf")
	if err != nil {
		t.Fatal(err)
	}

	savedName, err := domainmessage.NewFileName(id)
	if err != nil {
		t.Fatal(err)
	}

	return domainmessage.NewFile(
		fileID,
		cID,
		uID,
		originalName,
		savedName,
		id,
		"application/pdf",
		1024,
		time.Now(),
	)
}

func TestServiceReplaceChannelMembers(t *testing.T) {

	t.Run("requires administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		member := newTestUser(t, "member-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member.ID()] = member

		channel := newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)
		deps.channelRepo.channel = channel

		err := s.ReplaceChannelMembers(
			context.Background(),
			"channel-1",
			[]string{"member-1"},
			"actor-1",
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}

		if deps.channelRepo.saveCalls != 0 {
			t.Fatal("expected channel not to be saved")
		}
	})

	t.Run("returns channel-not-member when administrator is not a member", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		member1 := newTestUser(t, "member-1", iam.RoleClient)
		member2 := newTestUser(t, "member-2", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member1.ID()] = member1
		deps.userRepo.users[member2.ID()] = member2

		channel := newTestChannel(
			t,
			"channel-1",
			"member-1",
			"member-2",
		)
		deps.channelRepo.channel = channel

		err := s.ReplaceChannelMembers(
			context.Background(),
			"channel-1",
			[]string{"member-2"},
			"actor-1",
		)

		if !errors.Is(err, domainmessage.ErrNotChannelMember) {
			t.Fatalf("expected not-channel-member, got %v", err)
		}

		if deps.userRepo.existsAllCalls != 0 {
			t.Fatal("expected member existence check not to run")
		}
	})

	t.Run("returns user-not-found when a replacement member does not exist", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		deps.userRepo.users[actor.ID()] = actor

		channel := newTestChannel(t, "channel-1", "actor-1", "old-member")
		deps.channelRepo.channel = channel

		err := s.ReplaceChannelMembers(
			context.Background(),
			"channel-1",
			[]string{"missing-member"},
			"actor-1",
		)

		if !errors.Is(err, iam.ErrUserNotFound) {
			t.Fatalf("expected user-not-found, got %v", err)
		}

		if deps.channelRepo.saveCalls != 0 {
			t.Fatal("expected channel not to be saved")
		}
	})

	t.Run("always keeps actor as a member", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		member := newTestUser(t, "member-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member.ID()] = member

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"old-member",
		)

		err := s.ReplaceChannelMembers(
			context.Background(),
			"channel-1",
			[]string{"member-1"},
			"actor-1",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deps.uow.calls != 1 {
			t.Fatalf("expected one unit-of-work execution, got %d", deps.uow.calls)
		}

		if deps.channelRepo.saveCalls != 1 {
			t.Fatalf("expected one channel save, got %d", deps.channelRepo.saveCalls)
		}

		if !deps.channelRepo.lastSaved.HasMember(actor.ID()) {
			t.Fatal("expected actor to remain a channel member")
		}

		if !deps.channelRepo.lastSaved.HasMember(member.ID()) {
			t.Fatal("expected replacement member to be present")
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.getFn = func(
			context.Context,
			domainmessage.ChannelID,
		) (*domainmessage.Channel, error) {
			return nil, errChannelRepo
		}

		err := s.ReplaceChannelMembers(
			context.Background(),
			"channel-1",
			[]string{"member-1"},
			"actor-1",
		)

		if !errors.Is(err, errChannelRepo) {
			t.Fatalf("expected channel repository error, got %v", err)
		}
	})

	t.Run("propagates user existence error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		member := newTestUser(t, "member-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member.ID()] = member

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.userRepo.existsAllFn = func(
			context.Context,
			[]iam.UserID,
		) (bool, error) {
			return false, errUserRepo
		}

		err := s.ReplaceChannelMembers(
			context.Background(),
			"channel-1",
			[]string{"member-1"},
			"actor-1",
		)

		if !errors.Is(err, errUserRepo) {
			t.Fatalf("expected user repository error, got %v", err)
		}
	})

	t.Run("propagates unit-of-work error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		member := newTestUser(t, "member-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member.ID()] = member

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.uow.executeFn = func(
			context.Context,
			func(context.Context) error,
		) error {
			return errUOW
		}

		err := s.ReplaceChannelMembers(
			context.Background(),
			"channel-1",
			[]string{"member-1"},
			"actor-1",
		)

		if !errors.Is(err, errUOW) {
			t.Fatalf("expected unit-of-work error, got %v", err)
		}
	})

}

func TestServiceCreateChannel(t *testing.T) {
	t.Run("creates standalone channel", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		member := newTestUser(t, "member-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member.ID()] = member

		err := s.CreateChannel(
			context.Background(),
			CreateChannelParams{
				ActorID:   "actor-1",
				Name:      "General",
				MemberIDs: []string{"member-1"},
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deps.idGen.calls != 1 {
			t.Fatalf("expected one ID generation, got %d", deps.idGen.calls)
		}

		if deps.channelRepo.addCalls != 1 {
			t.Fatalf("expected one channel add, got %d", deps.channelRepo.addCalls)
		}

		if deps.channelRepo.lastAdded.ID().String() != "generated-id" {
			t.Fatalf(
				"expected generated channel ID, got %s",
				deps.channelRepo.lastAdded.ID(),
			)
		}

		if !deps.channelRepo.lastAdded.HasMember(actor.ID()) {
			t.Fatal("expected actor to be added as member")
		}

		if !deps.channelRepo.lastAdded.HasMember(member.ID()) {
			t.Fatal("expected requested member to be present")
		}
	})

	t.Run("rejects standalone channel without staff or administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		client := newTestUser(t, "client-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[client.ID()] = client

		// The administrator is automatically added to the member list,
		// therefore this case is valid. To test the domain validation itself,
		// use a staff/client configuration through ListByIDs.
		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return []*iam.User{client}, nil
		}

		err := s.CreateChannel(
			context.Background(),
			CreateChannelParams{
				ActorID:   "actor-1",
				Name:      "General",
				MemberIDs: []string{"client-1"},
			},
		)

		if !errors.Is(err, domainmessage.ErrChannelMustHaveStaffMember) {
			t.Fatalf(
				"expected ErrChannelMustHaveStaffMember, got %v",
				err,
			)
		}

		if deps.channelRepo.addCalls != 0 {
			t.Fatal("expected channel not to be added")
		}
	})

	t.Run("creates project channel when actor belongs to project", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		member := newTestUser(t, "member-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member.ID()] = member

		deps.projectRepo.project = newTestProject(
			t,
			"project-1",
			"actor-1",
			"actor-1",
			"member-1",
		)

		projectID := "project-1"

		err := s.CreateChannel(
			context.Background(),
			CreateChannelParams{
				ActorID:   "actor-1",
				Name:      "Project Channel",
				MemberIDs: []string{"member-1"},
				ProjectID: &projectID,
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deps.projectRepo.getCalls != 1 {
			t.Fatalf("expected one project lookup, got %d", deps.projectRepo.getCalls)
		}

		if deps.userRepo.existsAllCalls != 1 {
			t.Fatalf(
				"expected one member existence check, got %d",
				deps.userRepo.existsAllCalls,
			)
		}

		if deps.channelRepo.addCalls != 1 {
			t.Fatalf("expected one channel add, got %d", deps.channelRepo.addCalls)
		}

		if deps.channelRepo.lastAdded.ProjectID() == nil {
			t.Fatal("expected project ID on channel")
		}

		if deps.channelRepo.lastAdded.ProjectID().String() != "project-1" {
			t.Fatalf(
				"expected project-1, got %s",
				deps.channelRepo.lastAdded.ProjectID(),
			)
		}
	})

	t.Run("rejects project channel when actor is not project member", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		deps.userRepo.users[actor.ID()] = actor

		deps.projectRepo.project = newTestProject(
			t,
			"project-1",
			"member-1",
			"member-1",
		)

		projectID := "project-1"

		err := s.CreateChannel(
			context.Background(),
			CreateChannelParams{
				ActorID:   "actor-1",
				Name:      "Project Channel",
				ProjectID: &projectID,
			},
		)

		if !errors.Is(err, project.ErrNotProjectMember) {
			t.Fatalf(
				"expected project-not-member, got %v",
				err,
			)
		}

		if deps.channelRepo.addCalls != 0 {
			t.Fatal("expected channel not to be added")
		}
	})

	t.Run("rejects project channel with missing member", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		deps.userRepo.users[actor.ID()] = actor

		deps.projectRepo.project = newTestProject(
			t,
			"project-1",
			"actor-1",
			"actor-1",
		)

		deps.userRepo.existsAllFn = func(
			context.Context,
			[]iam.UserID,
		) (bool, error) {
			return false, nil
		}

		projectID := "project-1"

		err := s.CreateChannel(
			context.Background(),
			CreateChannelParams{
				ActorID:   "actor-1",
				Name:      "Project Channel",
				MemberIDs: []string{"missing-member"},
				ProjectID: &projectID,
			},
		)

		if !errors.Is(err, domainmessage.ErrNotChannelMember) {
			t.Fatalf(
				"expected not-channel-member, got %v",
				err,
			)
		}

		if deps.channelRepo.addCalls != 0 {
			t.Fatal("expected channel not to be added")
		}
	})

	t.Run("propagates ID generation error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleAdministrator)
		deps.userRepo.users[actor.ID()] = actor

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return []*iam.User{actor}, nil
		}

		deps.idGen.err = errIDGenerator

		err := s.CreateChannel(
			context.Background(),
			CreateChannelParams{
				ActorID: "actor-1",
				Name:    "General",
			},
		)

		if !errors.Is(err, errIDGenerator) {
			t.Fatalf("expected ID generator error, got %v", err)
		}

		if deps.channelRepo.addCalls != 0 {
			t.Fatal("expected channel not to be added")
		}
	})
}

func TestServiceListChannelsByMember(t *testing.T) {

	t.Run("returns channel DTOs", func(t *testing.T) {
		s, deps := newTestService()

		channel := newTestChannel(
			t,
			"channel-1",
			"member-1",
			"member-2",
		)
		deps.channelRepo.channel = channel

		result, err := s.ListChannelsByMember(
			context.Background(),
			"member-1",
			"member-1",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected one channel, got %d", len(result))
		}

		if result[0].ID != "channel-1" {
			t.Fatalf("expected channel-1, got %s", result[0].ID)
		}

		if result[0].Name != "Test Channel" {
			t.Fatalf("expected Test Channel, got %s", result[0].Name)
		}

		if result[0].ProjectID != nil {
			t.Fatal("expected nil project ID for standalone channel")
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.listByMemberFn = func(
			context.Context,
			iam.UserID,
		) ([]*domainmessage.Channel, error) {
			return nil, errChannelRepo
		}

		_, err := s.ListChannelsByMember(
			context.Background(),
			"member-1",
			"member-1",
		)

		if !errors.Is(err, errChannelRepo) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("forbids unrelated non-administrator", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		target := newTestUser(t, "target-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[target.ID()] = target

		_, err := s.ListChannelsByMember(
			context.Background(),
			"actor-1",
			"target-1",
		)

		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}

		if deps.channelRepo.listByMemberCalls != 0 {
			t.Fatal("expected repository not to be called")
		}
	})
}

func TestServiceListChannelMembers(t *testing.T) {
	t.Run("returns member DTOs", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		member := newTestUser(t, "member-1", iam.RoleStaff)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member.ID()] = member

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		result, err := s.ListChannelMembers(
			context.Background(),
			"actor-1",
			"channel-1",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected two members, got %d", len(result))
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"member-1",
			"member-2",
		)

		_, err := s.ListChannelMembers(
			context.Background(),
			"other-1",
			"channel-1",
		)

		if !errors.Is(err, project.ErrNotProjectMember) {
			t.Fatalf("expected not-project-member, got %v", err)
		}

		if deps.userRepo.listByIDsCalls != 0 {
			t.Fatal("expected user repository not to be called")
		}
	})

	t.Run("propagates user repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return nil, errUserRepo
		}

		_, err := s.ListChannelMembers(
			context.Background(),
			"actor-1",
			"channel-1",
		)

		if !errors.Is(err, errUserRepo) {
			t.Fatalf("expected user repository error, got %v", err)
		}
	})

}

func TestServiceCreateMessage(t *testing.T) {
	t.Run("creates message and publishes broadcast", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleStaff)
		member := newTestUser(t, "member-1", iam.RoleClient)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[member.ID()] = member

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		err := s.CreateMessage(
			context.Background(),
			CreateMessageParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Content:   "Hello",
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deps.messageRepo.addCalls != 1 {
			t.Fatalf("expected one message add, got %d", deps.messageRepo.addCalls)
		}

		if deps.messageRepo.lastAdded.ID().String() != "generated-id" {
			t.Fatalf(
				"expected generated-id, got %s",
				deps.messageRepo.lastAdded.ID(),
			)
		}

		if deps.messageRepo.lastAdded.Content().String() != "Hello" {
			t.Fatalf(
				"expected Hello, got %s",
				deps.messageRepo.lastAdded.Content(),
			)
		}

		if deps.externalBus.calls != 1 {
			t.Fatalf(
				"expected one external event, got %d",
				deps.externalBus.calls,
			)
		}

		broadcast, ok := deps.externalBus.events[0].(integration.MessageBroadcast)
		if !ok {
			t.Fatalf("expected MessageBroadcast, got %T", deps.externalBus.events[0])
		}

		if broadcast.MessageID != "generated-id" {
			t.Fatalf("expected generated-id, got %s", broadcast.MessageID)
		}

		if broadcast.ChannelID != "channel-1" {
			t.Fatalf("expected channel-1, got %s", broadcast.ChannelID)
		}

		if broadcast.SenderID != "actor-1" {
			t.Fatalf("expected actor-1, got %s", broadcast.SenderID)
		}

		if broadcast.Content != "Hello" {
			t.Fatalf("expected Hello, got %s", broadcast.Content)
		}

		if len(broadcast.To) != 2 {
			t.Fatalf("expected two recipients, got %d", len(broadcast.To))
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"member-1",
			"member-2",
		)

		err := s.CreateMessage(
			context.Background(),
			CreateMessageParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Content:   "Hello",
			},
		)

		if !errors.Is(err, domainmessage.ErrNotChannelMember) {
			t.Fatalf("expected not-channel-member, got %v", err)
		}

		if deps.messageRepo.addCalls != 0 {
			t.Fatal("expected message not to be added")
		}

		if deps.externalBus.calls != 0 {
			t.Fatal("expected event not to be published")
		}
	})

	t.Run("propagates message repository error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleStaff)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.messageRepo.addFn = func(
			context.Context,
			*domainmessage.Message,
		) error {
			return errMessageRepo
		}

		err := s.CreateMessage(
			context.Background(),
			CreateMessageParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Content:   "Hello",
			},
		)

		if !errors.Is(err, errMessageRepo) {
			t.Fatalf("expected message repository error, got %v", err)
		}

		if deps.externalBus.calls != 0 {
			t.Fatal("expected event not to be published")
		}
	})

	t.Run("does not fail when external event publication fails", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleStaff)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.externalBus.publishFn = func(
			context.Context,
			eventbus.Event,
		) error {
			return errEventBus
		}

		err := s.CreateMessage(
			context.Background(),
			CreateMessageParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Content:   "Hello",
			},
		)

		if err != nil {
			t.Fatalf(
				"expected publication failure to be swallowed, got %v",
				err,
			)
		}

		if deps.log.errorCalls != 1 {
			t.Fatalf(
				"expected one error log, got %d",
				deps.log.errorCalls,
			)
		}
	})

}

func TestServiceListChannelMessages(t *testing.T) {
	t.Run("returns messages in reverse repository order", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		sender := newTestUser(t, "sender-1", iam.RoleStaff)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[sender.ID()] = sender

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"sender-1",
		)

		first := newTestMessage(
			t,
			"message-1",
			"channel-1",
			"sender-1",
			"First",
		)

		second := newTestMessage(
			t,
			"message-2",
			"channel-1",
			"sender-1",
			"Second",
		)

		deps.messageRepo.messages = []*domainmessage.Message{
			first,
			second,
		}

		result, err := s.ListChannelMessages(
			context.Background(),
			ListChannelMessagesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     20,
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected two messages, got %d", len(result))
		}

		if result[0].ID != "message-2" {
			t.Fatalf("expected message-2 first, got %s", result[0].ID)
		}

		if result[1].ID != "message-1" {
			t.Fatalf("expected message-1 second, got %s", result[1].ID)
		}

		if result[0].Sender == nil {
			t.Fatal("expected sender DTO")
		}

		if result[0].Sender.ID != "sender-1" {
			t.Fatalf(
				"expected sender-1, got %s",
				result[0].Sender.ID,
			)
		}

		if deps.userRepo.listByIDsCalls != 1 {
			t.Fatalf(
				"expected one sender lookup, got %d",
				deps.userRepo.listByIDsCalls,
			)
		}
	})

	t.Run("deduplicates sender IDs", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		sender := newTestUser(t, "sender-1", iam.RoleStaff)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[sender.ID()] = sender

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"sender-1",
		)

		deps.messageRepo.messages = []*domainmessage.Message{
			newTestMessage(
				t,
				"message-1",
				"channel-1",
				"sender-1",
				"First",
			),
			newTestMessage(
				t,
				"message-2",
				"channel-1",
				"sender-1",
				"Second",
			),
		}

		_, err := s.ListChannelMessages(
			context.Background(),
			ListChannelMessagesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     20,
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(deps.userRepo.lastListByIDs) != 1 {
			t.Fatalf(
				"expected one unique sender ID, got %d",
				len(deps.userRepo.lastListByIDs),
			)
		}

		if deps.userRepo.lastListByIDs[0].String() != "sender-1" {
			t.Fatalf(
				"expected sender-1, got %s",
				deps.userRepo.lastListByIDs[0],
			)
		}
	})

	t.Run("passes before message ID and limit to repository", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		before := "message-before"

		_, err := s.ListChannelMessages(
			context.Background(),
			ListChannelMessagesParams{
				ActorID:         "actor-1",
				ChannelID:       "channel-1",
				BeforeMessageID: &before,
				Limit:           25,
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deps.messageRepo.lastBeforeID == nil {
			t.Fatal("expected before message ID")
		}

		if deps.messageRepo.lastBeforeID.String() != before {
			t.Fatalf(
				"expected before ID %s, got %s",
				before,
				deps.messageRepo.lastBeforeID,
			)
		}

		if deps.messageRepo.lastLimit != 25 {
			t.Fatalf(
				"expected limit 25, got %d",
				deps.messageRepo.lastLimit,
			)
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"member-1",
			"member-2",
		)

		_, err := s.ListChannelMessages(
			context.Background(),
			ListChannelMessagesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     10,
			},
		)

		if !errors.Is(err, domainmessage.ErrNotChannelMember) {
			t.Fatalf("expected not-channel-member, got %v", err)
		}

		if deps.messageRepo.listByChannelCalls != 0 {
			t.Fatal("expected message repository not to be called")
		}
	})

	t.Run("propagates user repository error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.messageRepo.messages = []*domainmessage.Message{
			newTestMessage(
				t,
				"message-1",
				"channel-1",
				"sender-1",
				"Hello",
			),
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return nil, errUserRepo
		}

		_, err := s.ListChannelMessages(
			context.Background(),
			ListChannelMessagesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     10,
			},
		)

		if !errors.Is(err, errUserRepo) {
			t.Fatalf("expected user repository error, got %v", err)
		}
	})

}

func TestServiceUploadFile(t *testing.T) {

	t.Run("saves file and persists metadata", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		err := s.UploadFile(
			context.Background(),
			UploadFileParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				FileName:  "document.pdf",
				MimeType:  "application/pdf",
				Size:      1024,
				Reader:    bytes.NewReader([]byte("content")),
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deps.fileStore.saveCalls != 1 {
			t.Fatalf(
				"expected one file-store save, got %d",
				deps.fileStore.saveCalls,
			)
		}

		if deps.fileStore.savedKey != "generated-id" {
			t.Fatalf(
				"expected generated-id storage key, got %s",
				deps.fileStore.savedKey,
			)
		}

		if deps.fileRepo.addCalls != 1 {
			t.Fatalf(
				"expected one file repository add, got %d",
				deps.fileRepo.addCalls,
			)
		}

		if deps.fileRepo.lastAdded.ID().String() != "generated-id" {
			t.Fatalf(
				"expected generated-id file ID, got %s",
				deps.fileRepo.lastAdded.ID(),
			)
		}

		if deps.fileRepo.lastAdded.OriginalName().String() != "document.pdf" {
			t.Fatalf(
				"expected document.pdf, got %s",
				deps.fileRepo.lastAdded.OriginalName(),
			)
		}

		if deps.fileRepo.lastAdded.StorageKey() != "generated-id" {
			t.Fatalf(
				"expected generated-id storage key, got %s",
				deps.fileRepo.lastAdded.StorageKey(),
			)
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"member-1",
			"member-2",
		)

		err := s.UploadFile(
			context.Background(),
			UploadFileParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				FileName:  "document.pdf",
				MimeType:  "application/pdf",
				Size:      1024,
				Reader:    bytes.NewReader([]byte("content")),
			},
		)

		if !errors.Is(err, domainmessage.ErrNotChannelMember) {
			t.Fatalf("expected not-channel-member, got %v", err)
		}

		if deps.fileStore.saveCalls != 0 {
			t.Fatal("expected file store not to be called")
		}

		if deps.fileRepo.addCalls != 0 {
			t.Fatal("expected file repository not to be called")
		}
	})

	t.Run("propagates file-store error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.fileStore.saveFn = func(
			context.Context,
			string,
			io.Reader,
		) error {
			return errFileStore
		}

		err := s.UploadFile(
			context.Background(),
			UploadFileParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				FileName:  "document.pdf",
				MimeType:  "application/pdf",
				Size:      1024,
				Reader:    bytes.NewReader([]byte("content")),
			},
		)

		if !errors.Is(err, errFileStore) {
			t.Fatalf("expected file-store error, got %v", err)
		}

		if deps.fileRepo.addCalls != 0 {
			t.Fatal("expected file repository not to be called")
		}
	})

	t.Run("propagates file repository error", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.fileRepo.addFn = func(
			context.Context,
			*domainmessage.File,
		) error {
			return errFileRepo
		}

		err := s.UploadFile(
			context.Background(),
			UploadFileParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				FileName:  "document.pdf",
				MimeType:  "application/pdf",
				Size:      1024,
				Reader:    bytes.NewReader([]byte("content")),
			},
		)

		if !errors.Is(err, errFileRepo) {
			t.Fatalf("expected file repository error, got %v", err)
		}
	})

}

func TestServiceDownloadFile(t *testing.T) {

	t.Run("returns file download DTO", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		deps.userRepo.users[actor.ID()] = actor

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		file := newTestFile(
			t,
			"file-1",
			"channel-1",
			"actor-1",
		)

		deps.fileRepo.file = file

		reader := io.NopCloser(strings.NewReader("download"))

		deps.fileStore.openFn = func(
			context.Context,
			string,
		) (io.ReadCloser, error) {
			return reader, nil
		}

		result, err := s.DownloadFile(
			context.Background(),
			"actor-1",
			"file-1",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected download DTO")
		}

		if result.Name != "document.pdf" {
			t.Fatalf("expected document.pdf, got %s", result.Name)
		}

		if result.MimeType != "application/pdf" {
			t.Fatalf(
				"expected application/pdf, got %s",
				result.MimeType,
			)
		}

		if result.Size != 1024 {
			t.Fatalf("expected size 1024, got %d", result.Size)
		}

		if result.Reader != reader {
			t.Fatal("expected returned reader to be file-store reader")
		}

		if deps.fileStore.openedKey != "file-1" {
			t.Fatalf(
				"expected file-1 storage key, got %s",
				deps.fileStore.openedKey,
			)
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"member-1",
			"member-2",
		)

		deps.fileRepo.file = newTestFile(
			t,
			"file-1",
			"channel-1",
			"member-1",
		)

		_, err := s.DownloadFile(
			context.Background(),
			"actor-1",
			"file-1",
		)

		if !errors.Is(err, domainmessage.ErrNotChannelMember) {
			t.Fatalf("expected not-channel-member, got %v", err)
		}

		if deps.fileStore.openCalls != 0 {
			t.Fatal("expected file store not to be opened")
		}
	})

	t.Run("propagates file repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.fileRepo.getFn = func(
			context.Context,
			domainmessage.FileID,
		) (*domainmessage.File, error) {
			return nil, errFileRepo
		}

		_, err := s.DownloadFile(
			context.Background(),
			"actor-1",
			"file-1",
		)

		if !errors.Is(err, errFileRepo) {
			t.Fatalf("expected file repository error, got %v", err)
		}
	})

	t.Run("propagates file-store open error", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.fileRepo.file = newTestFile(
			t,
			"file-1",
			"channel-1",
			"actor-1",
		)

		deps.fileStore.openFn = func(
			context.Context,
			string,
		) (io.ReadCloser, error) {
			return nil, errFileStore
		}

		_, err := s.DownloadFile(
			context.Background(),
			"actor-1",
			"file-1",
		)

		if !errors.Is(err, errFileStore) {
			t.Fatalf("expected file-store error, got %v", err)
		}
	})
}

func TestServiceListChannelFiles(t *testing.T) {
	t.Run("returns files, sender DTOs, and total count", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		uploader := newTestUser(t, "uploader-1", iam.RoleStaff)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[uploader.ID()] = uploader

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"uploader-1",
		)

		file := newTestFile(
			t,
			"file-1",
			"channel-1",
			"uploader-1",
		)

		deps.fileRepo.files = []*domainmessage.File{file}

		result, err := s.ListChannelFiles(
			context.Background(),
			ListChannelFilesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Items) != 1 {
			t.Fatalf("expected one file, got %d", len(result.Items))
		}

		if result.TotalCount != 1 {
			t.Fatalf(
				"expected total count 1, got %d",
				result.TotalCount,
			)
		}

		if result.Items[0].ID != "file-1" {
			t.Fatalf(
				"expected file-1, got %s",
				result.Items[0].ID,
			)
		}

		if result.Items[0].Name != "document.pdf" {
			t.Fatalf(
				"expected document.pdf, got %s",
				result.Items[0].Name,
			)
		}

		if result.Items[0].User.ID != "uploader-1" {
			t.Fatalf(
				"expected uploader-1, got %s",
				result.Items[0].User.ID,
			)
		}

		if result.Items[0].User.Role != uploader.Role().String() {
			t.Fatalf(
				"expected role %s, got %s",
				uploader.Role(),
				result.Items[0].User.Role,
			)
		}

		if deps.fileRepo.listByChannelCalls != 1 {
			t.Fatalf(
				"expected one file list call, got %d",
				deps.fileRepo.listByChannelCalls,
			)
		}

		if deps.fileRepo.countByChannelCalls != 1 {
			t.Fatalf(
				"expected one count call, got %d",
				deps.fileRepo.countByChannelCalls,
			)
		}
	})

	t.Run("rejects non-member", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"member-1",
			"member-2",
		)

		_, err := s.ListChannelFiles(
			context.Background(),
			ListChannelFilesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, domainmessage.ErrNotChannelMember) {
			t.Fatalf("expected not-channel-member, got %v", err)
		}

		if deps.fileRepo.listByChannelCalls != 0 {
			t.Fatal("expected file repository not to be called")
		}
	})

	t.Run("propagates list error", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.fileRepo.listByChannelFn = func(
			context.Context,
			domainmessage.ChannelFileFilter,
			common.Page,
		) ([]*domainmessage.File, error) {
			return nil, errFileRepo
		}

		_, err := s.ListChannelFiles(
			context.Background(),
			ListChannelFilesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errFileRepo) {
			t.Fatalf("expected file repository error, got %v", err)
		}
	})

	t.Run("propagates user lookup error", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.fileRepo.files = []*domainmessage.File{
			newTestFile(
				t,
				"file-1",
				"channel-1",
				"uploader-1",
			),
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return nil, errUserRepo
		}

		_, err := s.ListChannelFiles(
			context.Background(),
			ListChannelFilesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errUserRepo) {
			t.Fatalf("expected user repository error, got %v", err)
		}
	})

	t.Run("propagates count error", func(t *testing.T) {
		s, deps := newTestService()

		deps.channelRepo.channel = newTestChannel(
			t,
			"channel-1",
			"actor-1",
			"member-1",
		)

		deps.fileRepo.countByChannelFn = func(
			context.Context,
			domainmessage.ChannelFileFilter,
		) (int, error) {
			return 0, errFileRepo
		}

		_, err := s.ListChannelFiles(
			context.Background(),
			ListChannelFilesParams{
				ActorID:   "actor-1",
				ChannelID: "channel-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errFileRepo) {
			t.Fatalf("expected file repository error, got %v", err)
		}
	})
}

func TestServiceListProjectFiles(t *testing.T) {
	t.Run("returns project files and total count", func(t *testing.T) {
		s, deps := newTestService()

		actor := newTestUser(t, "actor-1", iam.RoleClient)
		uploader := newTestUser(t, "uploader-1", iam.RoleStaff)

		deps.userRepo.users[actor.ID()] = actor
		deps.userRepo.users[uploader.ID()] = uploader

		deps.projectRepo.project = newTestProject(
			t,
			"project-1",
			"actor-1",
			"actor-1",
			"uploader-1",
		)

		deps.fileRepo.files = []*domainmessage.File{
			newTestFile(
				t,
				"file-1",
				"channel-1",
				"uploader-1",
			),
		}

		result, err := s.ListProjectFiles(
			context.Background(),
			ListProjectFilesParams{
				ActorID:   "actor-1",
				ProjectID: "project-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Items) != 1 {
			t.Fatalf("expected one file, got %d", len(result.Items))
		}

		if result.TotalCount != 1 {
			t.Fatalf(
				"expected total count 1, got %d",
				result.TotalCount,
			)
		}

		if result.Items[0].ID != "file-1" {
			t.Fatalf(
				"expected file-1, got %s",
				result.Items[0].ID,
			)
		}

		if result.Items[0].User.ID != "uploader-1" {
			t.Fatalf(
				"expected uploader-1, got %s",
				result.Items[0].User.ID,
			)
		}
	})

	t.Run("rejects non-project-member", func(t *testing.T) {
		s, deps := newTestService()

		deps.projectRepo.project = newTestProject(
			t,
			"project-1",
			"member-1",
			"member-1",
		)

		_, err := s.ListProjectFiles(
			context.Background(),
			ListProjectFilesParams{
				ActorID:   "actor-1",
				ProjectID: "project-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, project.ErrNotProjectMember) {
			t.Fatalf(
				"expected project-not-member, got %v",
				err,
			)
		}

		if deps.fileRepo.listByProjectCalls != 0 {
			t.Fatal("expected file repository not to be called")
		}
	})

	t.Run("propagates project repository error", func(t *testing.T) {
		s, deps := newTestService()

		deps.projectRepo.getFn = func(
			context.Context,
			project.ProjectID,
		) (*project.Project, error) {
			return nil, errProjectRepo
		}

		_, err := s.ListProjectFiles(
			context.Background(),
			ListProjectFilesParams{
				ActorID:   "actor-1",
				ProjectID: "project-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errProjectRepo) {
			t.Fatalf(
				"expected project repository error, got %v",
				err,
			)
		}

		if deps.fileRepo.listByProjectCalls != 0 {
			t.Fatal("expected file repository not to be called")
		}
	})

	t.Run("propagates file repository list error", func(t *testing.T) {
		s, deps := newTestService()

		deps.projectRepo.project = newTestProject(
			t,
			"project-1",
			"actor-1",
			"actor-1",
		)

		deps.fileRepo.listByProjectFn = func(
			context.Context,
			domainmessage.ProjectFileFilter,
			common.Page,
		) ([]*domainmessage.File, error) {
			return nil, errFileRepo
		}

		_, err := s.ListProjectFiles(
			context.Background(),
			ListProjectFilesParams{
				ActorID:   "actor-1",
				ProjectID: "project-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errFileRepo) {
			t.Fatalf(
				"expected file repository error, got %v",
				err,
			)
		}
	})

	t.Run("propagates user lookup error", func(t *testing.T) {
		s, deps := newTestService()

		deps.projectRepo.project = newTestProject(
			t,
			"project-1",
			"actor-1",
			"actor-1",
		)

		deps.fileRepo.files = []*domainmessage.File{
			newTestFile(
				t,
				"file-1",
				"channel-1",
				"uploader-1",
			),
		}

		deps.userRepo.listByIDsFn = func(
			context.Context,
			[]iam.UserID,
			iam.UserFilter,
		) ([]*iam.User, error) {
			return nil, errUserRepo
		}

		_, err := s.ListProjectFiles(
			context.Background(),
			ListProjectFilesParams{
				ActorID:   "actor-1",
				ProjectID: "project-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errUserRepo) {
			t.Fatalf(
				"expected user repository error, got %v",
				err,
			)
		}
	})

	t.Run("propagates count error", func(t *testing.T) {
		s, deps := newTestService()

		deps.projectRepo.project = newTestProject(
			t,
			"project-1",
			"actor-1",
			"actor-1",
		)

		deps.fileRepo.countByProjectFn = func(
			context.Context,
			domainmessage.ProjectFileFilter,
		) (int, error) {
			return 0, errFileRepo
		}

		_, err := s.ListProjectFiles(
			context.Background(),
			ListProjectFilesParams{
				ActorID:   "actor-1",
				ProjectID: "project-1",
				Limit:     10,
				Offset:    0,
			},
		)

		if !errors.Is(err, errFileRepo) {
			t.Fatalf(
				"expected file repository error, got %v",
				err,
			)
		}
	})
}

func TestServicePublishEvents(t *testing.T) {
	t.Run("publishes every event", func(t *testing.T) {
		s, deps := newTestService()

		events := []common.Event{
			nil,
			nil,
		}

		// publishEvents only calls EventType when publication fails,
		// so successful publications can be represented by nil events.
		s.publishEvents(context.Background(), events)

		if deps.internalBus.calls != 2 {
			t.Fatalf(
				"expected two event publications, got %d",
				deps.internalBus.calls,
			)
		}
	})
}
