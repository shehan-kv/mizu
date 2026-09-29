package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
)

func TestFileRepository(t *testing.T) {
	db := newTestSQLite(t)
	repo := NewFileRepository(db)
	ctx := context.Background()

	user1ID := newTestFileUser(
		t,
		db,
		"10000000-0000-0000-0000-000000000001",
	)

	user2ID := newTestFileUser(
		t,
		db,
		"10000000-0000-0000-0000-000000000002",
	)

	t.Run("Add", func(t *testing.T) {
		t.Run("adds file", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"11000000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"11000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file := newTestFile(
				t,
				"11000000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"document.pdf",
				"document.pdf",
				"files/11000000-document.pdf",
				"application/pdf",
				1024,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, file); err != nil {
				t.Fatalf("failed to add file: %v", err)
			}

			got, err := repo.Get(ctx, file.ID())
			if err != nil {
				t.Fatalf("failed to get file: %v", err)
			}

			assertFileEqual(t, file, got)
		})

		t.Run("returns foreign key error for unknown channel", func(t *testing.T) {
			channelID, err := message.NewChannelID(
				"11000000-0000-0000-0000-000000000010",
			)
			if err != nil {
				t.Fatal(err)
			}

			file := newTestFile(
				t,
				"11000000-0000-0000-0000-000000000011",
				channelID,
				user1ID,
				"unknown.pdf",
				"unknown.pdf",
				"files/11000000-unknown.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			err = repo.Add(ctx, file)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns file", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"12000000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"12000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file := newTestFile(
				t,
				"12000000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"get.pdf",
				"get.pdf",
				"files/12000000-get.pdf",
				"application/pdf",
				200,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, file); err != nil {
				t.Fatalf("failed to add file: %v", err)
			}

			got, err := repo.Get(ctx, file.ID())
			if err != nil {
				t.Fatalf("failed to get file: %v", err)
			}

			assertFileEqual(t, file, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			fileID, err := message.NewFileID(
				"12000000-0000-0000-0000-000000000010",
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.Get(ctx, fileID)
			if !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("expected ErrFileNotFound, got %v", err)
			}
		})
	})

	t.Run("GetStatsByProject", func(t *testing.T) {
		t.Run("returns file count", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"13000000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"13000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			now := time.Now().UTC().Truncate(time.Microsecond)

			file1 := newTestFile(
				t,
				"13000000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"one.pdf",
				"one.pdf",
				"files/13000000-one.pdf",
				"application/pdf",
				100,
				now,
			)

			file2 := newTestFile(
				t,
				"13000000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"two.pdf",
				"two.pdf",
				"files/13000000-two.pdf",
				"application/pdf",
				200,
				now.Add(-time.Hour),
			)

			for _, file := range []*message.File{file1, file2} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			stats, err := repo.GetStatsByProject(ctx, projectID)
			if err != nil {
				t.Fatalf("failed to get stats: %v", err)
			}

			if stats.FileCount() != 2 {
				t.Fatalf("expected file count 2, got %d", stats.FileCount())
			}
		})

		t.Run("ignores files from another project", func(t *testing.T) {
			project1ID := newTestFileProject(
				t,
				db,
				"13100000-0000-0000-0000-000000000001",
			)

			project2ID := newTestFileProject(
				t,
				db,
				"13100000-0000-0000-0000-000000000002",
			)

			channel1ID := newTestFileChannel(
				t,
				db,
				"13100000-0000-0000-0000-000000000003",
				project1ID,
				user1ID,
				user2ID,
			)

			channel2ID := newTestFileChannel(
				t,
				db,
				"13100000-0000-0000-0000-000000000004",
				project2ID,
				user1ID,
				user2ID,
			)

			file1 := newTestFile(
				t,
				"13100000-0000-0000-0000-000000000005",
				channel1ID,
				user1ID,
				"project-one.pdf",
				"project-one.pdf",
				"files/13100000-project-one.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"13100000-0000-0000-0000-000000000006",
				channel2ID,
				user1ID,
				"project-two.pdf",
				"project-two.pdf",
				"files/13100000-project-two.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, file1); err != nil {
				t.Fatalf("failed to add file1: %v", err)
			}

			if err := repo.Add(ctx, file2); err != nil {
				t.Fatalf("failed to add file2: %v", err)
			}

			stats, err := repo.GetStatsByProject(ctx, project1ID)
			if err != nil {
				t.Fatalf("failed to get stats: %v", err)
			}

			if stats.FileCount() != 1 {
				t.Fatalf("expected file count 1, got %d", stats.FileCount())
			}
		})
	})

	t.Run("ListByChannel", func(t *testing.T) {
		t.Run("lists files by channel", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"14000000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"14000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			otherChannelID := newTestFileChannel(
				t,
				db,
				"14000000-0000-0000-0000-000000000003",
				projectID,
				user1ID,
				user2ID,
			)

			file1 := newTestFile(
				t,
				"14000000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"channel-one.pdf",
				"channel-one.pdf",
				"files/14000000-channel-one.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"14000000-0000-0000-0000-000000000005",
				channelID,
				user1ID,
				"channel-two.pdf",
				"channel-two.pdf",
				"files/14000000-channel-two.pdf",
				"application/pdf",
				200,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			otherFile := newTestFile(
				t,
				"14000000-0000-0000-0000-000000000006",
				otherChannelID,
				user1ID,
				"other-channel.pdf",
				"other-channel.pdf",
				"files/14000000-other-channel.pdf",
				"application/pdf",
				300,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			for _, file := range []*message.File{file1, file2, otherFile} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			page, err := common.NewPage(20, 0)
			if err != nil {
				t.Fatal(err)
			}

			files, err := repo.ListByChannel(
				ctx,
				message.ChannelFileFilter{
					ChannelID: channelID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}

			if len(files) != 2 {
				t.Fatalf("expected 2 files, got %d", len(files))
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"14100000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"14100000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file1 := newTestFile(
				t,
				"14100000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"invoice-report.pdf",
				"invoice-report.pdf",
				"files/14100000-invoice-report.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"14100000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"contract.pdf",
				"contract.pdf",
				"files/14100000-contract.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			for _, file := range []*message.File{file1, file2} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			keyword := "invoice"

			page, err := common.NewPage(20, 0)
			if err != nil {
				t.Fatal(err)
			}

			files, err := repo.ListByChannel(
				ctx,
				message.ChannelFileFilter{
					ChannelID: channelID,
					Keyword:   &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}

			if len(files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(files))
			}

			if files[0].ID() != file1.ID() {
				t.Fatalf("expected invoice file, got %s", files[0].ID())
			}
		})

		t.Run("applies pagination", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"14200000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"14200000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			now := time.Now().UTC()

			file1 := newTestFile(
				t,
				"14200000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"one.pdf",
				"one.pdf",
				"files/14200000-one.pdf",
				"application/pdf",
				100,
				now.Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"14200000-0000-0000-0000-000000000005",
				channelID,
				user1ID,
				"two.pdf",
				"two.pdf",
				"files/14200000-two.pdf",
				"application/pdf",
				100,
				now.Add(-time.Hour).Truncate(time.Microsecond),
			)

			file3 := newTestFile(
				t,
				"14200000-0000-0000-0000-000000000006",
				channelID,
				user1ID,
				"three.pdf",
				"three.pdf",
				"files/14200000-three.pdf",
				"application/pdf",
				100,
				now.Add(-2*time.Hour).Truncate(time.Microsecond),
			)

			for _, file := range []*message.File{file1, file2, file3} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			page, err := common.NewPage(1, 1)
			if err != nil {
				t.Fatal(err)
			}

			files, err := repo.ListByChannel(
				ctx,
				message.ChannelFileFilter{
					ChannelID: channelID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}

			if len(files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(files))
			}

			if files[0].ID() != file2.ID() {
				t.Fatalf("expected second file, got %s", files[0].ID())
			}
		})
	})

	t.Run("ListByProject", func(t *testing.T) {
		t.Run("lists files visible to project member", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"15000000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"15000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file := newTestFile(
				t,
				"15000000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"visible.pdf",
				"visible.pdf",
				"files/15000000-visible.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, file); err != nil {
				t.Fatalf("failed to add file: %v", err)
			}

			page, err := common.NewPage(20, 0)
			if err != nil {
				t.Fatal(err)
			}

			files, err := repo.ListByProject(
				ctx,
				message.ProjectFileFilter{
					ProjectID: projectID,
					MemberID:  user1ID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}

			if len(files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(files))
			}

			assertFileEqual(t, file, files[0])
		})

		t.Run("does not return files from channels where member is not a member", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"15100000-0000-0000-0000-000000000001",
			)

			channel1ID := newTestFileChannel(
				t,
				db,
				"15100000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			channel2ID := newTestFileChannel(
				t,
				db,
				"15100000-0000-0000-0000-000000000003",
				projectID,
				user1ID,
				user2ID,
			)

			// Remove user1 from channel2 so that the channel is not visible
			// to user1.
			_, err := db.ExecContext(
				ctx,
				`DELETE FROM channel_members
				 WHERE channel_id = $1 AND user_id = $2`,
				channel2ID,
				user1ID,
			)
			if err != nil {
				t.Fatalf("failed to remove channel member: %v", err)
			}

			file1 := newTestFile(
				t,
				"15100000-0000-0000-0000-000000000004",
				channel1ID,
				user1ID,
				"visible.pdf",
				"visible.pdf",
				"files/15100000-visible.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"15100000-0000-0000-0000-000000000005",
				channel2ID,
				user2ID,
				"hidden.pdf",
				"hidden.pdf",
				"files/15100000-hidden.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, file1); err != nil {
				t.Fatalf("failed to add file1: %v", err)
			}

			if err := repo.Add(ctx, file2); err != nil {
				t.Fatalf("failed to add file2: %v", err)
			}

			page, err := common.NewPage(20, 0)
			if err != nil {
				t.Fatal(err)
			}

			files, err := repo.ListByProject(
				ctx,
				message.ProjectFileFilter{
					ProjectID: projectID,
					MemberID:  user1ID,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}

			if len(files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(files))
			}

			if files[0].ID() != file1.ID() {
				t.Fatalf("expected visible file, got %s", files[0].ID())
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"15200000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"15200000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file1 := newTestFile(
				t,
				"15200000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"invoice.pdf",
				"invoice.pdf",
				"files/15200000-invoice.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"15200000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"contract.pdf",
				"contract.pdf",
				"files/15200000-contract.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			for _, file := range []*message.File{file1, file2} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			keyword := "invoice"

			page, err := common.NewPage(20, 0)
			if err != nil {
				t.Fatal(err)
			}

			files, err := repo.ListByProject(
				ctx,
				message.ProjectFileFilter{
					ProjectID: projectID,
					MemberID:  user1ID,
					Keyword:   &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list files: %v", err)
			}

			if len(files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(files))
			}

			if files[0].ID() != file1.ID() {
				t.Fatalf("expected invoice file, got %s", files[0].ID())
			}
		})
	})

	t.Run("CountByChannel", func(t *testing.T) {
		t.Run("returns count", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"16000000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"16000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file1 := newTestFile(
				t,
				"16000000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"one.pdf",
				"one.pdf",
				"files/16000000-one.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"16000000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"two.pdf",
				"two.pdf",
				"files/16000000-two.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			for _, file := range []*message.File{file1, file2} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			count, err := repo.CountByChannel(
				ctx,
				message.ChannelFileFilter{
					ChannelID: channelID,
				},
			)
			if err != nil {
				t.Fatalf("failed to count files: %v", err)
			}

			if count != 2 {
				t.Fatalf("expected count 2, got %d", count)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"16100000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"16100000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file1 := newTestFile(
				t,
				"16100000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"invoice.pdf",
				"invoice.pdf",
				"files/16100000-invoice.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"16100000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"contract.pdf",
				"contract.pdf",
				"files/16100000-contract.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			for _, file := range []*message.File{file1, file2} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			keyword := "invoice"

			count, err := repo.CountByChannel(
				ctx,
				message.ChannelFileFilter{
					ChannelID: channelID,
					Keyword:   &keyword,
				},
			)
			if err != nil {
				t.Fatalf("failed to count files: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected count 1, got %d", count)
			}
		})
	})

	t.Run("CountByProject", func(t *testing.T) {
		t.Run("returns count for project", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"17000000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"17000000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file1 := newTestFile(
				t,
				"17000000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"one.pdf",
				"one.pdf",
				"files/17000000-one.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"17000000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"two.pdf",
				"two.pdf",
				"files/17000000-two.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			for _, file := range []*message.File{file1, file2} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			count, err := repo.CountByProject(
				ctx,
				message.ProjectFileFilter{
					ProjectID: projectID,
					MemberID:  user1ID,
				},
			)
			if err != nil {
				t.Fatalf("failed to count files: %v", err)
			}

			if count != 2 {
				t.Fatalf("expected count 2, got %d", count)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			projectID := newTestFileProject(
				t,
				db,
				"17100000-0000-0000-0000-000000000001",
			)

			channelID := newTestFileChannel(
				t,
				db,
				"17100000-0000-0000-0000-000000000002",
				projectID,
				user1ID,
				user2ID,
			)

			file1 := newTestFile(
				t,
				"17100000-0000-0000-0000-000000000003",
				channelID,
				user1ID,
				"invoice.pdf",
				"invoice.pdf",
				"files/17100000-invoice.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			file2 := newTestFile(
				t,
				"17100000-0000-0000-0000-000000000004",
				channelID,
				user1ID,
				"contract.pdf",
				"contract.pdf",
				"files/17100000-contract.pdf",
				"application/pdf",
				100,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			for _, file := range []*message.File{file1, file2} {
				if err := repo.Add(ctx, file); err != nil {
					t.Fatalf("failed to add file: %v", err)
				}
			}

			keyword := "invoice"

			count, err := repo.CountByProject(
				ctx,
				message.ProjectFileFilter{
					ProjectID: projectID,
					MemberID:  user1ID,
					Keyword:   &keyword,
				},
			)
			if err != nil {
				t.Fatalf("failed to count files: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected count 1, got %d", count)
			}
		})
	})
}

func newTestFile(
	t *testing.T,
	id string,
	channelID message.ChannelID,
	userID iam.UserID,
	originalName string,
	savedName string,
	storageKey string,
	mimeType string,
	size int64,
	uploadedAt time.Time,
) *message.File {
	t.Helper()

	fileID, err := message.NewFileID(id)
	if err != nil {
		t.Fatal(err)
	}

	originalFileName, err := message.NewFileName(originalName)
	if err != nil {
		t.Fatal(err)
	}

	savedFileName, err := message.NewFileName(savedName)
	if err != nil {
		t.Fatal(err)
	}

	return message.RestoreFile(
		fileID,
		channelID,
		userID,
		originalFileName,
		savedFileName,
		storageKey,
		mimeType,
		size,
		uploadedAt,
	)
}

func newTestFileUser(
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

func newTestFileProject(
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

func newTestFileChannel(
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

func newTestFileChannelMember(
	t *testing.T,
	db *sql.DB,
	channelID message.ChannelID,
	userID iam.UserID,
) {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO channel_members(
			channel_id,
			user_id
		) VALUES ($1, $2)`,
		channelID,
		userID,
	)
	if err != nil {
		t.Fatalf("failed to insert test channel member: %v", err)
	}
}

func assertFileEqual(
	t *testing.T,
	expected *message.File,
	actual *message.File,
) {
	t.Helper()

	if actual == nil {
		t.Fatal("expected file, got nil")
	}

	if actual.ID() != expected.ID() {
		t.Fatalf("expected ID %s, got %s", expected.ID(), actual.ID())
	}

	if actual.ChannelID() != expected.ChannelID() {
		t.Fatalf(
			"expected channel ID %s, got %s",
			expected.ChannelID(),
			actual.ChannelID(),
		)
	}

	if actual.UserID() != expected.UserID() {
		t.Fatalf(
			"expected user ID %s, got %s",
			expected.UserID(),
			actual.UserID(),
		)
	}

	if actual.OriginalName() != expected.OriginalName() {
		t.Fatalf(
			"expected original name %s, got %s",
			expected.OriginalName(),
			actual.OriginalName(),
		)
	}

	if actual.SavedName() != expected.SavedName() {
		t.Fatalf(
			"expected saved name %s, got %s",
			expected.SavedName(),
			actual.SavedName(),
		)
	}

	if actual.StorageKey() != expected.StorageKey() {
		t.Fatalf(
			"expected storage key %s, got %s",
			expected.StorageKey(),
			actual.StorageKey(),
		)
	}

	if actual.MimeType() != expected.MimeType() {
		t.Fatalf(
			"expected MIME type %s, got %s",
			expected.MimeType(),
			actual.MimeType(),
		)
	}

	if actual.Size() != expected.Size() {
		t.Fatalf(
			"expected size %d, got %d",
			expected.Size(),
			actual.Size(),
		)
	}

	if !actual.UploadedAt().Equal(expected.UploadedAt()) {
		t.Fatalf(
			"expected uploaded_at %v, got %v",
			expected.UploadedAt(),
			actual.UploadedAt(),
		)
	}
}
