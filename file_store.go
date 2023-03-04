package uploader

import (
	"errors"
	"io"
	"os"
	"path"
)

type FileStore interface {
	Close() error
	Put(key string, file io.Reader) error
	Get(key string) (io.ReadCloser, error)
}

type DirectoryFileStore struct {
	prefix string
}

func NewDirectoryFileStore(path string) *DirectoryFileStore {
	return &DirectoryFileStore{
		prefix: path,
	}
}

func (d *DirectoryFileStore) Put(key string, r io.Reader) error {
	file, err := os.Create(path.Join(d.prefix, key))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, r)
	return err
}

func (d *DirectoryFileStore) Get(key string) (io.ReadCloser, error) {
	file, err := os.Open(path.Join(d.prefix, key))
	if errors.Is(err, os.ErrNotExist) {
		return nil, notFoundError
	} else if err != nil {
		return nil, err
	}
	return file, nil
}

func (d *DirectoryFileStore) Close() error {
	return nil
}
