package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

func TestChannelRepository(t *testing.T) {
	db := newTestPostgres(t)
	repo := NewChannelRepository(db)
	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		t.Run("adds channel with project and members", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50000000-0000-0000-0000-000000000001",
			)

			user1ID := newTestContractUser(
				t,
				db,
				"50000000-0000-0000-0000-000000000002",
			)

			user2ID := newTestContractUser(
				t,
				db,
				"50000000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50000000-0000-0000-0000-000000000004",
				&projectID,
				"General Channel 500000000000000000000000000000000004",
				1,
				[]iam.UserID{user1ID, user2ID},
				now,
			)

			if err := repo.Add(ctx, channel); err != nil {
				t.Fatalf("failed to add channel: %v", err)
			}

			got, err := repo.Get(ctx, channel.ID())
			if err != nil {
				t.Fatalf("failed to get added channel: %v", err)
			}

			assertChannelEqual(t, channel, got)
		})

		t.Run("adds channel without project", func(t *testing.T) {
			userID := newTestContractUser(
				t,
				db,
				"50000000-0000-0000-0000-000000000011",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50000000-0000-0000-0000-000000000012",
				nil,
				"Global Channel 500000000000000000000000000000000012",
				1,
				[]iam.UserID{userID},
				now,
			)

			if err := repo.Add(ctx, channel); err != nil {
				t.Fatalf("failed to add channel: %v", err)
			}

			got, err := repo.Get(ctx, channel.ID())
			if err != nil {
				t.Fatalf("failed to get added channel: %v", err)
			}

			assertChannelEqual(t, channel, got)
		})

		t.Run("adds channel without members", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50000000-0000-0000-0000-000000000021",
				nil,
				"Empty Channel 500000000000000000000000000000000021",
				1,
				nil,
				now,
			)

			if err := repo.Add(ctx, channel); err != nil {
				t.Fatalf("failed to add channel: %v", err)
			}

			got, err := repo.Get(ctx, channel.ID())
			if err != nil {
				t.Fatalf("failed to get added channel: %v", err)
			}

			assertChannelEqual(t, channel, got)
		})

		t.Run("rejects duplicate channel id", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50000000-0000-0000-0000-000000000031",
				nil,
				"Duplicate Channel 500000000000000000000000000000000031",
				1,
				nil,
				now,
			)

			if err := repo.Add(ctx, channel); err != nil {
				t.Fatalf("failed to add initial channel: %v", err)
			}

			err := repo.Add(ctx, channel)
			if err == nil {
				t.Fatal("expected duplicate channel error")
			}

		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns channel with members", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50100000-0000-0000-0000-000000000001",
			)

			user1ID := newTestContractUser(
				t,
				db,
				"50100000-0000-0000-0000-000000000002",
			)

			user2ID := newTestContractUser(
				t,
				db,
				"50100000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50100000-0000-0000-0000-000000000004",
				&projectID,
				"Get Channel 501000000000000000000000000000000004",
				3,
				[]iam.UserID{user1ID, user2ID},
				now,
			)

			insertTestChannelDirectly(t, db, channel)

			got, err := repo.Get(ctx, channel.ID())
			if err != nil {
				t.Fatalf("failed to get channel: %v", err)
			}

			assertChannelEqual(t, channel, got)
		})

		t.Run("returns channel without project", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50100000-0000-0000-0000-000000000011",
				nil,
				"Global Get Channel 501000000000000000000000000000000011",
				1,
				nil,
				now,
			)

			insertTestChannelDirectly(t, db, channel)

			got, err := repo.Get(ctx, channel.ID())
			if err != nil {
				t.Fatalf("failed to get channel: %v", err)
			}

			assertChannelEqual(t, channel, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			channelID, err := message.NewChannelID(
				"50100000-0000-0000-0000-000000000021",
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.Get(ctx, channelID)
			if err == nil {
				t.Fatal("expected not found error")
			}

			if !errors.Is(err, message.ErrChannelNotFound) {
				t.Fatalf(
					"expected ErrChannelNotFound, got %v",
					err,
				)
			}
		})
	})

	t.Run("ListByProject", func(t *testing.T) {
		t.Run("lists channels for project", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50200000-0000-0000-0000-000000000001",
			)

			otherProjectID := newTestContractProject(
				t,
				db,
				"50200000-0000-0000-0000-000000000002",
			)

			userID := newTestContractUser(
				t,
				db,
				"50200000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestChannel(
				t,
				"50200000-0000-0000-0000-000000000004",
				&projectID,
				"Project Channel One 502000000000000000000000000000000004",
				1,
				[]iam.UserID{userID},
				now,
			)

			c2 := newTestChannel(
				t,
				"50200000-0000-0000-0000-000000000005",
				&projectID,
				"Project Channel Two 502000000000000000000000000000000005",
				1,
				nil,
				now.Add(time.Second),
			)

			other := newTestChannel(
				t,
				"50200000-0000-0000-0000-000000000006",
				&otherProjectID,
				"Other Project Channel 502000000000000000000000000000000006",
				1,
				nil,
				now.Add(2*time.Second),
			)

			insertTestChannelDirectly(t, db, c1)
			insertTestChannelDirectly(t, db, c2)
			insertTestChannelDirectly(t, db, other)

			channels, err := repo.ListByProject(ctx, projectID)
			if err != nil {
				t.Fatalf("failed to list channels: %v", err)
			}

			if len(channels) != 2 {
				t.Fatalf("expected 2 channels, got %d", len(channels))
			}

			found := map[message.ChannelID]bool{}
			for _, channel := range channels {
				found[channel.ID()] = true
			}

			if !found[c1.ID()] || !found[c2.ID()] {
				t.Fatal("expected both project channels")
			}

			if found[other.ID()] {
				t.Fatal("unexpected channel from another project")
			}
		})

		t.Run("returns empty slice when project has no channels", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50200000-0000-0000-0000-000000000011",
			)

			channels, err := repo.ListByProject(ctx, projectID)
			if err != nil {
				t.Fatalf("failed to list channels: %v", err)
			}

			if channels == nil {
				t.Fatal("expected non-nil empty slice")
			}

			if len(channels) != 0 {
				t.Fatalf("expected 0 channels, got %d", len(channels))
			}
		})
	})

	t.Run("ListByMember", func(t *testing.T) {
		t.Run("lists channels for member", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50300000-0000-0000-0000-000000000001",
			)

			userID := newTestContractUser(
				t,
				db,
				"50300000-0000-0000-0000-000000000002",
			)

			otherUserID := newTestContractUser(
				t,
				db,
				"50300000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			c1 := newTestChannel(
				t,
				"50300000-0000-0000-0000-000000000004",
				&projectID,
				"Member Channel One 503000000000000000000000000000000004",
				1,
				[]iam.UserID{userID},
				now,
			)

			c2 := newTestChannel(
				t,
				"50300000-0000-0000-0000-000000000005",
				&projectID,
				"Member Channel Two 503000000000000000000000000000000005",
				1,
				[]iam.UserID{userID, otherUserID},
				now.Add(time.Second),
			)

			c3 := newTestChannel(
				t,
				"50300000-0000-0000-0000-000000000006",
				&projectID,
				"Other Member Channel 503000000000000000000000000000000006",
				1,
				[]iam.UserID{otherUserID},
				now.Add(2*time.Second),
			)

			insertTestChannelDirectly(t, db, c1)
			insertTestChannelDirectly(t, db, c2)
			insertTestChannelDirectly(t, db, c3)

			channels, err := repo.ListByMember(ctx, userID)
			if err != nil {
				t.Fatalf("failed to list channels: %v", err)
			}

			if len(channels) != 2 {
				t.Fatalf("expected 2 channels, got %d", len(channels))
			}

			found := map[message.ChannelID]bool{}
			for _, channel := range channels {
				found[channel.ID()] = true
			}

			if !found[c1.ID()] || !found[c2.ID()] {
				t.Fatal("expected both channels for member")
			}

			if found[c3.ID()] {
				t.Fatal("unexpected channel where user is not a member")
			}
		})

		t.Run("returns empty slice when member has no channels", func(t *testing.T) {
			userID := newTestContractUser(
				t,
				db,
				"50300000-0000-0000-0000-000000000011",
			)

			channels, err := repo.ListByMember(ctx, userID)
			if err != nil {
				t.Fatalf("failed to list channels: %v", err)
			}

			if channels == nil {
				t.Fatal("expected non-nil empty slice")
			}

			if len(channels) != 0 {
				t.Fatalf("expected 0 channels, got %d", len(channels))
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("updates channel and replaces members", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50400000-0000-0000-0000-000000000001",
			)

			user1ID := newTestContractUser(
				t,
				db,
				"50400000-0000-0000-0000-000000000002",
			)

			user2ID := newTestContractUser(
				t,
				db,
				"50400000-0000-0000-0000-000000000003",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50400000-0000-0000-0000-000000000004",
				&projectID,
				"Original Channel 504000000000000000000000000000000004",
				1,
				[]iam.UserID{user1ID},
				now,
			)

			insertTestChannelDirectly(t, db, channel)

			updatedAt := now.Add(time.Minute)

			updatedName, err := message.NewChannelName(
				"Updated Channel 504000000000000000000000000000000004",
			)
			if err != nil {
				t.Fatal(err)
			}

			updated := message.RestoreChannel(
				channel.ID(),
				channel.ProjectID(),
				updatedName,
				[]iam.UserID{user2ID},
				channel.Version(),
				channel.CreatedAt(),
				updatedAt,
			)

			if err := repo.Save(ctx, updated); err != nil {
				t.Fatalf("failed to save channel: %v", err)
			}

			got, err := repo.Get(ctx, channel.ID())
			if err != nil {
				t.Fatalf("failed to get saved channel: %v", err)
			}

			if got.Name().String() != updatedName.String() {
				t.Fatalf(
					"expected name %s, got %s",
					updatedName.String(),
					got.Name().String(),
				)
			}

			if got.Version() != channel.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					channel.Version()+1,
					got.Version(),
				)
			}

			if got.UpdatedAt() != updatedAt {
				t.Fatalf(
					"expected updated_at %v, got %v",
					updatedAt,
					got.UpdatedAt(),
				)
			}

			assertChannelMembers(
				t,
				got,
				[]iam.UserID{user2ID},
			)
		})

		t.Run("updates project", func(t *testing.T) {
			project1ID := newTestContractProject(
				t,
				db,
				"50400000-0000-0000-0000-000000000011",
			)

			project2ID := newTestContractProject(
				t,
				db,
				"50400000-0000-0000-0000-000000000012",
			)

			userID := newTestContractUser(
				t,
				db,
				"50400000-0000-0000-0000-000000000013",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50400000-0000-0000-0000-000000000014",
				&project1ID,
				"Move Channel 504000000000000000000000000000000014",
				1,
				[]iam.UserID{userID},
				now,
			)

			insertTestChannelDirectly(t, db, channel)

			updatedAt := now.Add(time.Minute)

			updated := message.RestoreChannel(
				channel.ID(),
				&project2ID,
				channel.Name(),
				[]iam.UserID{userID},
				channel.Version(),
				channel.CreatedAt(),
				updatedAt,
			)

			if err := repo.Save(ctx, updated); err != nil {
				t.Fatalf("failed to save channel: %v", err)
			}

			got, err := repo.Get(ctx, channel.ID())
			if err != nil {
				t.Fatalf("failed to get saved channel: %v", err)
			}

			if got.ProjectID() == nil {
				t.Fatal("expected project id")
			}

			if got.ProjectID().String() != project2ID.String() {
				t.Fatalf(
					"expected project %s, got %s",
					project2ID,
					got.ProjectID(),
				)
			}

			assertChannelMembers(
				t,
				got,
				[]iam.UserID{userID},
			)

			if got.Version() != channel.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					channel.Version()+1,
					got.Version(),
				)
			}

			if !got.UpdatedAt().Equal(updatedAt) {
				t.Fatalf(
					"expected updated_at %v, got %v",
					updatedAt,
					got.UpdatedAt(),
				)
			}
		})

		t.Run("removes project", func(t *testing.T) {
			projectID := newTestContractProject(
				t,
				db,
				"50400000-0000-0000-0000-000000000021",
			)

			userID := newTestContractUser(
				t,
				db,
				"50400000-0000-0000-0000-000000000022",
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50400000-0000-0000-0000-000000000023",
				&projectID,
				"Remove Project Channel 504000000000000000000000000000000023",
				1,
				[]iam.UserID{userID},
				now,
			)

			insertTestChannelDirectly(t, db, channel)

			updatedAt := now.Add(time.Minute)

			updated := message.RestoreChannel(
				channel.ID(),
				nil,
				channel.Name(),
				[]iam.UserID{userID},
				channel.Version(),
				channel.CreatedAt(),
				updatedAt,
			)

			if err := repo.Save(ctx, updated); err != nil {
				t.Fatalf("failed to save channel: %v", err)
			}

			got, err := repo.Get(ctx, channel.ID())
			if err != nil {
				t.Fatalf("failed to get saved channel: %v", err)
			}

			if got.ProjectID() != nil {
				t.Fatal("expected project id to be nil")
			}

			assertChannelMembers(
				t,
				got,
				[]iam.UserID{userID},
			)

			if got.Version() != channel.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					channel.Version()+1,
					got.Version(),
				)
			}

			if !got.UpdatedAt().Equal(updatedAt) {
				t.Fatalf(
					"expected updated_at %v, got %v",
					updatedAt,
					got.UpdatedAt(),
				)
			}
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			channel := newTestChannel(
				t,
				"50400000-0000-0000-0000-000000000031",
				nil,
				"Concurrent Channel 504000000000000000000000000000000031",
				2,
				nil,
				now,
			)

			insertTestChannelDirectly(t, db, channel)

			updatedName, err := message.NewChannelName(
				"Concurrent Updated Channel 504000000000000000000000000000031",
			)
			if err != nil {
				t.Fatal(err)
			}

			updated := message.RestoreChannel(
				channel.ID(),
				nil,
				updatedName,
				nil,
				1,
				channel.CreatedAt(),
				now.Add(time.Minute),
			)

			err = repo.Save(ctx, updated)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(
				err,
				message.ErrChannelConcurrentModification,
			) {
				t.Fatalf(
					"expected ErrChannelConcurrentModification, got %v",
					err,
				)
			}
		})
	})
}

func newTestChannel(
	t *testing.T,
	id string,
	projectID *project.ProjectID,
	name string,
	version int,
	members []iam.UserID,
	createdAt time.Time,
) *message.Channel {
	t.Helper()

	channelID, err := message.NewChannelID(id)
	if err != nil {
		t.Fatal(err)
	}

	channelName, err := message.NewChannelName(name)
	if err != nil {
		t.Fatal(err)
	}

	return message.RestoreChannel(
		channelID,
		projectID,
		channelName,
		members,
		version,
		createdAt,
		createdAt,
	)
}

func insertTestChannelDirectly(
	t *testing.T,
	db *sql.DB,
	channel *message.Channel,
) {
	t.Helper()

	var projectID any
	if channel.ProjectID() != nil {
		projectID = channel.ProjectID().String()
	}

	_, err := db.Exec(
		`INSERT INTO channels (
			id,
			project_id,
			name,
			version,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		channel.ID().String(),
		projectID,
		channel.Name().String(),
		channel.Version(),
		channel.CreatedAt(),
		channel.UpdatedAt(),
	)
	if err != nil {
		t.Fatalf("failed to insert test channel: %v", err)
	}

	for _, memberID := range channel.Members() {
		_, err := db.Exec(
			`INSERT INTO channel_members (
				channel_id,
				user_id
			) VALUES ($1, $2)`,
			channel.ID().String(),
			memberID.String(),
		)
		if err != nil {
			t.Fatalf("failed to insert test channel member: %v", err)
		}
	}
}

func assertChannelEqual(
	t *testing.T,
	expected *message.Channel,
	actual *message.Channel,
) {
	t.Helper()

	if actual.ID() != expected.ID() {
		t.Fatalf(
			"expected id %s, got %s",
			expected.ID(),
			actual.ID(),
		)
	}

	if expected.ProjectID() == nil {
		if actual.ProjectID() != nil {
			t.Fatal("expected nil project id")
		}
	} else {
		if actual.ProjectID() == nil {
			t.Fatal("expected project id, got nil")
		}

		if actual.ProjectID().String() != expected.ProjectID().String() {
			t.Fatalf(
				"expected project id %s, got %s",
				expected.ProjectID(),
				actual.ProjectID(),
			)
		}
	}

	if actual.Name().String() != expected.Name().String() {
		t.Fatalf(
			"expected name %s, got %s",
			expected.Name(),
			actual.Name(),
		)
	}

	if actual.Version() != expected.Version() {
		t.Fatalf(
			"expected version %d, got %d",
			expected.Version(),
			actual.Version(),
		)
	}

	if !actual.CreatedAt().Equal(expected.CreatedAt()) {
		t.Fatalf(
			"expected created_at %v, got %v",
			expected.CreatedAt(),
			actual.CreatedAt(),
		)
	}

	if !actual.UpdatedAt().Equal(expected.UpdatedAt()) {
		t.Fatalf(
			"expected updated_at %v, got %v",
			expected.UpdatedAt(),
			actual.UpdatedAt(),
		)
	}

	assertChannelMembers(t, actual, expected.Members())
}

func assertChannelMembers(
	t *testing.T,
	channel *message.Channel,
	expected []iam.UserID,
) {
	t.Helper()

	actual := channel.Members()

	if len(actual) != len(expected) {
		t.Fatalf(
			"expected %d members, got %d",
			len(expected),
			len(actual),
		)
	}

	expectedSet := make(map[iam.UserID]bool, len(expected))
	for _, userID := range expected {
		expectedSet[userID] = true
	}

	actualSet := make(map[iam.UserID]bool, len(actual))
	for _, userID := range actual {
		actualSet[userID] = true
	}

	for userID := range expectedSet {
		if !actualSet[userID] {
			t.Fatalf("expected member %s", userID)
		}
	}
}
