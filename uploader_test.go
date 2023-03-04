package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUploaderUploadFileFull(t *testing.T) {
	meta := &SpyMeta{
		users: map[string]string{
			"12345": "test_user",
		},
	}
	store := &SpyStore{}
	uploader := NewUploader("http://localhost/", meta, store)

	contentType, body, err := uploadFile("./test/data/test_file.txt", "file")
	if err != nil {
		t.Fatalf("failed to create upload body: %s", err)
	}

	request := httptest.NewRequest("POST", "/blobs", body)
	request.Header.Set("Content-Type", contentType)
	request.Header.Set(AuthHeader, "Bearer 12345")
	response := httptest.NewRecorder()

	uploader.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Errorf("wrong status code, got %d want %d", response.Code, http.StatusAccepted)
	}
	assertJSONResponse(t, response)
	decoded, err := decodeUploadResponse(response)
	if err != nil {
		t.Fatalf("failed to decode response %s", err)
	}
	assertJSONNoErrors(t, decoded.ResponseHeader)
	result := decoded.Results
	if result.URL == "" {
		t.Errorf("no url returned")
	}
	if result.DeleteURL == "" {
		t.Errorf("no delete url returned what is 1 ")
	}
}

func TestHandleUpload(t *testing.T) {
	meta := &SpyMeta{}
	store := &SpyStore{}
	uploader := NewUploader("http://localhost/", meta, store)

	fileName := "./test/data/test_file.txt"
	stats, err := os.Stat(fileName)
	file, err := os.Open(fileName)

	result, err := uploader.HandleUpload(file, stats.Name(), stats.Size(), "test_user")
	if err != nil {
		t.Fatalf("did not expect error, but received %s", err)
	}
	if meta.keyCalls != 1 {
		t.Errorf("unexpected number of calls for file key, expected %d got %d", 1, meta.keyCalls)
	}
	if meta.deleteCalls != 1 {
		t.Errorf("unexpected number of calls for delete key, expected %d got %d", 1, meta.deleteCalls)
	}
	if len(meta.putCalls) != 1 {
		t.Errorf("unexpected number of calls for putting file metadata, expected %d got %d", 1, len(meta.putCalls))
	}
	if len(store.putCalls) != 1 {
		t.Errorf("unexpected number of calls for putting file in store, expected %d got %d", 1, len(store.putCalls))
	}
	if result.fileKey != "1" {
		t.Errorf("expected file key of 1, got %s", result.fileKey)
	}
}

func decodeUploadResponse(response *httptest.ResponseRecorder) (*UploadResponse, error) {
	uploadResponse := &UploadResponse{}
	if err := json.Unmarshal(response.Body.Bytes(), uploadResponse); err != nil {
		return nil, err
	}
	return uploadResponse, nil
}

func assertJSONNoErrors(t testing.TB, header ResponseHeader) {
	t.Helper()
	if header.Ok != true {
		t.Error("expected okay status in response header")
	}
	if header.Message != "" {
		t.Error("expected error message to be blank in response header")
	}
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
