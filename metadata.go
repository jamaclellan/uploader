package uploader

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"

	"go.etcd.io/bbolt"
)

type FileInfo struct {
	Key      string
	Filename string
	Size     int64
	User     string
}

type MetaStore interface {
	Close()

	// FileKey returns the key of a newly created file instance.
	// A file key is guaranteed to be unique and valid within the store at the time it is returned.
	FileKey() (string, error)
	// DeleteKey returns a key that is used to validate the deletion of a file.
	// It has no guarantee of being unique and is an opaque value without meaning. The implementation makes an attempt to generate
	// a securely random key.
	DeleteKey() (string, error)

	UserRegister(string) (*User, error)
	UserGetAuth(string) (*User, error)

	FilePut(details UploadDetails) error
	FileGet(key string) (*UploadDetails, error)
}

type BoltStore struct {
	db *bbolt.DB
}

var (
	authBucket       = []byte("auth")
	userBucket       = []byte("user")
	userUploadBucket = []byte("user_upload")
	uploadBucket     = []byte("upload")
	bucketList       = [][]byte{authBucket, userBucket, userUploadBucket, uploadBucket}
	duplicateError   = errors.New("duplicate key")
	notFoundError    = errors.New("key not found")
)

func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	if err = db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range bucketList {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &BoltStore{db}, nil
}

func (b *BoltStore) Close() {
	b.db.Close()
}

// FileKey returns a validated unique key that has been reserved within the bolt datastore.
// This is accomplished by inserting a placeholder within the bucket that is expected to be replaced
// with a call to PutFile in a later step.
func (b *BoltStore) FileKey() (string, error) {
	key, err := rand64b()
	if err != nil {
		// Catastrophic key generation error. Somehow.
		return "", err
	}
	for {
		err := b.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(uploadBucket)
			result := b.Get([]byte(key))
			if result != nil {
				return duplicateError
			}
			return b.Put([]byte(key), []byte{})
		})
		// A valid, unused key was found, and a placeholder has been inserted.
		if err == nil {
			return key, nil
		}
		// Key was a duplicate or could not be inserted. Need to try again.
		key, err = rand64b()
	}
}

func (b *BoltStore) DeleteKey() (string, error) {
	return rand64b()
}

func rand64b() (string, error) {
	target := make([]byte, 8)
	for i := 0; i < 32; i++ {
		if _, err := rand.Read(target); err == nil {
			return base64.URLEncoding.EncodeToString(target), nil
		}
	}
	return "", errors.New("failed to generate random key")
}

func (b *BoltStore) UserGetAuth(token string) (*User, error) {
	var user *User
	err := b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(authBucket)
		userBytes := b.Get([]byte(token))
		if userBytes == nil {
			return notFoundError
		}
		user = &User{}
		return json.Unmarshal(userBytes, user)
	})
	return user, err
}

func (b *BoltStore) FilePut(upload UploadDetails) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		uploadValue, err := json.Marshal(upload)
		if err != nil {
			return err
		}
		b := tx.Bucket(uploadBucket)
		return b.Put([]byte(upload.Key), uploadValue)
	})
}

func (b *BoltStore) FileGet(key string) (*UploadDetails, error) {
	return nil, notFoundError
}

func (b *BoltStore) UserRegister(name string) (*User, error) {
	auth, err := rand64b()
	if err != nil {
		return nil, err
	}
	user := &User{AuthToken: auth, Name: name}
	err = b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(authBucket)
		found := b.Get([]byte(auth))
		if found != nil {
			return duplicateError
		}
		value, err := json.Marshal(user)
		if err != nil {
			return err
		}
		return b.Put([]byte(auth), value)
	})
	return user, err
}
