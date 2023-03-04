package uploader

import (
	"fmt"
	"net/url"

	"uploader/internal/http_responses"
)

type UploadResponseBody struct {
	URL       string `json:"url"`
	DeleteURL string `json:"delete_url"`
	Size      int    `json:"size"`
	Filename  string `json:"filename"`
}

type UploadResponse struct {
	http_responses.ResponseHeader
	Results UploadResponseBody `json:"results"`
}

func (u *UploadResponse) FromDetails(details *UploadDetails) {
	u.Ok = true
	u.Results.URL = details.url
	u.Results.DeleteURL = details.deleteUrl
}

type UploadDetails struct {
	Key         string `json:"key"`
	DeleteKey   string `json:"delete"`
	Filename    string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"type"`
	User        string `json:"user"`

	url       string
	deleteUrl string
}

func (u *UploadDetails) BuildUrl(base string) {
	target, _ := url.JoinPath(base, fmt.Sprintf("/files/%s", u.Key))
	u.url = target
	target, _ = url.JoinPath(base, fmt.Sprintf("/uploads/%s/%s/%s", u.User, u.Key, u.DeleteKey))
	u.deleteUrl = target
}
