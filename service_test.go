package uploader

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUploadService_Upload(t *testing.T) {
	meta := newTestMeta()
	store := newMemoryFileStore()
	uploader := NewUploadService(meta, store)

	fileName := "./test/data/test_file.txt"
	stats, err := os.Stat(fileName)
	file, err := os.Open(fileName)

	result, err := uploader.Upload(file, stats.Name(), stats.Size(), "test_user")
	if err != nil {
		t.Fatalf("did not expect error, but received %s", err)
	}
	if len(meta.putCalls) != 1 {
		t.Errorf("unexpected number of calls for putting file metadata, expected %d got %d", 1, len(meta.putCalls))
	}
	if result.Key != "1" {
		t.Errorf("expected file key of 1, got %s", result.Key)
	}
}

func TestUploadService_Get(t *testing.T) {
	meta := newTestMeta()
	store := newMemoryFileStore()
	uploader := NewUploadService(meta, store)

	t.Run("not found", func(t *testing.T) {
		_, _, err := uploader.Get("123")
		if !errors.Is(err, os.ErrNotExist) {
			t.Error("expected to get not found")
		}
	})
	t.Run("found", func(t *testing.T) {
		contents := "Hello, World!"

		meta.addFile("abc123", "text/plain")
		if err := store.Put("abc123", strings.NewReader(contents)); err != nil {
			t.Fatalf("unexpected error seeding store %s", err)
		}

		deets, reader, err := uploader.Get("abc123")
		if err != nil {
			t.Fatalf("did not expect error %s", err)
		}
		readContents := &bytes.Buffer{}
		io.Copy(readContents, reader)
		if diff := cmp.Diff(readContents.Bytes(), []byte(contents)); diff != "" {
			t.Errorf("contents mismatch %s", diff)
		}
		if deets.ContentType != "text/plain" {
			t.Errorf("wanted content type %s, got %s", "text/plain", deets.ContentType)
		}
	})
}
