package uploader

import (
	"os"
	"testing"
)

func TestHandleUpload(t *testing.T) {
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

func TestFileGet(t *testing.T) {

}
