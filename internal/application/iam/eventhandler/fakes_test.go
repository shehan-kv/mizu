package eventhandler

import (
	"context"
	"errors"
	"testing"
	"time"

	"mizu/internal/application/mailer"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
)

var (
	errIDGeneration = errors.New("id generation failed")
	errRepository   = errors.New("repository failed")
	errMailDelivery = errors.New("mail delivery failed")
)

type fakeMailer struct {
	sendRecoveryEmailCalls     int
	sendVerificationEmailCalls int
	sendVerifiedEmailCalls     int

	recoveryEmail     iam.Email
	recoveryToken     iam.RecoveryToken
	verificationEmail iam.Email
	verificationID    iam.VerificationID
	verifiedEmail     iam.Email

	sendRecoveryEmailErr     error
	sendVerificationEmailErr error
	sendVerifiedEmailErr     error
}

var _ mailer.Mailer = (*fakeMailer)(nil)

func (f *fakeMailer) SendRecoveryEmail(
	_ context.Context,
	email iam.Email,
	token iam.RecoveryToken,
) error {
	f.sendRecoveryEmailCalls++
	f.recoveryEmail = email
	f.recoveryToken = token

	return f.sendRecoveryEmailErr
}

func (f *fakeMailer) SendVerificationEmail(
	_ context.Context,
	email iam.Email,
	verificationID iam.VerificationID,
) error {
	f.sendVerificationEmailCalls++
	f.verificationEmail = email
	f.verificationID = verificationID

	return f.sendVerificationEmailErr
}

func (f *fakeMailer) SendVerifiedEmail(
	_ context.Context,
	email iam.Email,
) error {
	f.sendVerifiedEmailCalls++
	f.verifiedEmail = email

	return f.sendVerifiedEmailErr
}

func (f *fakeMailer) SendContractEmail(
	_ context.Context,
	_ mailer.ContractEmail,
) error {
	return nil
}

func (f *fakeMailer) SendInvoiceEmail(
	_ context.Context,
	_ mailer.InvoiceEmail,
) error {
	return nil
}

type fakeUserRepository struct {
	user *iam.User

	getByIDCalls int
	getByIDErr   error
}

var _ iam.UserRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) Add(
	_ context.Context,
	_ *iam.User,
) error {
	return nil
}

func (f *fakeUserRepository) Exists(
	_ context.Context,
	_ iam.UserID,
) (bool, error) {
	return false, nil
}

func (f *fakeUserRepository) ExistsAll(
	_ context.Context,
	_ []iam.UserID,
) (bool, error) {
	return false, nil
}

func (f *fakeUserRepository) GetByID(
	_ context.Context,
	_ iam.UserID,
) (*iam.User, error) {
	f.getByIDCalls++

	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}

	return f.user, nil
}

func (f *fakeUserRepository) GetByEmail(
	_ context.Context,
	_ iam.Email,
) (*iam.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) List(
	_ context.Context,
	_ iam.UserFilter,
	_ common.Page,
) ([]*iam.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) ListByIDs(
	_ context.Context,
	_ []iam.UserID,
	_ iam.UserFilter,
) ([]*iam.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) IsAnyAdministrator(
	_ context.Context,
	_ []iam.UserID,
) (bool, error) {
	return false, nil
}

func (f *fakeUserRepository) HasAdministrator(
	_ context.Context,
) (bool, error) {
	return false, nil
}

func (f *fakeUserRepository) Count(
	_ context.Context,
	_ iam.UserFilter,
) (int, error) {
	return 0, nil
}

func (f *fakeUserRepository) Save(
	_ context.Context,
	_ *iam.User,
) error {
	return nil
}

func (f *fakeUserRepository) Remove(
	_ context.Context,
	_ *iam.User,
) error {
	return nil
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

type fakeVerificationRepository struct {
	addCalls int
	addErr   error

	verification *iam.Verification
}

var _ iam.VerificationRepository = (*fakeVerificationRepository)(nil)

func (f *fakeVerificationRepository) Add(_ context.Context, verification *iam.Verification) error {
	f.addCalls++
	f.verification = verification

	return f.addErr
}

func (f *fakeVerificationRepository) Get(
	_ context.Context,
	_ iam.VerificationID,
) (*iam.Verification, error) {
	return nil, nil
}

func (f *fakeVerificationRepository) Remove(
	_ context.Context,
	_ *iam.Verification,
) error {
	return nil
}

func (f *fakeVerificationRepository) RemoveByUserID(_ context.Context, _ iam.UserID) error {
	return nil
}

func newTestUserID(t *testing.T, id string) iam.UserID {
	t.Helper()

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatalf("failed to create test user ID: %v", err)
	}

	return userID
}

func newTestEmail(t *testing.T, email string) iam.Email {
	t.Helper()

	e, err := iam.NewEmail(email)
	if err != nil {
		t.Fatalf("failed to create test email: %v", err)
	}

	return e
}

func newTestRecoveryToken(t *testing.T, token string) iam.RecoveryToken {
	t.Helper()

	recoveryToken, err := iam.NewRecoveryToken(token)
	if err != nil {
		t.Fatalf("failed to create test recovery token: %v", err)
	}

	return recoveryToken
}

func newTestUser(t *testing.T, id string, role iam.Role, isActive bool) *iam.User {
	t.Helper()

	userID := newTestUserID(t, id)
	name, err := iam.NewName("Test", "User")
	if err != nil {
		t.Fatalf("failed to create test name: %v", err)
	}

	email := newTestEmail(t, "user@example.com")
	now := time.Now()

	return iam.NewUser(
		userID,
		name,
		email,
		nil,
		role,
		isActive,
		userID,
		now,
	)
}

func newTestVerificationID(t *testing.T, id string) iam.VerificationID {
	t.Helper()

	verificationID, err := iam.NewVerificationID(id)
	if err != nil {
		t.Fatalf("failed to create test verification ID: %v", err)
	}

	return verificationID
}
