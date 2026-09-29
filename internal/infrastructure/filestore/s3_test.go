package filestore

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newTestS3Store(t *testing.T) *S3Store {
	t.Helper()

	ctx := context.Background()

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "chrislusf/seaweedfs:4.47",
				ExposedPorts: []string{"8333/tcp"},
				Cmd: []string{
					"server",
					"-s3",
					"-s3.port=8333",
				},
				WaitingFor: wait.ForListeningPort("8333/tcp"),
			},
			Started: true,
		},
	)
	if err != nil {
		t.Fatalf("failed to start SeaweedFS container: %v", err)
	}

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get SeaweedFS host: %v", err)
	}

	port, err := container.MappedPort(ctx, "8333/tcp")
	if err != nil {
		t.Fatalf("failed to get SeaweedFS port: %v", err)
	}

	endpoint := "http://" + host + ":" + port.Port()

	store, err := NewS3Store(ctx, S3Config{
		Region:    "us-east-1",
		Bucket:    "test-bucket",
		AccessKey: "any",
		SecretKey: "any",
		Endpoint:  endpoint,
	})
	if err != nil {
		t.Fatalf("failed to create S3 store: %v", err)
	}

	_, err = store.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(store.bucket),
	})
	if err != nil {
		t.Fatalf("failed to create test bucket: %v", err)
	}

	return store
}

func TestS3Store(t *testing.T) {
	store := newTestS3Store(t)

	t.Run("saves file contents", func(t *testing.T) {
		key := "save-test.txt"
		content := "hello, world"

		err := store.Save(
			context.Background(),
			key,
			strings.NewReader(content),
		)
		if err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		out, err := store.client.GetObject(
			context.Background(),
			&s3.GetObjectInput{
				Bucket: aws.String(store.bucket),
				Key:    aws.String(key),
			},
		)
		if err != nil {
			t.Fatalf("failed to get saved object: %v", err)
		}
		defer out.Body.Close()

		data, err := io.ReadAll(out.Body)
		if err != nil {
			t.Fatalf("failed to read saved object: %v", err)
		}

		if string(data) != content {
			t.Fatalf("expected %q, got %q", content, string(data))
		}
	})

	t.Run("saves empty file", func(t *testing.T) {
		key := "empty-test.txt"

		err := store.Save(
			context.Background(),
			key,
			strings.NewReader(""),
		)
		if err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		out, err := store.client.GetObject(
			context.Background(),
			&s3.GetObjectInput{
				Bucket: aws.String(store.bucket),
				Key:    aws.String(key),
			},
		)
		if err != nil {
			t.Fatalf("failed to get saved object: %v", err)
		}
		defer out.Body.Close()

		data, err := io.ReadAll(out.Body)
		if err != nil {
			t.Fatalf("failed to read saved object: %v", err)
		}

		if len(data) != 0 {
			t.Fatalf("expected empty object, got %d bytes", len(data))
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		key := "overwrite-test.txt"

		err := store.Save(
			context.Background(),
			key,
			strings.NewReader("old content"),
		)
		if err != nil {
			t.Fatalf("first Save() returned error: %v", err)
		}

		err = store.Save(
			context.Background(),
			key,
			strings.NewReader("new content"),
		)
		if err != nil {
			t.Fatalf("second Save() returned error: %v", err)
		}

		reader, err := store.Open(context.Background(), key)
		if err != nil {
			t.Fatalf("Open() returned error: %v", err)
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read object: %v", err)
		}

		if string(data) != "new content" {
			t.Fatalf("expected %q, got %q", "new content", string(data))
		}
	})

	t.Run("opens saved file", func(t *testing.T) {
		key := "open-test.txt"
		content := "hello from S3"

		err := store.Save(
			context.Background(),
			key,
			strings.NewReader(content),
		)
		if err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		reader, err := store.Open(context.Background(), key)
		if err != nil {
			t.Fatalf("Open() returned error: %v", err)
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read object: %v", err)
		}

		if string(data) != content {
			t.Fatalf("expected %q, got %q", content, string(data))
		}
	})

	t.Run("returns error when file does not exist", func(t *testing.T) {
		_, err := store.Open(
			context.Background(),
			"does-not-exist.txt",
		)
		if err == nil {
			t.Fatal("expected Open() to return an error")
		}

		if !strings.Contains(err.Error(), "s3store.Open: get object") {
			t.Fatalf("expected wrapped Open error, got: %v", err)
		}
	})

	t.Run("deletes existing file", func(t *testing.T) {
		key := "delete-test.txt"

		err := store.Save(
			context.Background(),
			key,
			strings.NewReader("delete me"),
		)
		if err != nil {
			t.Fatalf("Save() returned error: %v", err)
		}

		err = store.Delete(context.Background(), key)
		if err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}

		_, err = store.Open(context.Background(), key)
		if err == nil {
			t.Fatal("expected Open() to fail after Delete()")
		}
	})

	t.Run("succeeds when file does not exist", func(t *testing.T) {
		err := store.Delete(
			context.Background(),
			"already-missing.txt",
		)
		if err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}
	})
}
