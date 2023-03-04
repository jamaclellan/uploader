package uploader

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
)

type UploadService interface {
	Close() error
	Upload(r io.ReadSeekCloser, name string, size int64, user string) (*UploadDetails, error)
	Get(key string) (*UploadDetails, io.ReadCloser, error)
	Delete(key, deleteKey string) error
}

type KeyMeta interface {
	FileKey() (string, error)
	DeleteKey() (string, error)
}

type UploadMeta interface {
	KeyMeta
	Close() error
	FilePut(details UploadDetails) error
	FileGet(key string) (*UploadDetails, error)
	FileDelete(key string) error
}

type uploadService struct {
	meta  UploadMeta
	store FileStore
}

func NewUploadService(meta UploadMeta, store FileStore) *uploadService {
	return &uploadService{
		meta,
		store,
	}
}

func (u *uploadService) Upload(file io.ReadSeekCloser, fileName string, fileSize int64, user string) (*UploadDetails, error) {
	fileKey, err := u.meta.FileKey()
	if err != nil {
		return nil, err
	}
	deleteKey, err := u.meta.DeleteKey()
	if err != nil {
		return nil, err
	}

	details := UploadDetails{
		Key:         fileKey,
		DeleteKey:   deleteKey,
		Filename:    fileName,
		Size:        fileSize,
		ContentType: contentTypeFromFile(file),
		User:        user,
	}
	if err := u.meta.FilePut(details); err != nil {
		return nil, err
	}
	if err := u.store.Put(fileKey, file); err != nil {
		return nil, err
	}
	return &UploadDetails{
		Key:       fileKey,
		DeleteKey: deleteKey,
		Filename:  fileName,
		Size:      fileSize,
		User:      user,
	}, nil
}

func contentTypeFromFile(file io.ReadSeeker) string {
	// Content detection
	start := &bytes.Buffer{}
	io.CopyN(start, file, 512)
	contentType := http.DetectContentType(start.Bytes())

	// Rewind file so that full file is copied later.
	file.Seek(0, 0)

	return contentType
}

func (u *uploadService) Get(key string) (*UploadDetails, io.ReadCloser, error) {
	meta, err := u.meta.FileGet(key)
	if errors.Is(err, notFoundError) {
		return nil, nil, os.ErrNotExist
	}
	file, err := u.store.Get(key)
	if err != nil {
		return nil, nil, err
	}
	return meta, file, nil
}

func (u *uploadService) Delete(key, deleteKey string) error {
	return nil
}

func (u *uploadService) Close() error {
	u.store.Close()
	return u.meta.Close()
}
