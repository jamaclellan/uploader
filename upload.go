package uploader

import (
	"fmt"
	"net/url"
)

type UploadResponseBody struct {
	URL       string `json:"url"`
	DeleteURL string `json:"delete_url"`
	Size      int    `json:"size"`
	Filename  string `json:"filename"`
}

type UploadResponse struct {
	ResponseHeader
	Results UploadResponseBody `json:"results"`
}

func (u *UploadResponse) FromDetails(details *UploadDetails) {
	u.Ok = true
	u.Results.URL = details.url
	u.Results.DeleteURL = details.deleteUrl
}

type UploadDetails struct {
	fileKey   string
	deleteKey string
	name      string
	size      int64
	user      string

	url       string
	deleteUrl string
}

func (u *UploadDetails) BuildUrl(base string) {
	target, _ := url.JoinPath(base, fmt.Sprintf("/blobs/%s/%s", u.fileKey, u.name))
	u.url = target
	target, _ = url.JoinPath(base, fmt.Sprintf("/uploads/%s/%s/%s", u.user, u.fileKey, u.deleteKey))
	u.deleteUrl = target
}
