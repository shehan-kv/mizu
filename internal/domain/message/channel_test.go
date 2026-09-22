package message

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

func newTestChannel(t *testing.T, now time.Time) (*Channel, iam.UserID, iam.UserID, iam.UserID) {
	t.Helper()

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := NewChannelName("general")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := iam.NewUserID("user-2")
	if err != nil {
		t.Fatal(err)
	}

	user3, err := iam.NewUserID("user-3")
	if err != nil {
		t.Fatal(err)
	}

	channel, err := NewProjectChannel(
		channelID,
		projectID,
		name,
		[]iam.UserID{user1, user2},
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	return channel, user1, user2, user3
}

func TestNewChannel(t *testing.T) {
	now := time.Now()

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := NewChannelName("general")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := iam.NewUserID("user-2")
	if err != nil {
		t.Fatal(err)
	}

	channel, err := NewChannel(
		channelID,
		&projectID,
		name,
		[]iam.UserID{user1, user2},
		now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if channel.ID() != channelID {
		t.Fatalf("expected ID %q, got %q", channelID, channel.ID())
	}

	if channel.ProjectID() == nil {
		t.Fatal("expected project ID")
	}

	if *channel.ProjectID() != projectID {
		t.Fatalf("expected project ID %q, got %q", projectID, *channel.ProjectID())
	}

	if channel.Name() != name {
		t.Fatalf("expected name %q, got %q", name, channel.Name())
	}

	if !channel.HasMember(user1) || !channel.HasMember(user2) {
		t.Fatal("expected both members to be present")
	}

	if channel.Version() != 1 {
		t.Fatalf("expected version 1, got %d", channel.Version())
	}

	if !channel.CreatedAt().Equal(now) {
		t.Fatalf("expected createdAt %v, got %v", now, channel.CreatedAt())
	}

	if !channel.UpdatedAt().Equal(now) {
		t.Fatalf("expected updatedAt %v, got %v", now, channel.UpdatedAt())
	}
}

func TestNewChannelRequiresAtLeastTwoMembers(t *testing.T) {
	now := time.Now()

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := NewChannelName("general")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewChannel(
		channelID,
		nil,
		name,
		[]iam.UserID{user1},
		now,
	)

	if err != ErrChannelMustHaveAtLeastTwoMembers {
		t.Fatalf("expected %v, got %v", ErrChannelMustHaveAtLeastTwoMembers, err)
	}
}

func TestNewChannelRejectsDuplicateMembers(t *testing.T) {
	now := time.Now()

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := NewChannelName("general")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewChannel(
		channelID,
		nil,
		name,
		[]iam.UserID{user1, user1},
		now,
	)

	if err != ErrChannelMembersMustBeUnique {
		t.Fatalf("expected %v, got %v", ErrChannelMembersMustBeUnique, err)
	}
}

func TestNewProjectChannel(t *testing.T) {
	now := time.Now()

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := NewChannelName("project")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := iam.NewUserID("user-2")
	if err != nil {
		t.Fatal(err)
	}

	channel, err := NewProjectChannel(
		channelID,
		projectID,
		name,
		[]iam.UserID{user1, user2},
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	if channel.ProjectID() == nil {
		t.Fatal("expected project ID")
	}

	if *channel.ProjectID() != projectID {
		t.Fatalf("expected project ID %q, got %q", projectID, *channel.ProjectID())
	}
}

func TestNewUserChannel(t *testing.T) {
	now := time.Now()

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := NewChannelName("direct")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := iam.NewUserID("user-2")
	if err != nil {
		t.Fatal(err)
	}

	channel, err := NewUserChannel(
		channelID,
		name,
		[]iam.UserID{user1, user2},
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	if channel.ProjectID() != nil {
		t.Fatal("expected user channel to have no project ID")
	}
}

func TestRestoreChannel(t *testing.T) {
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Second)

	channelID, err := NewChannelID("channel-1")
	if err != nil {
		t.Fatal(err)
	}

	projectID, err := project.NewProjectID("project-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := NewChannelName("general")
	if err != nil {
		t.Fatal(err)
	}

	user1, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	user2, err := iam.NewUserID("user-2")
	if err != nil {
		t.Fatal(err)
	}

	channel := RestoreChannel(
		channelID,
		&projectID,
		name,
		[]iam.UserID{user1, user2},
		7,
		createdAt,
		updatedAt,
	)

	if channel.Version() != 7 {
		t.Fatalf("expected version 7, got %d", channel.Version())
	}

	if !channel.CreatedAt().Equal(createdAt) {
		t.Fatalf("expected createdAt %v, got %v", createdAt, channel.CreatedAt())
	}

	if !channel.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, channel.UpdatedAt())
	}
}

func TestChannelHasMember(t *testing.T) {
	now := time.Now()
	channel, user1, _, user3 := newTestChannel(t, now)

	if !channel.HasMember(user1) {
		t.Fatal("expected user1 to be a member")
	}

	if channel.HasMember(user3) {
		t.Fatal("expected user3 not to be a member")
	}
}

func TestChannelAddMember(t *testing.T) {
	now := time.Now()
	channel, _, _, user3 := newTestChannel(t, now)

	updatedAt := now.Add(time.Second)

	if err := channel.AddMember(user3, updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !channel.HasMember(user3) {
		t.Fatal("expected user3 to be added")
	}

	if !channel.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, channel.UpdatedAt())
	}
}

func TestChannelAddMemberRejectsExistingMember(t *testing.T) {
	now := time.Now()
	channel, user1, _, _ := newTestChannel(t, now)

	if err := channel.AddMember(user1, now.Add(time.Second)); err != ErrChannelMemberAlreadyExists {
		t.Fatalf("expected %v, got %v", ErrChannelMemberAlreadyExists, err)
	}
}

func TestChannelRemoveMember(t *testing.T) {
	now := time.Now()
	channel, user1, user2, user3 := newTestChannel(t, now)

	if err := channel.AddMember(user3, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	updatedAt := now.Add(2 * time.Second)

	if err := channel.RemoveMember(user3, updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if channel.HasMember(user3) {
		t.Fatal("expected user3 to be removed")
	}

	if !channel.HasMember(user1) || !channel.HasMember(user2) {
		t.Fatal("expected original members to remain")
	}

	if !channel.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, channel.UpdatedAt())
	}
}

func TestChannelRemoveMemberRequiresAtLeastTwoMembers(t *testing.T) {
	now := time.Now()
	channel, _, _, _ := newTestChannel(t, now)

	user3, err := iam.NewUserID("user-3")
	if err != nil {
		t.Fatal(err)
	}

	if err := channel.AddMember(user3, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if err := channel.RemoveMember(user3, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}

	// The channel now has exactly two members.
	if err := channel.RemoveMember(user3, now.Add(3*time.Second)); err != ErrChannelMustHaveAtLeastTwoMembers {
		t.Fatalf("expected %v, got %v", ErrChannelMustHaveAtLeastTwoMembers, err)
	}
}

func TestChannelRemoveMemberNotFound(t *testing.T) {
	now := time.Now()
	channel, _, _, user3 := newTestChannel(t, now)

	if err := channel.AddMember(user3, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	unknownUser, err := iam.NewUserID("user-4")
	if err != nil {
		t.Fatal(err)
	}

	if err := channel.RemoveMember(unknownUser, now.Add(2*time.Second)); err != ErrChannelMemberNotFound {
		t.Fatalf("expected %v, got %v", ErrChannelMemberNotFound, err)
	}
}

func TestChannelReplaceMembers(t *testing.T) {
	now := time.Now()
	channel, user1, _, user3 := newTestChannel(t, now)

	updatedAt := now.Add(time.Second)

	if err := channel.ReplaceMembers([]iam.UserID{user1, user3}, updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !channel.HasMember(user1) || !channel.HasMember(user3) {
		t.Fatal("expected replacement members to be present")
	}

	if channel.Members()[0] != user1 || channel.Members()[1] != user3 {
		t.Fatalf("unexpected members: %v", channel.Members())
	}

	if !channel.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updatedAt %v, got %v", updatedAt, channel.UpdatedAt())
	}
}

func TestChannelReplaceMembersRequiresAtLeastTwoMembers(t *testing.T) {
	now := time.Now()
	channel, user1, _, _ := newTestChannel(t, now)

	if err := channel.ReplaceMembers([]iam.UserID{user1}, now.Add(time.Second)); err != ErrChannelMustHaveAtLeastTwoMembers {
		t.Fatalf("expected %v, got %v", ErrChannelMustHaveAtLeastTwoMembers, err)
	}
}

func TestChannelReplaceMembersRejectsDuplicates(t *testing.T) {
	now := time.Now()
	channel, user1, _, _ := newTestChannel(t, now)

	if err := channel.ReplaceMembers([]iam.UserID{user1, user1}, now.Add(time.Second)); err != ErrChannelMembersMustBeUnique {
		t.Fatalf("expected %v, got %v", ErrChannelMembersMustBeUnique, err)
	}
}

func TestChannelMembersReturnsCopy(t *testing.T) {
	now := time.Now()
	channel, user1, user2, _ := newTestChannel(t, now)

	members := channel.Members()
	members[0] = user2

	if !channel.HasMember(user1) {
		t.Fatal("modifying returned members should not modify the channel")
	}
}
