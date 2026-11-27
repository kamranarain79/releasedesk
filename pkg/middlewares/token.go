package middlewares

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/services/app"
	"github.com/brewbits-co/releasedesk/internal/services/instance"
)

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

// APITokenAuthorization authenticates a per-app CI token or the generated instance token.
func APITokenAuthorization(apps app.Service, inst instance.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := extractAPIToken(r)
			if got == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			if token, err := apps.LookupToken(got); err == nil && !token.Revoked {
				if token.Scope == "read" && r.Method != http.MethodGet && r.Method != http.MethodHead {
					http.Error(w, "read-only token", http.StatusForbidden)
					return
				}
				scope := token.Scope
				if scope == "" {
					scope = "write"
				}
				ctx := context.WithValue(r.Context(), "tokenAppID", token.AppID)
				ctx = context.WithValue(ctx, "tokenPrefix", token.Prefix)
				ctx = context.WithValue(ctx, "tokenScope", scope)
				ctx = context.WithValue(ctx, "actor", "ci:"+token.Prefix)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if inst.Matches(got) || envTokenMatches(got) {
				ctx := context.WithValue(r.Context(), "tokenGlobal", true)
				ctx = context.WithValue(ctx, "actor", "ci:instance")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		})
	}
}

func envTokenMatches(got string) bool {
	want := os.Getenv("RELEASEDESK_API_TOKEN")
	if want == "" || len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// SafeLogger records the method and path without query strings so tokens are not written to logs.
func SafeLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
