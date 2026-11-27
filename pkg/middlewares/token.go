package middlewares

import (
	"context"
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/services/auth"
)

const defaultAPIToken = "90a514ab93e2c32fdd1072154b26a100"

func expectedAPIToken() string {
	if token := os.Getenv("RELEASEDESK_API_TOKEN"); token != "" {
		return token
	}
	return defaultAPIToken
}

func extractAPIToken(r *http.Request) string {
	if token := r.Header.Get("X-Auth"); token != "" {
		return token
	}
	if authz := r.Header.Get("Authorization"); strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	if _, password, ok := r.BasicAuth(); ok {
		return password
	}
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	return ""
}

// APITokenAuthorization authenticates the Python / Filebrowser-style client.
func APITokenAuthorization(authSrv auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := extractAPIToken(r)
			want := expectedAPIToken()
			if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			sessionCtx := context.WithValue(r.Context(), "userID", 1)
			sessionCtx = context.WithValue(sessionCtx, "username", "admin")
			next.ServeHTTP(w, r.WithContext(sessionCtx))
		})
	}
}
