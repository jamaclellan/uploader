package uploader

import (
	"bytes"
	"io"
)

type SpyStore struct {
	putCalls []string
	getCalls []string
}

type bufferCloser struct {
	*bytes.Buffer
}

func (b bufferCloser) Close() error {
	return nil
}

func (s *SpyStore) Put(key string, file io.Reader) error {
	s.putCalls = append(s.putCalls, key)
	return nil
}

func (s *SpyStore) Get(key string) (io.ReadCloser, error) {
	s.getCalls = append(s.getCalls, key)
	b := bufferCloser{bytes.NewBufferString("Hello, World!")}
	return b, nil
}
