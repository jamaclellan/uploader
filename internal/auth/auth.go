package auth

import (
	"context"
	"net/http"
	"strings"

	"uploader/internal/responses"
)

type authCtxKey int
type authError string

func (a authError) Error() string {
	return string(a)
}

const NotFoundError = authError("user not found")
const DuplicateError = authError("duplicate username")
const ContextKey authCtxKey = 0
const CodeAuthFailed = -2000
const HTTPHeaderName = "Authorization"

type Store interface {
	UserByAuthToken(string) (*User, error)
	UserRegister(string) (*User, error)
}

func BearerAuth(m Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		resp := &responses.BaseResponse{Results: nil}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authFail := func() {
				responses.Error(w, resp, http.StatusUnauthorized, CodeAuthFailed, "Access token is missing or invalid")
			}
			token := r.Header.Get(HTTPHeaderName)
			if token == "" {
				authFail()
				return
			}
			splitToken := strings.SplitN(token, " ", 2)
			if len(splitToken) != 2 {
				authFail()
				return
			}
			user, err := m.UserByAuthToken(splitToken[1])
			if err != nil {
				authFail()
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), ContextKey, user))
			next.ServeHTTP(w, r)
		})
	}
}

func AuthUser(ctx context.Context) *User {
	if ctx == nil {
		return nil
	}
	if user, ok := ctx.Value(ContextKey).(*User); ok {
		return user
	}
	return nil
}
