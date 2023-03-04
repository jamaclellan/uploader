package uploader

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type authSpy struct {
	user *User
}

func (a *authSpy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if user, ok := r.Context().Value(AuthContextKey).(*User); ok {
		a.user = user
	}
	w.WriteHeader(200)
}

type authTest struct {
	auth     string
	username string
	status   int
}

func TestBearerAuth(t *testing.T) {
	spy := &SpyMeta{
		users: map[string]string{
			"12345": "test_user",
		},
	}
	tests := map[string]authTest{
		"valid case": {
			auth:     "Bearer 12345",
			username: "test_user",
			status:   200,
		},
		"unknown token": {
			auth:     "Bearer 54321",
			username: "",
			status:   401,
		},
		"invalid auth": {
			auth:     "BBBBBBBBBB",
			username: "",
			status:   401,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			authSpy := &authSpy{}
			a := BearerAuth(spy)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set(AuthHeader, test.auth)

			handler := a(authSpy)
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Errorf("incorrect http status, want %d got %d", test.status, recorder.Code)
			}
		})
	}
}
