package uploader

import (
	"net/url"

	"uploader/internal/responses"
)

type UploadResponseBody struct {
	URL       string `json:"url"`
	DeleteURL string `json:"delete_url"`
	Size      int    `json:"size"`
	Filename  string `json:"filename"`
}

type UploadResponse struct {
	responses.ResponseHeader
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

func (u *UploadDetails) BuildUrl(base *url.URL) {
	target := base.JoinPath("/files/", u.Key)
	u.url = target.String()
	target = base.JoinPath("/uploads/", u.User, u.Key, "delete", u.DeleteKey)
	u.deleteUrl = target.String()
}
