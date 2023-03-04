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
