package web

import (
	"context"
	"net/http"
	"slices"

	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func chainMiddleware(mws ...Middleware) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		for _, mw := range slices.Backward(mws) {
			h = mw(h)
		}
		return h
	}
}

func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			w.Header().Set("Vary", "Origin")
			if origin != "" && slices.Contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Content-Length")
				w.Header().Set("Access-Control-Expose-Headers", "WWW-Authenticate, Content-Type, Content-Length")

				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next(w, r)
		}
	}
}

type ctxUserKey struct{}

func BasicAuthMiddleware(store db.Store) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			username, password, hasAuth := r.BasicAuth()

			if !hasAuth {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				writeJSON(w, http.StatusUnauthorized, JD{"error": "missing authentication credentials"})
				return
			}

			user, err := store.RetrieveUser(r.Context(), username)
			if err != nil || !db.VerifyPassword(user.Password, password) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				writeJSON(w, http.StatusUnauthorized, JD{"error": "incorrect username and/or password"})
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserKey{}, user)
			next(w, r.WithContext(ctx))
		}
	}
}

func AdminFilterMiddleware() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(ctxUserKey{}).(*db.User)

			if !ok || !user.IsAdmin {
				writeJSON(w, http.StatusForbidden, JD{"error": "user is not admin"})
				return
			}

			next(w, r)
		}
	}
}
