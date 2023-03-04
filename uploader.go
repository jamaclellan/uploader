package uploader

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
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

func (u *Uploader) HandleUpload(file io.ReadSeekCloser, fileName string, fileSize int64, user string) (*UploadDetails, error) {
	fileKey := u.Meta.FileKey()
	deleteKey := u.Meta.DeleteKey()
	if err := u.Meta.FilePut(fileKey, deleteKey, fileName, fileSize, user); err != nil {
		return nil, err
	}
	if err := u.Store.Put(fileKey, file); err != nil {
		return nil, err
	}
	return &UploadDetails{
		fileKey:   fileKey,
		deleteKey: deleteKey,
		name:      fileName,
		size:      fileSize,
		user:      user,
	}, nil
}

func NewUploader(base string, meta MetaStore, store FileStore) *Uploader {
	u := &Uploader{baseURL: base, Meta: meta, Store: store}
	router := chi.NewRouter()
	//router.Use(middleware.Logger)
	// router.Get("/blobs/{key}", u.blobGet)
	// This intentionally has a slash after it, as blobs can be fetched with a fake filename to force an attachment.
	//router.Get("/blobs/{key}/", u.blobGet)
	router.With(BearerAuth(u.Meta)).Post("/blobs", u.uploadHandler)
	// Authenticated get uploads for user
	//router.Get("/uploads/{user}", u.listHandler)
	//router.Delete("/uploads/{user}/{key}", u.deleteHandler)
	// The below route is required for ShareX, as it does not make explicit DELETE requests.
	//router.Get("/uploads/{user}/{key}/{secret}", u.deletePublicHandler)

	u.Handler = router

	return u
}
