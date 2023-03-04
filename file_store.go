package uploader

import "io"

type FileStore interface {
	Put(key string, file io.Reader) error
}
