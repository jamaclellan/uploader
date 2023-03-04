package uploader

import "io"

type SpyStore struct {
	putCalls []string
}

func (s *SpyStore) Put(key string, file io.Reader) error {
	s.putCalls = append(s.putCalls, key)
	return nil
}
