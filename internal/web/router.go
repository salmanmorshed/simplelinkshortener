package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func SetupRouter(rootCtx context.Context, conf *cfg.Config, store db.Store, codec *intstrcodec.Codec) func() error {
	handler := Handler{conf, store, codec}

	apiCommon := chainMiddleware(CORSMiddleware(conf.Server.CORSOrigins))
	authedOnly := chainMiddleware(apiCommon, BasicAuthMiddleware(store))
	adminOnly := chainMiddleware(authedOnly, AdminFilterMiddleware())

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handler.OpenHomePage())
	mux.HandleFunc("GET /{id}", handler.OpenShortLink())

	mux.HandleFunc("GET /api", apiCommon(handler.APIVersion()))
	mux.HandleFunc("GET /api/links", authedOnly(handler.LinkList()))
	mux.HandleFunc("POST /api/links", authedOnly(handler.LinkCreate()))
	mux.HandleFunc("GET /api/links/{id}", authedOnly(handler.LinkDetails()))
	mux.HandleFunc("DELETE /api/links/{id}", authedOnly(handler.LinkDelete()))

	mux.HandleFunc("GET /api/users", adminOnly(handler.UserList()))
	mux.HandleFunc("POST /api/users", adminOnly(handler.UserCreate()))
	mux.HandleFunc("GET /api/users/{username}", adminOnly(handler.UserDetailsOrEdit()))
	mux.HandleFunc("PATCH /api/users/{username}", adminOnly(handler.UserDetailsOrEdit()))
	mux.HandleFunc("DELETE /api/users/{username}", adminOnly(handler.UserDelete()))

	return func() error {
		address := fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port)
		fmt.Println("Serving server at:", address)
		if !conf.Server.UseTLS {
			return http.ListenAndServe(address, mux)
		}
		return http.ListenAndServeTLS(address, conf.Server.TLSCertificate, conf.Server.TLSPrivateKey, mux)
	}
}
