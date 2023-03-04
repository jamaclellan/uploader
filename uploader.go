package uploader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Uploader struct {
	http.Handler
	baseURL string
	Meta    MetaStore
	Store   FileStore
}

const fileFieldName = "file"

func (u *Uploader) uploadHandler(w http.ResponseWriter, r *http.Request) {
	user := AuthUser(r.Context())
	response := &UploadResponse{}
	file, fileHeader, err := r.FormFile(fileFieldName)
	if err != nil {
		errorResponse(w, response, http.StatusBadRequest, -1001, "file not found in request")
		return
	}
	uploadDetails, err := u.HandleUpload(file, fileHeader.Filename, fileHeader.Size, user.Name)
	if err != nil {
		errorResponseFromError(w, response, err)
	}
	uploadDetails.BuildUrl(u.baseURL)
	response.FromDetails(uploadDetails)
	jsonResponse(w, response, http.StatusAccepted)
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

func (u *Uploader) HandleUpload(file io.ReadSeekCloser, fileName string, fileSize int64, user string) (*UploadDetails, error) {
	fileKey, err := u.Meta.FileKey()
	if err != nil {
		return nil, err
	}
	deleteKey, err := u.Meta.DeleteKey()
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
	if err := u.Meta.FilePut(details); err != nil {
		return nil, err
	}
	if err := u.Store.Put(fileKey, file); err != nil {
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

func (u *Uploader) fileGet(w http.ResponseWriter, r *http.Request) {
	response := &BaseResponse{}
	key := chi.URLParam(r, "key")
	name := chi.URLParam(r, "name")

	meta, err := u.Meta.FileGet(key)
	if errors.Is(err, notFoundError) {
		errorResponse(w, response, 404, -1004, "file not found")
		return
	} else if err != nil {
		errorResponseFromError(w, response, err)
		return
	}

	if name != "" && meta.Filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", meta.Filename))
	}
	w.Header().Set("Content-Type", meta.ContentType)

	file, err := u.Store.Get(key)
	if errors.Is(err, notFoundError) {
		errorResponse(w, response, 404, -1004, "file not found")
		return
	} else if err != nil {
		errorResponseFromError(w, response, err)
		return
	}
	defer file.Close()

	// Send file contents
	if _, err := io.Copy(w, file); err != nil {
		// Will this work? Haven't we already written too much?
		errorResponse(w, response, 500, -5000, "unknown error")
		return
	}
}

func NewUploader(base string, meta MetaStore, store FileStore) *Uploader {
	u := &Uploader{baseURL: base, Meta: meta, Store: store}
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/files/{key}", u.fileGet)
	// This intentionally has a slash after it, as blobs can be fetched with a fake filename to force an attachment.
	router.Get("/files/{key}/{name}", u.fileGet)
	router.With(BearerAuth(u.Meta)).Post("/files", u.uploadHandler)
	// Authenticated get uploads for user
	//router.Get("/uploads/{user}", u.listHandler)
	//router.Delete("/uploads/{user}/{key}", u.deleteHandler)
	// The below route is required for ShareX, as it does not make explicit DELETE requests.
	//router.Get("/uploads/{user}/{key}/{secret}", u.deletePublicHandler)

	u.Handler = router

	return u
}

func NewUploaderFromConfig(cfg *Config) (*Uploader, error) {
	var store FileStore
	var meta MetaStore
	var err error
	if cfg.BoltConfig != nil {
		bc := cfg.BoltConfig
		meta, err = NewBoltStore(bc.Path)
		if err != nil {
			return nil, err
		}
	}
	if cfg.DirConfig != nil {
		dc := cfg.DirConfig
		store = NewDirectoryFileStore(dc.Path)
	}
	if store == nil || meta == nil {
		return nil, errors.New("must have a file store and meta storage configured")
	}
	return NewUploader(cfg.BaseURL, meta, store), nil
}

func (u *Uploader) Close() {
	u.Meta.Close()
	u.Store.Close()
}
