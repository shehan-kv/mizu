package sqlite

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

func TestMessageRepository(t *testing.T) {
	db := newTestSQLite(t)
	repo := NewMessageRepository(db)
	ctx := context.Background()

	user1ID := newTestMessageUser(
		t,
		db,
		"20000000-0000-0000-0000-000000000001",
	)

	user2ID := newTestMessageUser(
		t,
		db,
		"20000000-0000-0000-0000-000000000002",
	)

	t.Run("Add", func(t *testing.T) {
		t.Run("adds message", func(t *testing.T) {
			projectID := newTestMessageProject(
				t,
				db,
				"21000000-0000-0000-0000-000000000001",
			)

			channelID := newTestMessageChannel(
				t,
				db,
				"21000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			msg := newTestMessage(
				t,
				"21000000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				false,
				"Hello, world!",
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, msg); err != nil {
				t.Fatalf("failed to add message: %v", err)
			}

			got, err := repo.Get(ctx, msg.ID())
			if err != nil {
				t.Fatalf("failed to get message: %v", err)
			}

			assertMessageEqual(t, msg, got)
		})

		t.Run("adds system message", func(t *testing.T) {
			projectID := newTestMessageProject(
				t,
				db,
				"21100000-0000-0000-0000-000000000001",
			)

			channelID := newTestMessageChannel(
				t,
				db,
				"21100000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			msg := newTestMessage(
				t,
				"21100000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				true,
				"User joined the channel.",
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, msg); err != nil {
				t.Fatalf("failed to add system message: %v", err)
			}

			got, err := repo.Get(ctx, msg.ID())
			if err != nil {
				t.Fatalf("failed to get message: %v", err)
			}

			assertMessageEqual(t, msg, got)
		})

		t.Run("returns foreign key error for unknown channel", func(t *testing.T) {
			channelID, err := message.NewChannelID(
				"21200000-0000-0000-0000-000000000001",
			)
			if err != nil {
				t.Fatal(err)
			}

			msg := newTestMessage(
				t,
				"21200000-0000-0000-0000-000000000002",
				channelID,
				user1ID,
				false,
				"Hello",
				time.Now().UTC().Truncate(time.Microsecond),
			)

			err = repo.Add(ctx, msg)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns message", func(t *testing.T) {
			projectID := newTestMessageProject(
				t,
				db,
				"22000000-0000-0000-0000-000000000001",
			)

			channelID := newTestMessageChannel(
				t,
				db,
				"22000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			msg := newTestMessage(
				t,
				"22000000-0000-0000-0000-000000000003",
				channelID,
				user2ID,
				false,
				"Test message",
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, msg); err != nil {
				t.Fatalf("failed to add message: %v", err)
			}

			got, err := repo.Get(ctx, msg.ID())
			if err != nil {
				t.Fatalf("failed to get message: %v", err)
			}

			assertMessageEqual(t, msg, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			messageID, err := message.NewMessageID(
				"22000000-0000-0000-0000-000000000010",
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.Get(ctx, messageID)

			if !errors.Is(err, message.ErrMessageNotFound) {
				t.Fatalf(
					"expected ErrMessageNotFound, got %v",
					err,
				)
			}
		})
	})

	t.Run("ListByChannel", func(t *testing.T) {
		t.Run("lists messages by channel", func(t *testing.T) {
			projectID := newTestMessageProject(
				t,
				db,
				"23000000-0000-0000-0000-000000000001",
			)

			channelID := newTestMessageChannel(
				t,
				db,
				"23000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			otherChannelID := newTestMessageChannel(
				t,
				db,
				"23000000-0000-0000-0000-000000000003",
				projectID,
				user1ID,
				user2ID,
			)

			now := time.Now().UTC()

			msg1 := newTestMessage(
				t,
				"23000000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				false,
				"First message",
				now.Truncate(time.Microsecond),
			)

			msg2 := newTestMessage(
				t,
				"23000000-0000-0000-0000-000000000005",
				channelID,
				user2ID,
				false,
				"Second message",
				now.Add(time.Second).Truncate(time.Microsecond),
			)

			otherMsg := newTestMessage(
				t,
				"23000000-0000-0000-0000-000000000006",
				otherChannelID,
				user1ID,
				false,
				"Other channel message",
				now.Add(2*time.Second).Truncate(time.Microsecond),
			)

			for _, msg := range []*message.Message{
				msg1,
				msg2,
				otherMsg,
			} {
				if err := repo.Add(ctx, msg); err != nil {
					t.Fatalf("failed to add message: %v", err)
				}
			}

			messages, err := repo.ListByChannel(
				ctx,
				channelID,
				nil,
				10,
			)
			if err != nil {
				t.Fatalf("failed to list messages: %v", err)
			}

			if len(messages) != 2 {
				t.Fatalf(
					"expected 2 messages, got %d",
					len(messages),
				)
			}

			// ListByChannel orders by ID DESC.
			if messages[0].ID() != msg2.ID() {
				t.Fatalf(
					"expected first message %s, got %s",
					msg2.ID(),
					messages[0].ID(),
				)
			}

			if messages[1].ID() != msg1.ID() {
				t.Fatalf(
					"expected second message %s, got %s",
					msg1.ID(),
					messages[1].ID(),
				)
			}
		})

		t.Run("applies limit", func(t *testing.T) {
			projectID := newTestMessageProject(
				t,
				db,
				"23100000-0000-0000-0000-000000000001",
			)

			channelID := newTestMessageChannel(
				t,
				db,
				"23100000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			msg1 := newTestMessage(
				t,
				"23100000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				false,
				"First",
				time.Now().UTC().Truncate(time.Microsecond),
			)

			msg2 := newTestMessage(
				t,
				"23100000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				false,
				"Second",
				time.Now().UTC().Add(time.Second).Truncate(time.Microsecond),
			)

			msg3 := newTestMessage(
				t,
				"23100000-0000-0000-0000-000000000005",
				channelID,
				user1ID,
				false,
				"Third",
				time.Now().UTC().Add(2*time.Second).Truncate(time.Microsecond),
			)

			for _, msg := range []*message.Message{
				msg1,
				msg2,
				msg3,
			} {
				if err := repo.Add(ctx, msg); err != nil {
					t.Fatalf("failed to add message: %v", err)
				}
			}

			messages, err := repo.ListByChannel(
				ctx,
				channelID,
				nil,
				2,
			)
			if err != nil {
				t.Fatalf("failed to list messages: %v", err)
			}

			if len(messages) != 2 {
				t.Fatalf(
					"expected 2 messages, got %d",
					len(messages),
				)
			}

			if messages[0].ID() != msg3.ID() {
				t.Fatalf(
					"expected first message %s, got %s",
					msg3.ID(),
					messages[0].ID(),
				)
			}

			if messages[1].ID() != msg2.ID() {
				t.Fatalf(
					"expected second message %s, got %s",
					msg2.ID(),
					messages[1].ID(),
				)
			}
		})

		t.Run("applies before cursor", func(t *testing.T) {
			projectID := newTestMessageProject(
				t,
				db,
				"23200000-0000-0000-0000-000000000001",
			)

			channelID := newTestMessageChannel(
				t,
				db,
				"23200000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			msg1 := newTestMessage(
				t,
				"23200000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				false,
				"First",
				time.Now().UTC().Truncate(time.Microsecond),
			)

			msg2 := newTestMessage(
				t,
				"23200000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				false,
				"Second",
				time.Now().UTC().Add(time.Second).Truncate(time.Microsecond),
			)

			msg3 := newTestMessage(
				t,
				"23200000-0000-0000-0000-000000000005",
				channelID,
				user1ID,
				false,
				"Third",
				time.Now().UTC().Add(2*time.Second).Truncate(time.Microsecond),
			)

			for _, msg := range []*message.Message{
				msg1,
				msg2,
				msg3,
			} {
				if err := repo.Add(ctx, msg); err != nil {
					t.Fatalf("failed to add message: %v", err)
				}
			}

			msg3ID := msg3.ID()

			messages, err := repo.ListByChannel(
				ctx,
				channelID,
				&msg3ID,
				10,
			)
			if err != nil {
				t.Fatalf("failed to list messages: %v", err)
			}

			if len(messages) != 2 {
				t.Fatalf(
					"expected 2 messages, got %d",
					len(messages),
				)
			}

			if messages[0].ID() != msg2.ID() {
				t.Fatalf(
					"expected first message %s, got %s",
					msg2.ID(),
					messages[0].ID(),
				)
			}

			if messages[1].ID() != msg1.ID() {
				t.Fatalf(
					"expected second message %s, got %s",
					msg1.ID(),
					messages[1].ID(),
				)
			}
		})

		t.Run("applies before cursor and limit", func(t *testing.T) {
			projectID := newTestMessageProject(
				t,
				db,
				"23300000-0000-0000-0000-000000000001",
			)

			channelID := newTestMessageChannel(
				t,
				db,
				"23300000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			msg1 := newTestMessage(
				t,
				"23300000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				false,
				"First",
				time.Now().UTC().Truncate(time.Microsecond),
			)

			msg2 := newTestMessage(
				t,
				"23300000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				false,
				"Second",
				time.Now().UTC().Add(time.Second).Truncate(time.Microsecond),
			)

			msg3 := newTestMessage(
				t,
				"23300000-0000-0000-0000-000000000005",
				channelID,
				user1ID,
				false,
				"Third",
				time.Now().UTC().Add(2*time.Second).Truncate(time.Microsecond),
			)

			msg4 := newTestMessage(
				t,
				"23300000-0000-0000-0000-000000000006",
				channelID,
				user1ID,
				false,
				"Fourth",
				time.Now().UTC().Add(3*time.Second).Truncate(time.Microsecond),
			)

			for _, msg := range []*message.Message{
				msg1,
				msg2,
				msg3,
				msg4,
			} {
				if err := repo.Add(ctx, msg); err != nil {
					t.Fatalf("failed to add message: %v", err)
				}
			}

			msg4ID := msg4.ID()

			messages, err := repo.ListByChannel(
				ctx,
				channelID,
				&msg4ID,
				2,
			)
			if err != nil {
				t.Fatalf("failed to list messages: %v", err)
			}

			if len(messages) != 2 {
				t.Fatalf(
					"expected 2 messages, got %d",
					len(messages),
				)
			}

			if messages[0].ID() != msg3.ID() {
				t.Fatalf(
					"expected first message %s, got %s",
					msg3.ID(),
					messages[0].ID(),
				)
			}

			if messages[1].ID() != msg2.ID() {
				t.Fatalf(
					"expected second message %s, got %s",
					msg2.ID(),
					messages[1].ID(),
				)
			}
		})

		t.Run("returns empty result for channel without messages", func(t *testing.T) {
			projectID := newTestMessageProject(
				t,
				db,
				"23400000-0000-0000-0000-000000000001",
			)

			channelID := newTestMessageChannel(
				t,
				db,
				"23400000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			messages, err := repo.ListByChannel(
				ctx,
				channelID,
				nil,
				10,
			)
			if err != nil {
				t.Fatalf("failed to list messages: %v", err)
			}

			if len(messages) != 0 {
				t.Fatalf(
					"expected 0 messages, got %d",
					len(messages),
				)
			}
		})
	})
}

func newTestMessage(
	t *testing.T,
	id string,
	channelID message.ChannelID,
	senderID iam.UserID,
	isSystem bool,
	content string,
	createdAt time.Time,
) *message.Message {
	t.Helper()

	messageID, err := message.NewMessageID(id)
	if err != nil {
		t.Fatal(err)
	}

	messageContent, err := message.NewContent(content)
	if err != nil {
		t.Fatal(err)
	}

	return message.RestoreMessage(
		messageID,
		channelID,
		senderID,
		isSystem,
		messageContent,
		createdAt,
	)
}

func newTestMessageUser(
	t *testing.T,
	db *sql.DB,
	id string,
) iam.UserID {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO users(
			id,
			first_name,
			last_name,
			email,
			role
		) VALUES ($1, $2, $3, $4, $5)`,
		id,
		"Test",
		"User",
		id+"@example.com",
		"client",
	)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

func newTestMessageProject(
	t *testing.T,
	db *sql.DB,
	id string,
) project.ProjectID {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO projects(
			id,
			name,
			status,
			created_at,
			updated_at,
			version
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		id,
		"Test Project "+id,
		"active",
		now,
		now,
		1,
	)
	if err != nil {
		t.Fatalf("failed to insert test project: %v", err)
	}

	projectID, err := project.NewProjectID(id)
	if err != nil {
		t.Fatalf("failed to create project ID: %v", err)
	}

	return projectID
}

func newTestMessageChannel(
	t *testing.T,
	db *sql.DB,
	id string,
	projectID project.ProjectID,
	user1ID iam.UserID,
	user2ID iam.UserID,
) message.ChannelID {
	t.Helper()

	channelID, err := message.NewChannelID(id)
	if err != nil {
		t.Fatalf("failed to create channel ID: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	_, err = db.ExecContext(
		context.Background(),
		`INSERT INTO channels(
			id,
			project_id,
			name,
			version,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		id,
		projectID,
		"Test Channel",
		1,
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to insert test channel: %v", err)
	}

	_, err = db.ExecContext(
		context.Background(),
		`INSERT INTO channel_members(
			channel_id,
			user_id
		) VALUES ($1, $2), ($1, $3)`,
		id,
		user1ID,
		user2ID,
	)
	if err != nil {
		t.Fatalf("failed to insert test channel members: %v", err)
	}

	return channelID
}

func assertMessageEqual(
	t *testing.T,
	expected *message.Message,
	actual *message.Message,
) {
	t.Helper()

	if actual == nil {
		t.Fatal("expected message, got nil")
	}

	if actual.ID() != expected.ID() {
		t.Fatalf(
			"expected ID %s, got %s",
			expected.ID(),
			actual.ID(),
		)
	}

	if actual.ChannelID() != expected.ChannelID() {
		t.Fatalf(
			"expected channel ID %s, got %s",
			expected.ChannelID(),
			actual.ChannelID(),
		)
	}

	if actual.SenderID() != expected.SenderID() {
		t.Fatalf(
			"expected sender ID %s, got %s",
			expected.SenderID(),
			actual.SenderID(),
		)
	}

	if actual.IsSystemMessage() != expected.IsSystemMessage() {
		t.Fatalf(
			"expected system message %t, got %t",
			expected.IsSystemMessage(),
			actual.IsSystemMessage(),
		)
	}

	if actual.Content() != expected.Content() {
		t.Fatalf(
			"expected content %s, got %s",
			expected.Content(),
			actual.Content(),
		)
	}

	if !actual.CreatedAt().Equal(expected.CreatedAt()) {
		t.Fatalf(
			"expected created_at %v, got %v",
			expected.CreatedAt(),
			actual.CreatedAt(),
		)
	}
}
