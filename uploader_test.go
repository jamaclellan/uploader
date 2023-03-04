package uploader

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUploaderUploadFile(t *testing.T) {
	uploader := NewUploader()

	contentType, body, err := uploadFile("./test/data/test_file.txt", "file")
	if err != nil {
		t.Fatalf("failed to create upload body: %s", err)
	}

	request := httptest.NewRequest("POST", "/uploads/test_user", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	uploader.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Errorf("wrong status code, got %d want %d", response.Code, http.StatusAccepted)
	}
	assertJSONResponse(t, response)
}

func assertJSONResponse(t testing.TB, response *httptest.ResponseRecorder) {
	t.Helper()
	const want = "application/json"
	contentType := response.Header().Get("Content-Type")
	if contentType != want {
		t.Errorf("bad response content type, got %s wanted %s", contentType, want)
	}
}

func uploadFile(path, field string) (string, io.Reader, error) {
	buffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(buffer)

	fileWriter, err := bodyWriter.CreateFormFile(field, path)
	if err != nil {
		return "", nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open upload file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(fileWriter, file); err != nil {
		return "", nil, fmt.Errorf("failed to copy file to buffer: %w", err)
	}

	bodyWriter.Close()

	return bodyWriter.FormDataContentType(), buffer, nil
}
