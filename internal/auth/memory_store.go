package auth

import "fmt"

type MemoryAuthStore struct {
	tokens map[string]*User
	users  map[string]*User
	key    int
}

func NewMemoryAuthStore() *MemoryAuthStore {
	return &MemoryAuthStore{users: map[string]*User{}, tokens: map[string]*User{}}
}

func (s *MemoryAuthStore) UserByAuthToken(token string) (*User, error) {
	if user, found := s.tokens[token]; found {
		return user, nil
	}
	return nil, NotFoundError
}

func (s *MemoryAuthStore) UserRegister(name string) (*User, error) {
	user := &User{
		Name:      name,
		AuthToken: fmt.Sprintf("%d", s.key),
	}
	s.key++
	s.users[name] = user
	s.tokens[user.AuthToken] = user
	return user, nil
}
