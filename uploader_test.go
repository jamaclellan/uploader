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

	"github.com/google/go-cmp/cmp"
)

func TestUploaderIntegrationUpload(t *testing.T) {
	tempdir := t.TempDir()
	tempfile, err := os.CreateTemp(tempdir, "db")
	if err != nil {
		t.Fatal("failed to create temp db file")
	}
	tempfile.Close()
	os.Remove(tempfile.Name())
	meta, err := NewBoltStore(tempfile.Name())
	if err != nil {
		t.Fatalf("recieved error creating meta store %s", err)
	}
	defer meta.Close()
	store := NewDirectoryFileStore(tempdir)

	uploader := NewUploader("http://localhost/", meta, store)
	user, err := meta.UserRegister("test_user")
	if err != nil {
		t.Fatalf("failed to register test user")
	}

	request := uploadRequest(t, user.AuthToken)
	response := httptest.NewRecorder()

	uploader.ServeHTTP(response, request)

	assertStatusCode(t, response, http.StatusAccepted)
	assertJSONResponse(t, response)
	decoded, err := decodeUploadResponse(response)
	if err != nil {
		t.Fatalf("failed to decode response %s", err)
	}
	assertJSONNoErrors(t, decoded.ResponseHeader)
}

func assertStatusCode(t testing.TB, response *httptest.ResponseRecorder, status int) {
	t.Helper()
	if response.Code != status {
		t.Errorf("wrong status code, got %d want %d", response.Code, status)
	}
}

func uploadRequest(t testing.TB, token string) *http.Request {
	t.Helper()
	contentType, body, err := uploadFile("./test/data/test_file.txt", "file")
	if err != nil {
		t.Fatalf("failed to create upload body: %s", err)
	}

	request := httptest.NewRequest("POST", "/files", body)
	request.Header.Set("Content-Type", contentType)
	request.Header.Set(AuthHeader, fmt.Sprintf("Bearer %s", token))
	return request
}

func TestUploaderUploadFileFull(t *testing.T) {
	meta := NewSpyMeta()
	valid, _ := meta.UserRegister("test_user")
	store := &SpyStore{}
	uploader := NewUploader("http://localhost/", meta, store)

	request := uploadRequest(t, valid.AuthToken)
	response := httptest.NewRecorder()

	uploader.ServeHTTP(response, request)

	assertStatusCode(t, response, http.StatusAccepted)
	assertJSONResponse(t, response)
	decoded, err := decodeUploadResponse(response)
	if err != nil {
		t.Fatalf("failed to decode response %s", err)
	}
	assertJSONNoErrors(t, decoded.ResponseHeader)
	result := decoded.Results
	want := "http://localhost/files/1"
	if result.URL != want {
		t.Errorf("incorrect file url returned, got %s want %s", result.URL, want)
	}
	want = "http://localhost/uploads/test_user/1/delete"
	if result.DeleteURL != want {
		t.Errorf("incorrect delete url returned, got %s want %s", result.DeleteURL, want)
	}
}

func TestHandleUpload(t *testing.T) {
	meta := NewSpyMeta()
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
	if result.Key != "1" {
		t.Errorf("expected file key of 1, got %s", result.Key)
	}
}

func TestGetFile(t *testing.T) {
	meta := NewSpyMeta()
	meta.addFile("1", "text/plain")
	store := &SpyStore{}
	uploader := NewUploader("http://localhost/", meta, store)

	request := httptest.NewRequest(http.MethodGet, "/files/1", nil)
	response := httptest.NewRecorder()

	uploader.ServeHTTP(response, request)

	assertStatusCode(t, response, http.StatusOK)
	if diff := cmp.Diff(meta.getCalls, []string{"1"}); diff != "" {
		t.Errorf("expected meta get calls to match. %s", diff)
	}
	if diff := cmp.Diff(store.getCalls, []string{"1"}); diff != "" {
		t.Errorf("expected store get calls to match. %s", diff)
	}
	if diff := cmp.Diff(response.Body.Bytes(), []byte("Hello, World!")); diff != "" {
		t.Errorf("response body did not return expected result. %s", diff)
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
