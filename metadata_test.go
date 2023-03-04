package uploader

import (
	"os"
	"testing"
)

func TestBoltStore_UserRegistration(t *testing.T) {
	meta := newTestBolt(t)
	defer meta.Close()
	user, err := meta.UserRegister("test_user")
	if err != nil {
		t.Fatalf("did not expect error on user registration, got %s", err)
	}
	if user.AuthToken == "" {
		t.Error("expected auth token to be filled in")
	}
	throughUser, err := meta.UserByAuthToken(user.AuthToken)
	if err != nil {
		t.Fatalf("failed to fetch registered user by auth token: %s", err)
	}
	if throughUser.AuthToken != user.AuthToken {
		t.Errorf("users auth tokens don't match")
	}
}

func TestBoltStore_FileDelete(t *testing.T) {
	meta := newTestBolt(t)
	defer meta.Close()
	err := meta.FilePut(UploadDetails{
		Key:         "abc",
		DeleteKey:   "delete",
		Filename:    "test_filename.txt",
		Size:        123,
		ContentType: "test/plain",
		User:        "test_user",
	})
	if err != nil {
		t.Fatalf("failed creating store file: %s", err)
	}
	err = meta.FileDelete("abc")
	if err != nil {
		t.Errorf("unexpected error removing file meta: %s", err)
	}
}

func TestBoltStore_FileKey(t *testing.T) {
	meta := newTestBolt(t)
	defer meta.Close()
	key, err := meta.FileKey()
	if err != nil {
		t.Fatalf("unexpected error on generating file key %s", err)
	}
	if key == "" {
		t.Errorf("expected file key to not be blank")
	}
	// TODO: Ensure file key is unique somehow by forcing a duplicate.
}

func TestBoltStore_DeleteKey(t *testing.T) {
	meta := newTestBolt(t)
	defer meta.Close()
	key, err := meta.DeleteKey()
	if err != nil {
		t.Fatalf("unexpected error on generating delete key %s", err)
	}
	if key == "" {
		t.Errorf("expected delete key to not be blank")
	}
}

func TestBoltStore_FilePut(t *testing.T) {
	meta := newTestBolt(t)
	defer meta.Close()
	t.Run("normal case", func(t *testing.T) {
		err := meta.FilePut(UploadDetails{
			Key: "abc",
		})
		if err != nil {
			t.Errorf("unexpected error on putting file details into store %s", err)
		}
	})
	t.Run("duplicate replaces", func(t *testing.T) {
		err := meta.FilePut(UploadDetails{Key: "123", Size: 15})
		if err != nil {
			t.Fatalf("unexpected error on putting file details into store %s", err)
		}
		err = meta.FilePut(UploadDetails{Key: "123", Size: 31})
		if err != nil {
			t.Fatalf("unexpected error on putting file second time into store %s", err)
		}
		details, err := meta.FileGet("123")
		if err != nil {
			t.Fatalf("unexpected errer on getting file from store %s", err)
		}
		if details.Size != 31 {
			t.Errorf("expected size %d, got %d", 31, details.Size)
		}
	})
}

func newTestBolt(t testing.TB) *BoltStore {
	f, err := os.CreateTemp(t.TempDir(), "testdb-")
	if err != nil {
		t.Fatalf("failed to create test db file %s", err)
	}
	f.Close()
	os.Remove(f.Name())

	b, err := NewBoltStore(f.Name())
	if err != nil {
		t.Fatalf("failed to create test db %s", err)
	}
	return b
}
