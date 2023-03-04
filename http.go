package uploader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"uploader/internal/auth"
	"uploader/internal/http_responses"

	"github.com/go-chi/chi/v5"
)

type Uploader struct {
	http.Handler
	baseURL string
	us      UploadService
}

const fileFieldName = "file"

func (u *Uploader) uploadHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.AuthUser(r.Context())
	response := &UploadResponse{}
	file, fileHeader, err := r.FormFile(fileFieldName)
	if err != nil {
		http_responses.Error(w, response, http.StatusBadRequest, -1001, "file not found in request")
		return
	}
	uploadDetails, err := u.us.Upload(file, fileHeader.Filename, fileHeader.Size, user.Name)
	if err != nil {
		http_responses.ErrorFromError(w, response, err)
	}
	uploadDetails.BuildUrl(u.baseURL)
	response.FromDetails(uploadDetails)
	http_responses.Json(w, response, http.StatusAccepted)
}

func (u *Uploader) fileGet(w http.ResponseWriter, r *http.Request) {
	response := &http_responses.BaseResponse{}
	key := chi.URLParam(r, "key")
	name := chi.URLParam(r, "name")

	details, reader, err := u.us.Get(key)
	if errors.Is(err, os.ErrNotExist) {
		http_responses.Error(w, response, 404, -1004, "file not found")
		return
	} else if err != nil {
		http_responses.ErrorFromError(w, response, err)
		return
	}
	defer reader.Close()

	if name != "" && details.Filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", details.Filename))
	}
	w.Header().Set("Content-Type", details.ContentType)

	// Send file contents
	if _, err := io.Copy(w, reader); err != nil {
		// Will this work? Haven't we already written too much?
		http_responses.Error(w, response, 500, -5000, "unknown error")
		return
	}
}

func NewUploaderHTTP(base string, meta MetaStore, store FileStore) *Uploader {
	u := &Uploader{baseURL: base, us: NewUploadService(meta, store)}
	router := chi.NewRouter()
	//router.Use(middleware.Logger)
	router.Get("/files/{key}", u.fileGet)
	// This intentionally has a slash after it, as blobs can be fetched with a fake filename to force an attachment.
	router.Get("/files/{key}/{name}", u.fileGet)
	router.With(auth.BearerAuth(meta)).Post("/files", u.uploadHandler)
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
	return NewUploaderHTTP(cfg.BaseURL, meta, store), nil
}

func (u *Uploader) Close() {
	u.us.Close()
}
