package uploader

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"uploader/internal/auth"
)

type testMeta struct {
	as          auth.Store
	putCalls    []string
	getCalls    []string
	deleteCalls int
	keyCalls    int

	files map[string]*UploadDetails
}

func newTestMeta() *testMeta {
	return &testMeta{as: auth.NewMemoryAuthStore(), files: map[string]*UploadDetails{}}
}

func (s *testMeta) addFile(key, contentType string) {
	s.files[key] = &UploadDetails{
		Key:         key,
		DeleteKey:   "delete",
		Filename:    "test_file",
		Size:        18,
		ContentType: contentType,
		User:        "test_user",
	}
}

func (s *testMeta) Close() error {
	return nil
}

func (s *testMeta) UserRegister(name string) (*auth.User, error) {
	return s.as.UserRegister(name)
}

func (s *testMeta) UserByAuthToken(token string) (*auth.User, error) {
	return s.as.UserByAuthToken(token)
}

func (s *testMeta) FileKey() (string, error) {
	s.keyCalls++
	return fmt.Sprintf("%d", s.keyCalls), nil
}

func (s *testMeta) DeleteKey() (string, error) {
	s.deleteCalls++
	return "delete", nil
}

func (s *testMeta) FilePut(details UploadDetails) error {
	s.putCalls = append(s.putCalls, details.Key)
	return nil
}

func (s *testMeta) FileGet(key string) (*UploadDetails, error) {
	s.getCalls = append(s.getCalls, key)
	entry, found := s.files[key]
	if !found {
		return nil, notFoundError
	}
	return entry, nil
}

type memFile struct {
	*bytes.Buffer
}

func (m *memFile) Close() error {
	return nil
}

type memoryFileStore struct {
	files map[string]*memFile
}

func (m *memoryFileStore) Close() error {
	return nil
}

func (m *memoryFileStore) Get(key string) (io.ReadCloser, error) {
	if file, found := m.files[key]; found {
		return file, nil
	}
	return nil, os.ErrNotExist
}

func (m *memoryFileStore) Put(key string, r io.Reader) error {
	f := &memFile{&bytes.Buffer{}}
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	m.files[key] = f
	return nil
}

func newMemoryFileStore() *memoryFileStore {
	return &memoryFileStore{files: map[string]*memFile{}}
}
