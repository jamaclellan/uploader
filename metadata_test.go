package uploader

import (
	"errors"
	"fmt"
)

type SpyMeta struct {
	userReqs    []string
	putCalls    []string
	deleteCalls int
	keyCalls    int

	users map[string]string
}

var authNotFound = errors.New("token not found")

func (s *SpyMeta) UserGetAuth(token string) (*User, error) {
	s.userReqs = append(s.userReqs, token)
	if name, found := s.users[token]; found {
		return &User{
			Name:    name,
			AuthKey: token,
		}, nil
	}
	return nil, authNotFound
}

func (s *SpyMeta) FileKey() string {
	s.keyCalls++
	return fmt.Sprintf("%d", s.keyCalls)
}

func (s *SpyMeta) DeleteKey() string {
	s.deleteCalls++
	return "delete"
}

func (s *SpyMeta) FilePut(key, delete, filename string, size int64, user string) error {
	s.putCalls = append(s.putCalls, key)
	return nil
}
