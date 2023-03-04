package uploader

import (
	"context"
	"net/http"
	"strings"
)

type authCtxKey int

const AuthContextKey authCtxKey = 0
const CodeAuthFailed = -2000
const AuthHeader = "Authorization"

func BearerAuth(m MetaStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		resp := &BaseResponse{Results: nil}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authFail := func() {
				errorResponse(w, resp, http.StatusUnauthorized, CodeAuthFailed, "Access token is missing or invalid")
			}
			token := r.Header.Get(AuthHeader)
			if token == "" {
				authFail()
				return
			}
			splitToken := strings.SplitN(token, " ", 2)
			if len(splitToken) != 2 {
				authFail()
				return
			}
			user, err := m.UserGetAuth(splitToken[1])
			if err != nil {
				authFail()
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), AuthContextKey, user))
			next.ServeHTTP(w, r)
		})
	}
}

func AuthUser(ctx context.Context) *User {
	if ctx == nil {
		return nil
	}
	if user, ok := ctx.Value(AuthContextKey).(*User); ok {
		return user
	}
	return nil
}
