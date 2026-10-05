package iam

import (
	"testing"
	"time"
)

func testUserDependencies(t *testing.T) (
	UserID,
	Name,
	Email,
	Role,
) {
	t.Helper()

	id, err := NewUserID("user-123")
	if err != nil {
		t.Fatalf("unexpected error creating user ID: %v", err)
	}

	name, err := NewName("John", "Doe")
	if err != nil {
		t.Fatalf("unexpected error creating name: %v", err)
	}

	email, err := NewEmail("john@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating email: %v", err)
	}

	role, err := NewRole("staff")
	if err != nil {
		t.Fatalf("unexpected error creating role: %v", err)
	}

	return id, name, email, role
}

func TestNewUser(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	actor := UserID("actor-123")
	title := "Developer"
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		&title,
		role,
		true,
		actor,
		now,
	)

	if user.ID() != id {
		t.Errorf("expected ID %q, got %q", id, user.ID())
	}

	if user.FirstName() != "John" {
		t.Errorf("expected first name %q, got %q", "John", user.FirstName())
	}

	if user.LastName() != "Doe" {
		t.Errorf("expected last name %q, got %q", "Doe", user.LastName())
	}

	if user.Email() != email {
		t.Errorf("expected email %q, got %q", email, user.Email())
	}

	if user.Title() == nil || *user.Title() != title {
		t.Errorf("expected title %q", title)
	}

	if user.Role() != role {
		t.Errorf("expected role %q, got %q", role, user.Role())
	}

	if !user.IsActive() {
		t.Error("expected user to be active")
	}

	if user.IsVerified() {
		t.Error("expected user to be unverified")
	}

	if user.Image() != nil {
		t.Error("expected new user to have no image")
	}

	if user.LastSignInAt() != nil {
		t.Error("expected new user to have no last sign-in time")
	}

	if user.Version() != 1 {
		t.Errorf("expected version 1, got %d", user.Version())
	}

	if !user.CreatedAt().Equal(now) {
		t.Errorf("expected createdAt %v, got %v", now, user.CreatedAt())
	}

	if !user.UpdatedAt().Equal(now) {
		t.Errorf("expected updatedAt %v, got %v", now, user.UpdatedAt())
	}
}

func TestNewUserCreatesUserCreatedEvent(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	actor := UserID("actor-123")
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		actor,
		now,
	)

	events := user.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(UserCreatedEvent)
	if !ok {
		t.Fatalf("expected UserCreatedEvent, got %T", events[0])
	}

	if event.UserID != id {
		t.Errorf("expected user ID %q, got %q", id, event.UserID)
	}

	if event.Name != name {
		t.Errorf("expected name %+v, got %+v", name, event.Name)
	}

	if event.Email != email {
		t.Errorf("expected email %q, got %q", email, event.Email)
	}

	if event.CreatedBy != actor {
		t.Errorf("expected CreatedBy %q, got %q", actor, event.CreatedBy)
	}

	if !event.OccurredAt.Equal(now) {
		t.Errorf("expected OccurredAt %v, got %v", now, event.OccurredAt)
	}
}

func TestNewSystemUserDoesNotCreateUserCreatedEvent(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewSystemUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		now,
	)

	if events := user.PullEvents(); len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

func TestRestoreUser(t *testing.T) {
	id, name, email, role := testUserDependencies(t)

	title := "Developer"
	imageName := ImageName("avatar.png")
	mimeType := MimeType("image/png")
	image := NewImage(imageName, mimeType)

	lastSignInAt := time.Now()
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()
	version := 7

	user := RestoreUser(
		id,
		name,
		email,
		&title,
		role,
		&image,
		true,
		true,
		&lastSignInAt,
		version,
		createdAt,
		updatedAt,
	)

	if user.ID() != id {
		t.Errorf("expected ID %q, got %q", id, user.ID())
	}

	if user.Title() == nil || *user.Title() != title {
		t.Errorf("expected title %q", title)
	}

	if user.Image() == nil {
		t.Fatal("expected image")
	}

	if user.Image().Name() != imageName {
		t.Errorf("expected image name %q, got %q", imageName, user.Image().Name())
	}

	if user.Image().MimeType() != mimeType {
		t.Errorf("expected MIME type %q, got %q", mimeType, user.Image().MimeType())
	}

	if !user.IsActive() {
		t.Error("expected user to be active")
	}

	if !user.IsVerified() {
		t.Error("expected user to be verified")
	}

	if user.LastSignInAt() == nil {
		t.Fatal("expected last sign-in time")
	}

	if !user.LastSignInAt().Equal(lastSignInAt) {
		t.Errorf(
			"expected last sign-in %v, got %v",
			lastSignInAt,
			*user.LastSignInAt(),
		)
	}

	if user.Version() != version {
		t.Errorf("expected version %d, got %d", version, user.Version())
	}

	if !user.CreatedAt().Equal(createdAt) {
		t.Errorf("expected createdAt %v, got %v", createdAt, user.CreatedAt())
	}

	if !user.UpdatedAt().Equal(updatedAt) {
		t.Errorf("expected updatedAt %v, got %v", updatedAt, user.UpdatedAt())
	}
}

func TestUserActivate(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	initialTime := time.Now()
	updatedTime := initialTime.Add(time.Second)

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		false,
		UserID("actor"),
		initialTime,
	)

	user.PullEvents()

	user.Activate(updatedTime)

	if !user.IsActive() {
		t.Error("expected user to be active")
	}

	if !user.UpdatedAt().Equal(updatedTime) {
		t.Errorf("expected updatedAt %v, got %v", updatedTime, user.UpdatedAt())
	}
}

func TestUserDeactivate(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	initialTime := time.Now()
	updatedTime := initialTime.Add(time.Second)

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		initialTime,
	)

	user.PullEvents()

	user.Deactivate(updatedTime)

	if user.IsActive() {
		t.Error("expected user to be inactive")
	}

	if !user.UpdatedAt().Equal(updatedTime) {
		t.Errorf("expected updatedAt %v, got %v", updatedTime, user.UpdatedAt())
	}
}

func TestUserVerify(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	user.PullEvents()

	verifyTime := now.Add(time.Second)

	user.Verify(verifyTime)

	if !user.IsVerified() {
		t.Error("expected user to be verified")
	}

	if !user.UpdatedAt().Equal(verifyTime) {
		t.Errorf("expected updatedAt %v, got %v", verifyTime, user.UpdatedAt())
	}

	events := user.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(UserVerifiedEvent)
	if !ok {
		t.Fatalf("expected UserVerifiedEvent, got %T", events[0])
	}

	if event.UserID != id {
		t.Errorf("expected user ID %q, got %q", id, event.UserID)
	}

	if !event.OccurredAt.Equal(verifyTime) {
		t.Errorf("expected OccurredAt %v, got %v", verifyTime, event.OccurredAt)
	}
}

func TestUserRecordSignIn(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	user.PullEvents()

	signInTime := now.Add(time.Second)

	user.RecordSignIn(signInTime)

	if user.LastSignInAt() == nil {
		t.Fatal("expected last sign-in time")
	}

	if !user.LastSignInAt().Equal(signInTime) {
		t.Errorf(
			"expected last sign-in time %v, got %v",
			signInTime,
			*user.LastSignInAt(),
		)
	}

	if !user.UpdatedAt().Equal(signInTime) {
		t.Errorf("expected updatedAt %v, got %v", signInTime, user.UpdatedAt())
	}
}

func TestUserChangeName(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	newName, err := NewName("Jane", "Smith")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	changeTime := now.Add(time.Second)

	user.ChangeName(newName, changeTime)

	if user.FirstName() != "Jane" {
		t.Errorf("expected first name %q, got %q", "Jane", user.FirstName())
	}

	if user.LastName() != "Smith" {
		t.Errorf("expected last name %q, got %q", "Smith", user.LastName())
	}

	if !user.UpdatedAt().Equal(changeTime) {
		t.Errorf("expected updatedAt %v, got %v", changeTime, user.UpdatedAt())
	}
}

func TestUserAssignRole(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	newRole := RoleAdministrator
	changeTime := now.Add(time.Second)

	user.AssignRole(newRole, changeTime)

	if user.Role() != newRole {
		t.Errorf("expected role %q, got %q", newRole, user.Role())
	}

	if !user.UpdatedAt().Equal(changeTime) {
		t.Errorf("expected updatedAt %v, got %v", changeTime, user.UpdatedAt())
	}
}

func TestUserChangeEmail(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	user.PullEvents()

	newEmail, err := NewEmail("new@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	changeTime := now.Add(time.Second)

	user.ChangeEmail(newEmail, changeTime)

	if user.Email() != newEmail {
		t.Errorf("expected email %q, got %q", newEmail, user.Email())
	}

	if user.IsVerified() {
		t.Error("expected email change to invalidate verification")
	}

	if !user.UpdatedAt().Equal(changeTime) {
		t.Errorf("expected updatedAt %v, got %v", changeTime, user.UpdatedAt())
	}

	events := user.PullEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event, ok := events[0].(UserEmailChangedEvent)
	if !ok {
		t.Fatalf("expected UserEmailChangedEvent, got %T", events[0])
	}

	if event.UserID != id {
		t.Errorf("expected user ID %q, got %q", id, event.UserID)
	}

	if event.NewEmail != newEmail {
		t.Errorf("expected email %q, got %q", newEmail, event.NewEmail)
	}

	if !event.OccurredAt.Equal(changeTime) {
		t.Errorf("expected OccurredAt %v, got %v", changeTime, event.OccurredAt)
	}
}

func TestUserChangeEmailToSameEmailDoesNothing(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	user.PullEvents()

	changeTime := now.Add(time.Second)

	user.ChangeEmail(email, changeTime)

	if user.Email() != email {
		t.Errorf("expected email to remain %q, got %q", email, user.Email())
	}

	if !user.UpdatedAt().Equal(now) {
		t.Errorf("expected updatedAt to remain %v, got %v", now, user.UpdatedAt())
	}

	if user.IsVerified() {
		t.Error("expected user to remain unverified")
	}

	if events := user.PullEvents(); len(events) != 0 {
		t.Fatalf("expected no events, got %d", len(events))
	}
}

func TestUserChangeTitle(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	title := "Senior Developer"
	changeTime := now.Add(time.Second)

	user.ChangeTitle(&title, changeTime)

	if user.Title() == nil {
		t.Fatal("expected title")
	}

	if *user.Title() != title {
		t.Errorf("expected title %q, got %q", title, *user.Title())
	}

	if !user.UpdatedAt().Equal(changeTime) {
		t.Errorf("expected updatedAt %v, got %v", changeTime, user.UpdatedAt())
	}
}

func TestUserChangeTitleToNil(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	title := "Developer"

	user := NewUser(
		id,
		name,
		email,
		&title,
		role,
		true,
		UserID("actor"),
		now,
	)

	changeTime := now.Add(time.Second)

	user.ChangeTitle(nil, changeTime)

	if user.Title() != nil {
		t.Errorf("expected title to be nil, got %v", *user.Title())
	}
}

func TestUserChangeImage(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	image := NewImage(
		ImageName("avatar.png"),
		MimeType("image/png"),
	)

	changeTime := now.Add(time.Second)

	user.ChangeImage(&image, changeTime)

	if user.Image() == nil {
		t.Fatal("expected image")
	}

	if user.Image().Name() != image.Name() {
		t.Errorf(
			"expected image name %q, got %q",
			image.Name(),
			user.Image().Name(),
		)
	}

	if !user.UpdatedAt().Equal(changeTime) {
		t.Errorf("expected updatedAt %v, got %v", changeTime, user.UpdatedAt())
	}
}

func TestUserChangeImageToNil(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	image := NewImage(
		ImageName("avatar.png"),
		MimeType("image/png"),
	)

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	user.ChangeImage(&image, now)

	changeTime := now.Add(time.Second)

	user.ChangeImage(nil, changeTime)

	if user.Image() != nil {
		t.Error("expected image to be nil")
	}
}

func TestUserPullEventsClearsEvents(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	firstPull := user.PullEvents()

	if len(firstPull) != 1 {
		t.Fatalf("expected first pull to contain 1 event, got %d", len(firstPull))
	}

	secondPull := user.PullEvents()

	if len(secondPull) != 0 {
		t.Fatalf("expected second pull to contain 0 events, got %d", len(secondPull))
	}
}

func TestUserRoleChecks(t *testing.T) {
	id, name, email, _ := testUserDependencies(t)
	now := time.Now()

	tests := []struct {
		name            string
		role            Role
		isAdministrator bool
		isStaff         bool
		isClient        bool
	}{
		{
			name:            "administrator",
			role:            RoleAdministrator,
			isAdministrator: true,
		},
		{
			name:    "staff",
			role:    RoleStaff,
			isStaff: true,
		},
		{
			name:     "client",
			role:     RoleClient,
			isClient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(
				id,
				name,
				email,
				nil,
				tt.role,
				true,
				UserID("actor"),
				now,
			)

			if user.IsAdministrator() != tt.isAdministrator {
				t.Errorf(
					"expected IsAdministrator() = %v, got %v",
					tt.isAdministrator,
					user.IsAdministrator(),
				)
			}

			if user.IsStaff() != tt.isStaff {
				t.Errorf(
					"expected IsStaff() = %v, got %v",
					tt.isStaff,
					user.IsStaff(),
				)
			}

			if user.IsClient() != tt.isClient {
				t.Errorf(
					"expected IsClient() = %v, got %v",
					tt.isClient,
					user.IsClient(),
				)
			}
		})
	}
}

func TestUserEquals(t *testing.T) {
	id, name, email, role := testUserDependencies(t)
	now := time.Now()

	user1 := NewUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		UserID("actor"),
		now,
	)

	user2 := RestoreUser(
		id,
		name,
		email,
		nil,
		role,
		nil,
		true,
		false,
		nil,
		2,
		now,
		now,
	)

	differentID := UserID("different-user")

	user3 := RestoreUser(
		differentID,
		name,
		email,
		nil,
		role,
		nil,
		true,
		false,
		nil,
		2,
		now,
		now,
	)

	if !user1.Equals(user2) {
		t.Error("expected users with the same ID to be equal")
	}

	if user1.Equals(user3) {
		t.Error("expected users with different IDs to not be equal")
	}
}
