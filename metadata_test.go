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

func newTestBolt(t testing.TB) *BoltStore {
	f, err := os.CreateTemp("", "testdb-")
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
