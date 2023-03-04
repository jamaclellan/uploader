package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type authSpy struct {
	user *User
}

func (a *authSpy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.user = AuthUser(r.Context())
	w.WriteHeader(200)
}

type authTest struct {
	auth     string
	username string
	status   int
}

func TestBearerAuth(t *testing.T) {
	store := NewMemoryAuthStore()
	valid, _ := store.UserRegister("test_user")
	tests := map[string]authTest{
		"valid case": {
			auth:     fmt.Sprintf("Bearer %s", valid.AuthToken),
			username: "test_user",
			status:   200,
		},
		"unknown token": {
			auth:     "Bearer abc",
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
			a := BearerAuth(store)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set(HTTPHeaderName, test.auth)

			handler := a(authSpy)
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Errorf("incorrect http status, want %d got %d", test.status, recorder.Code)
			}
			if test.username != "" {
				if authSpy.user == nil {
					t.Error("expected user to be found, but was not")
				} else if authSpy.user.Name != test.username {
					t.Errorf("expected user to be %s but found %s", test.username, authSpy.user.Name)
				}
			}
		})
	}
}
