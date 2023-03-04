package uploader

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Uploader struct {
	http.Handler
}

func (u *Uploader) uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

func NewUploader() *Uploader {
	u := &Uploader{}
	router := chi.NewRouter()
	//router.Use(middleware.Logger)
	router.Post("/uploads/{user}", u.uploadHandler)

	u.Handler = router

	return u
}
