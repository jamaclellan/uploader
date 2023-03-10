package uploader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"uploader/internal/auth"
	"uploader/internal/responses"

	"github.com/google/go-cmp/cmp"
)

var baseURL = &url.URL{
	Scheme: "http",
	Host:   "localhost",
	Path:   "/",
}

func TestUploaderIntegrationUpload(t *testing.T) {
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "db")
	if err != nil {
		t.Fatal("failed to create temp db file")
	}
	tempFile.Close()
	os.Remove(tempFile.Name())
	meta, err := NewBoltStore(tempFile.Name())
	if err != nil {
		t.Fatalf("recieved error creating meta store %s", err)
	}
	defer meta.Close()
	store := NewDirectoryFileStore(tempDir)

	uploader := NewUploaderHTTP(baseURL, meta, store)
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

	request := httptest.NewRequest("POST", "/uploads/test_user", body)
	request.Header.Set("Content-Type", contentType)
	request.Header.Set(auth.HTTPHeaderName, fmt.Sprintf("Bearer %s", token))
	return request
}

func TestUploaderUploadFileHTTP(t *testing.T) {
	meta := newTestMeta()
	valid, _ := meta.UserRegister("test_user")
	store := newMemoryFileStore()
	uploader := NewUploaderHTTP(baseURL, meta, store)

	request := uploadRequest(t, valid.AuthToken)
	response := httptest.NewRecorder()

	uploader.ServeHTTP(response, request)

	// HTTP Stuff
	assertStatusCode(t, response, http.StatusAccepted)
	assertJSONResponse(t, response)
	decoded, err := decodeUploadResponse(response)
	if err != nil {
		t.Fatalf("failed to decode response %s", err)
	}
	assertJSONNoErrors(t, decoded.ResponseHeader)

	// Actual behavior
	result := decoded.Results
	want := "http://localhost/files/1"
	if result.URL != want {
		t.Errorf("incorrect file url returned, got %s want %s", result.URL, want)
	}
	want = "http://localhost/uploads/test_user/1/delete/delete"
	if result.DeleteURL != want {
		t.Errorf("incorrect delete url returned, got %s want %s", result.DeleteURL, want)
	}
}

func TestFileGetHTTP(t *testing.T) {
	fileKey := "1"
	contents := "Hello, World!"

	meta := newTestMeta()
	meta.addFile(fileKey, "text/plain")
	store := newMemoryFileStore()
	store.Put(fileKey, strings.NewReader(contents))

	uploader := NewUploaderHTTP(baseURL, meta, store)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/files/%s", fileKey), nil)
	response := httptest.NewRecorder()

	uploader.ServeHTTP(response, request)

	assertStatusCode(t, response, http.StatusOK)
	if diff := cmp.Diff(response.Body.Bytes(), []byte("Hello, World!")); diff != "" {
		t.Errorf("response body did not return expected result. %s", diff)
	}
}

func TestFileDeleteHTTP(t *testing.T) {
	fileKey := "1"
	contents := "Hello, World!"

	t.Run("delete method", func(t *testing.T) {
		meta := newTestMeta()
		user, err := meta.UserRegister("test_user")
		if err != nil {
			t.Fatalf("unexpected error registering test user %s", err)
		}
		meta.addFile(fileKey, "text/plain")
		store := newMemoryFileStore()
		store.Put(fileKey, strings.NewReader(contents))

		uploader := NewUploaderHTTP(baseURL, meta, store)

		request := httptest.NewRequest(http.MethodDelete, "http://localhost/uploads/test_user/1", nil)
		request.Header.Set(auth.HTTPHeaderName, fmt.Sprintf("Bearer %s", user.AuthToken))
		response := httptest.NewRecorder()

		uploader.ServeHTTP(response, request)

		assertStatusCode(t, response, 200)
		assertJSONResponse(t, response)

		if _, err = store.Get(fileKey); !errors.Is(err, os.ErrNotExist) {
			t.Error("expected file to be removed from file store")
		}
		if _, err = meta.FileGet(fileKey); !errors.Is(err, ErrNotFound) {
			t.Error("expected file to be removed from meta store")
		}
	})

	t.Run("public delete", func(t *testing.T) {
		meta := newTestMeta()
		meta.addFile(fileKey, "text/plain")
		store := newMemoryFileStore()
		store.Put(fileKey, strings.NewReader(contents))

		uploader := NewUploaderHTTP(baseURL, meta, store)

		// Test files have a fixed delete key of "delete"
		request := httptest.NewRequest(http.MethodGet, "http://localhost/uploads/test_user/1/delete/delete", nil)
		response := httptest.NewRecorder()

		uploader.ServeHTTP(response, request)

		assertStatusCode(t, response, 200)
		assertJSONResponse(t, response)

		if _, err := store.Get(fileKey); !errors.Is(err, os.ErrNotExist) {
			t.Error("expected file to be removed from file store")
		}
		if _, err := meta.FileGet(fileKey); !errors.Is(err, ErrNotFound) {
			t.Error("expected file to be removed from meta store")
		}
	})
}

func decodeUploadResponse(response *httptest.ResponseRecorder) (*UploadResponse, error) {
	uploadResponse := &UploadResponse{}
	if err := json.Unmarshal(response.Body.Bytes(), uploadResponse); err != nil {
		return nil, err
	}
	return uploadResponse, nil
}

func assertJSONNoErrors(t testing.TB, header responses.ResponseHeader) {
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
