package uploader

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"

	"uploader/internal/auth"

	"go.etcd.io/bbolt"
)

type MetaStore interface {
	Close() error

	auth.Store

	// FileKey returns the key of a newly created file instance.
	// A file key is guaranteed to be unique and valid within the store at the time it is returned.
	FileKey() (string, error)
	// DeleteKey returns a key that is used to validate the deletion of a file.
	// It has no guarantee of being unique and is an opaque value without meaning. The implementation makes an attempt to generate
	// a securely random key.
	DeleteKey() (string, error)

	UploadMeta
}

type BoltStore struct {
	db *bbolt.DB
}

const (
	bucketAuth        = "auth"
	bucketUsers       = "user"
	bucketUserUploads = "user_upload"
	bucketUpload      = "upload"
)

var (
	bucketList     = []string{bucketAuth, bucketUsers, bucketUserUploads, bucketUpload}
	duplicateError = errors.New("duplicate key")
	notFoundError  = errors.New("key not found")
)

func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	if err = db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range bucketList {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &BoltStore{db}, nil
}

func (b *BoltStore) Close() error {
	return b.db.Close()
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
			b := tx.Bucket([]byte(bucketUpload))
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

func (b *BoltStore) getJson(bucket, key string, target any) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v := b.Get([]byte(key))
		if v == nil {
			return notFoundError
		}
		return json.Unmarshal(v, target)
	})
}

func (b *BoltStore) putJson(bucket, key string, value any) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		uploadValue, err := json.Marshal(value)
		if err != nil {
			return err
		}
		b := tx.Bucket([]byte(bucket))
		return b.Put([]byte(key), uploadValue)
	})
}

func (b *BoltStore) putJsonNoDupe(bucket, key string, value any) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if v := b.Get([]byte(key)); v != nil {
			return duplicateError
		}
		uploadValue, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), uploadValue)
	})
}

func (b *BoltStore) FilePut(upload UploadDetails) error {
	return b.putJson(bucketUpload, upload.Key, upload)
}

func (b *BoltStore) FileGet(key string) (*UploadDetails, error) {
	upload := &UploadDetails{}
	return upload, b.getJson(bucketUpload, key, upload)
}

func (b *BoltStore) UserByAuthToken(token string) (*auth.User, error) {
	user := &auth.User{}
	err := b.getJson(bucketAuth, token, user)
	return user, err
}

func (b *BoltStore) UserRegister(name string) (*auth.User, error) {
	token, err := rand64b()
	if err != nil {
		return nil, err
	}
	user := &auth.User{AuthToken: token, Name: name}
	return user, b.putJsonNoDupe(bucketAuth, user.AuthToken, user)
}
