package uploader

type FileInfo struct {
	Key      string
	Filename string
	Size     int64
	User     string
}

type MetaStore interface {
	FileKey() string
	DeleteKey() string
	UserGetAuth(string) (*User, error)
	FilePut(key, delete, filename string, size int64, user string) error
}
