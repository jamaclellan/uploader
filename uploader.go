package uploader

import "net/http"

type Uploader struct {
	http.Handler
}

func (u *Uploader) uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

func NewUploader() *Uploader {
	u := &Uploader{}
	router := http.NewServeMux()
	router.Handle("/upload/", http.HandlerFunc(u.uploadHandler))

	u.Handler = router

	return u
}
