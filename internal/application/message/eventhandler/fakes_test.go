package eventhandler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"mizu/internal/application/eventbus"
	"mizu/internal/domain/billing"
	"mizu/internal/domain/common"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

var (
	errChannelRepo = errors.New("channel repository error")
	errMessageRepo = errors.New("message repository error")
	errUserRepo    = errors.New("user repository error")
	errExternalBus = errors.New("external bus error")
	errIDGenerator = errors.New("id generator error")
	errUOW         = errors.New("unit of work error")
)

type fakeChannelRepository struct {
	getFn           func(context.Context, message.ChannelID) (*message.Channel, error)
	listByProjectFn func(context.Context, project.ProjectID) ([]*message.Channel, error)
	saveFn          func(context.Context, *message.Channel) error
	addFn           func(context.Context, *message.Channel) error

	getCalls           int
	listByProjectCalls int
	saveCalls          int
	addCalls           int

	lastGetID           message.ChannelID
	lastListByProjectID project.ProjectID
	lastSaved           *message.Channel
	lastAdded           *message.Channel

	channel  *message.Channel
	channels []*message.Channel
}

func (f *fakeChannelRepository) Add(
	ctx context.Context,
	c *message.Channel,
) error {
	f.addCalls++
	f.lastAdded = c

	if f.addFn != nil {
		return f.addFn(ctx, c)
	}

	return nil
}

func (f *fakeChannelRepository) Get(
	ctx context.Context,
	id message.ChannelID,
) (*message.Channel, error) {
	f.getCalls++
	f.lastGetID = id

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	return f.channel, nil
}

func (f *fakeChannelRepository) ListByProject(
	ctx context.Context,
	projectID project.ProjectID,
) ([]*message.Channel, error) {
	f.listByProjectCalls++
	f.lastListByProjectID = projectID

	if f.listByProjectFn != nil {
		return f.listByProjectFn(ctx, projectID)
	}

	return f.channels, nil
}

func (f *fakeChannelRepository) ListByMember(
	ctx context.Context,
	memberID iam.UserID,
) ([]*message.Channel, error) {
	return nil, nil
}

func (f *fakeChannelRepository) Save(
	ctx context.Context,
	c *message.Channel,
) error {
	f.saveCalls++
	f.lastSaved = c

	if f.saveFn != nil {
		return f.saveFn(ctx, c)
	}

	return nil
}

type fakeMessageRepository struct {
	addFn func(context.Context, *message.Message) error

	getFn           func(context.Context, message.MessageID) (*message.Message, error)
	listByChannelFn func(context.Context, message.ChannelID, *message.MessageID, int) ([]*message.Message, error)

	addCalls           int
	getCalls           int
	listByChannelCalls int

	lastAdded *message.Message
	lastGetID message.MessageID

	lastListChannelID message.ChannelID
	lastBefore        *message.MessageID
	lastLimit         int

	messages []*message.Message
	message  *message.Message
}

func (f *fakeMessageRepository) Add(
	ctx context.Context,
	m *message.Message,
) error {
	f.addCalls++
	f.lastAdded = m

	if f.addFn != nil {
		return f.addFn(ctx, m)
	}

	return nil
}

func (f *fakeMessageRepository) Get(
	ctx context.Context,
	id message.MessageID,
) (*message.Message, error) {
	f.getCalls++
	f.lastGetID = id

	if f.getFn != nil {
		return f.getFn(ctx, id)
	}

	return f.message, nil
}

func (f *fakeMessageRepository) ListByChannel(
	ctx context.Context,
	channelID message.ChannelID,
	before *message.MessageID,
	limit int,
) ([]*message.Message, error) {
	f.listByChannelCalls++
	f.lastListChannelID = channelID
	f.lastBefore = before
	f.lastLimit = limit

	if f.listByChannelFn != nil {
		return f.listByChannelFn(ctx, channelID, before, limit)
	}

	return f.messages, nil
}

type fakeUserRepository struct {
	getByIDFn func(context.Context, iam.UserID) (*iam.User, error)

	addFn                func(context.Context, *iam.User) error
	existsFn             func(context.Context, iam.UserID) (bool, error)
	existsAllFn          func(context.Context, []iam.UserID) (bool, error)
	getByEmailFn         func(context.Context, iam.Email) (*iam.User, error)
	listFn               func(context.Context, iam.UserFilter, common.Page) ([]*iam.User, error)
	listByIDsFn          func(context.Context, []iam.UserID, iam.UserFilter) ([]*iam.User, error)
	isAnyAdministratorFn func(context.Context, []iam.UserID) (bool, error)
	hasAdministratorFn   func(context.Context) (bool, error)
	countFn              func(context.Context, iam.UserFilter) (int, error)
	saveFn               func(context.Context, *iam.User) error
	removeFn             func(context.Context, *iam.User) error

	getByIDCalls int
	addCalls     int

	lastGetByID iam.UserID
	lastAdded   *iam.User

	user  *iam.User
	users []*iam.User
}

func (f *fakeUserRepository) Add(
	ctx context.Context,
	user *iam.User,
) error {
	f.addCalls++

	if f.addFn != nil {
		return f.addFn(ctx, user)
	}

	return nil
}

func (f *fakeUserRepository) Exists(
	ctx context.Context,
	id iam.UserID,
) (bool, error) {
	if f.existsFn != nil {
		return f.existsFn(ctx, id)
	}

	return false, nil
}

func (f *fakeUserRepository) ExistsAll(
	ctx context.Context,
	ids []iam.UserID,
) (bool, error) {
	if f.existsAllFn != nil {
		return f.existsAllFn(ctx, ids)
	}

	return false, nil
}

func (f *fakeUserRepository) GetByID(
	ctx context.Context,
	id iam.UserID,
) (*iam.User, error) {
	f.getByIDCalls++
	f.lastGetByID = id

	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	return f.user, nil
}

func (f *fakeUserRepository) GetByEmail(
	ctx context.Context,
	email iam.Email,
) (*iam.User, error) {
	if f.getByEmailFn != nil {
		return f.getByEmailFn(ctx, email)
	}

	return nil, nil
}

func (f *fakeUserRepository) List(
	ctx context.Context,
	filter iam.UserFilter,
	page common.Page,
) ([]*iam.User, error) {
	if f.listFn != nil {
		return f.listFn(ctx, filter, page)
	}

	return f.users, nil
}

func (f *fakeUserRepository) ListByIDs(
	ctx context.Context,
	ids []iam.UserID,
	filter iam.UserFilter,
) ([]*iam.User, error) {
	if f.listByIDsFn != nil {
		return f.listByIDsFn(ctx, ids, filter)
	}

	return f.users, nil
}

func (f *fakeUserRepository) IsAnyAdministrator(
	ctx context.Context,
	ids []iam.UserID,
) (bool, error) {
	if f.isAnyAdministratorFn != nil {
		return f.isAnyAdministratorFn(ctx, ids)
	}

	return false, nil
}

func (f *fakeUserRepository) HasAdministrator(
	ctx context.Context,
) (bool, error) {
	if f.hasAdministratorFn != nil {
		return f.hasAdministratorFn(ctx)
	}

	return false, nil
}

func (f *fakeUserRepository) Count(
	ctx context.Context,
	filter iam.UserFilter,
) (int, error) {
	if f.countFn != nil {
		return f.countFn(ctx, filter)
	}

	return 0, nil
}

func (f *fakeUserRepository) Save(
	ctx context.Context,
	user *iam.User,
) error {
	if f.saveFn != nil {
		return f.saveFn(ctx, user)
	}

	return nil
}

func (f *fakeUserRepository) Remove(
	ctx context.Context,
	user *iam.User,
) error {
	if f.removeFn != nil {
		return f.removeFn(ctx, user)
	}

	return nil
}

type fakeExternalBus struct {
	publishFn func(context.Context, eventbus.Event) error

	publishCalls int
	events       []eventbus.Event

	mu sync.Mutex
}

func (f *fakeExternalBus) Publish(
	ctx context.Context,
	event eventbus.Event,
) error {
	f.mu.Lock()
	f.publishCalls++
	f.events = append(f.events, event)
	f.mu.Unlock()

	if f.publishFn != nil {
		return f.publishFn(ctx, event)
	}

	return nil
}

func (f *fakeExternalBus) Subscribe(
	eventType eventbus.EventType,
	handler eventbus.ExtEventHandler,
) {
	// Not needed by these tests.
}

type fakeIDGenerator struct {
	generateFn func() (string, error)

	generateCalls int
	ids           []string
	index         int
}

func (f *fakeIDGenerator) Generate() (string, error) {
	f.generateCalls++

	if f.generateFn != nil {
		return f.generateFn()
	}

	if f.index < len(f.ids) {
		id := f.ids[f.index]
		f.index++
		return id, nil
	}

	return "generated-id", nil
}

type fakeUnitOfWork struct {
	executeFn func(context.Context, func(context.Context) error) error

	executeCalls  int
	callbackCalls int
}

func (f *fakeUnitOfWork) Execute(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	f.executeCalls++

	if f.executeFn != nil {
		return f.executeFn(ctx, fn)
	}

	f.callbackCalls++
	return fn(ctx)
}

type testDeps struct {
	channelRepo *fakeChannelRepository
	messageRepo *fakeMessageRepository
	userRepo    *fakeUserRepository
	externalBus *fakeExternalBus
	idGen       *fakeIDGenerator
	uow         *fakeUnitOfWork
}

func newTestDeps() *testDeps {
	return &testDeps{
		channelRepo: &fakeChannelRepository{},
		messageRepo: &fakeMessageRepository{},
		userRepo:    &fakeUserRepository{},
		externalBus: &fakeExternalBus{},
		idGen:       &fakeIDGenerator{},
		uow:         &fakeUnitOfWork{},
	}
}

func newTestSystemMessagePublisher(
	deps *testDeps,
) *SystemMessagePublisher {
	return NewSystemMessagePublisher(
		deps.channelRepo,
		deps.messageRepo,
		deps.externalBus,
		deps.uow,
		deps.idGen,
	)
}

func newTestContractCreatedEvent(
	t *testing.T,
) contract.ContractCreatedEvent {
	t.Helper()

	return contract.ContractCreatedEvent{
		OccurredAt: time.Now(),
	}
}

func newTestContractStatusChangedEvent(
	t *testing.T,
) contract.ContractStatusChangedEvent {
	t.Helper()

	return contract.ContractStatusChangedEvent{
		OccurredAt: time.Now(),
	}
}

func newTestFileCreatedEvent(
	t *testing.T,
) message.FileCreatedEvent {
	t.Helper()

	return message.FileCreatedEvent{
		OccurredAt: time.Now(),
	}
}

func newTestInvoiceConvertedEvent(
	t *testing.T,
) billing.InvoiceConvertedEvent {
	t.Helper()

	return billing.InvoiceConvertedEvent{
		OccurredAt: time.Now(),
	}
}

func newTestInvoiceCreatedEvent(
	t *testing.T,
) billing.InvoiceCreatedEvent {
	t.Helper()

	return billing.InvoiceCreatedEvent{
		OccurredAt: time.Now(),
	}
}

func newTestInvoiceStatusChangedEvent(
	t *testing.T,
) billing.InvoiceStatusChangedEvent {
	t.Helper()

	return billing.InvoiceStatusChangedEvent{
		OccurredAt: time.Now(),
	}
}

func newTestUserCreatedEvent(
	t *testing.T,
) iam.UserCreatedEvent {
	t.Helper()

	return iam.UserCreatedEvent{
		OccurredAt: time.Now(),
	}
}

func newTestProjectCreatedEvent(
	t *testing.T,
) project.ProjectCreatedEvent {
	t.Helper()

	return project.ProjectCreatedEvent{
		OccurredAt: time.Now(),
	}
}

func assertEventType(
	t *testing.T,
	event common.Event,
	expected common.EventType,
) {
	t.Helper()

	if got := event.EventType(); got != expected {
		t.Fatalf(
			"unexpected event type: got %q, want %q",
			got,
			expected,
		)
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

	firstName, err := iam.NewName("Test", "User")
	if err != nil {
		t.Fatal(err)
	}

	email, err := iam.NewEmail(id + "@example.com")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	return iam.NewUser(
		userID,
		firstName,
		email,
		nil,
		role,
		true,
		userID,
		now,
	)
}
