package uploader

import (
	"errors"
	"fmt"
	"math/rand"
)

type SpyMeta struct {
	userReqs    []string
	putCalls    []string
	getCalls    []string
	deleteCalls int
	keyCalls    int

	users map[string]*User
	files map[string]*UploadDetails
}

func NewSpyMeta() *SpyMeta {
	return &SpyMeta{users: map[string]*User{}, files: map[string]*UploadDetails{}}
}

var authNotFound = errors.New("token not found")

func (s *SpyMeta) addFile(key, contentType string) {
	s.files[key] = &UploadDetails{
		Key:         key,
		DeleteKey:   "delete",
		Filename:    "test_file",
		Size:        18,
		ContentType: contentType,
		User:        "test_user",
	}
}

func (s *SpyMeta) Close() {

}

func (s *SpyMeta) UserGetAuth(token string) (*User, error) {
	s.userReqs = append(s.userReqs, token)
	if user, found := s.users[token]; found {
		return user, nil
	}
	return nil, authNotFound
}

func (s *SpyMeta) UserRegister(name string) (*User, error) {
	token := fmt.Sprintf("%d", rand.Int63())
	user := &User{
		Name:      name,
		AuthToken: token,
	}
	s.users[token] = user
	return user, nil
}

func (s *SpyMeta) FileKey() (string, error) {
	s.keyCalls++
	return fmt.Sprintf("%d", s.keyCalls), nil
}

func (s *SpyMeta) DeleteKey() (string, error) {
	s.deleteCalls++
	return "delete", nil
}

func (s *SpyMeta) FilePut(details UploadDetails) error {
	s.putCalls = append(s.putCalls, details.Key)
	return nil
}

func (s *SpyMeta) FileGet(key string) (*UploadDetails, error) {
	s.getCalls = append(s.getCalls, key)
	entry, found := s.files[key]
	if !found {
		return nil, notFoundError
	}
	return entry, nil
}
