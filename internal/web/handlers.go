package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/salmanmorshed/intstrcodec"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

type Handler struct {
	Conf  *cfg.Config
	Store db.Store
	Codec *intstrcodec.Codec
}

//go:embed index.html
var index_html []byte

func (h *Handler) OpenHomePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.Conf.HomeRedirect == "" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write(index_html)
			return
		}

		http.Redirect(w, r, h.Conf.HomeRedirect, http.StatusFound)
	}
}

func (h *Handler) OpenShortLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodedID := r.PathValue("id")
		decodedID := h.Codec.Decode(encodedID)
		if decodedID <= 0 {
			http.NotFound(w, r)
			return
		}

		link, err := h.Store.RetrieveLinkAndBumpVisits(r.Context(), uint(decodedID))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, link.URL, http.StatusFound)
	}
}

type JD map[string]any

func writeJSON(w http.ResponseWriter, status int, data JD) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to write json: %w", err)
	}
	return nil
}

func (h *Handler) APIVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, JD{"version": cfg.GetVersion()})
	}
}

func defaultQuery(r *http.Request, key, defaultValue string) string {
	if values, ok := r.URL.Query()[key]; ok && len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func GetBaseURL(conf *cfg.Config) string {
	if conf.URLPrefix != "" {
		return conf.URLPrefix
	}
	scheme := "http"
	portSuffix := ""
	if conf.Server.UseTLS {
		scheme = "https"
		if conf.Server.Port != 443 {
			portSuffix = fmt.Sprintf(":%d", conf.Server.Port)
		}
	} else {
		if conf.Server.Port != 80 {
			portSuffix = fmt.Sprintf(":%d", conf.Server.Port)
		}
	}
	return fmt.Sprintf("%s://%s%s", scheme, conf.Server.Host, portSuffix)
}

func getAuthedUser(r *http.Request) *db.User {
	if au, ok := r.Context().Value(ctxUserKey{}).(*db.User); ok {
		return au
	}
	return nil
}

func (h *Handler) LinkList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getAuthedUser(r)
		if user == nil {
			writeJSON(w, http.StatusUnauthorized, JD{"error": "unauthorized"})
			return
		}

		limit, err := strconv.Atoi(defaultQuery(r, "limit", "10"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, JD{"error": "invalid limit value"})
			return
		}
		if limit < 1 || limit > 100 {
			writeJSON(w, http.StatusBadRequest, JD{"error": "limit must be between 1 and 100"})
			return
		}

		offset, err := strconv.Atoi(defaultQuery(r, "offset", "0"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, JD{"error": "invalid offset value"})
			return
		}
		totalLinkCount := h.Store.GetLinkCountForUser(r.Context(), user.Username)
		if offset < 0 || uint(offset) > totalLinkCount {
			writeJSON(w, http.StatusBadRequest, JD{"error": "offset value out of bounds"})
			return
		}

		links, err := h.Store.RetrieveLinksForUser(r.Context(), user.Username, limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, JD{"error": "internal error"})
			return
		}

		results := make([]JD, len(links))
		for i, link := range links {
			encodedID := h.Codec.Encode(int(link.ID))
			results[i] = JD{
				"id":         encodedID,
				"url":        link.URL,
				"short_url":  fmt.Sprintf("%s/%s", GetBaseURL(h.Conf), encodedID),
				"visits":     link.Visits,
				"created_at": link.CreatedAt,
			}
		}

		writeJSON(w, http.StatusOK, JD{
			"results": results,
			"total":   totalLinkCount,
			"limit":   limit,
			"offset":  offset,
			"prefix":  GetBaseURL(h.Conf),
		})
	}
}

func (h *Handler) LinkCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getAuthedUser(r)
		if user == nil {
			writeJSON(w, http.StatusUnauthorized, JD{"error": "unauthorized"})
			return
		}

		var data struct {
			URL string `json:"url"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil || data.URL == "" {
			writeJSON(w, http.StatusBadRequest, JD{"error": "url is required"})
			return
		}

		parsedURL, err := url.Parse(data.URL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			writeJSON(w, http.StatusBadRequest, JD{"error": "url is invalid"})
			return
		}

		link, err := h.Store.CreateLink(r.Context(), data.URL, user.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, JD{"error": "internal error"})
			return
		}

		encodedID := h.Codec.Encode(int(link.ID))
		if encodedID == "api" {
			writeJSON(w, http.StatusInternalServerError, JD{"error": "please try again"})
			return
		}

		writeJSON(w, http.StatusCreated, JD{"short_url": fmt.Sprintf("%s/%s", GetBaseURL(h.Conf), encodedID)})
	}
}

func (h *Handler) LinkDetails() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodedID := r.PathValue("id")
		if encodedID == "" {
			writeJSON(w, http.StatusNotFound, JD{"error": "not found"})
			return
		}

		decodedID := h.Codec.Decode(encodedID)

		link, err := h.Store.RetrieveLink(r.Context(), uint(decodedID))
		if err != nil {
			writeJSON(w, http.StatusNotFound, JD{"error": "not found"})
			return
		}

		user := getAuthedUser(r)
		if user == nil {
			writeJSON(w, http.StatusUnauthorized, JD{"error": "unauthorized"})
			return
		}
		if user.Username != link.CreatedBy {
			writeJSON(w, http.StatusForbidden, JD{"error": "permission denied"})
			return
		}

		writeJSON(w, http.StatusOK, JD{
			"id":         encodedID,
			"url":        link.URL,
			"short_url":  fmt.Sprintf("%s/%s", GetBaseURL(h.Conf), encodedID),
			"visits":     link.Visits,
			"created_at": link.CreatedAt,
		})
	}
}

func (h *Handler) LinkDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodedID := r.PathValue("id")
		if encodedID == "" {
			writeJSON(w, http.StatusNotFound, JD{"error": "not found"})
			return
		}

		decodedID := h.Codec.Decode(encodedID)

		link, err := h.Store.RetrieveLink(r.Context(), uint(decodedID))
		if err != nil {
			writeJSON(w, http.StatusNotFound, JD{"error": "not found"})
			return
		}

		user := getAuthedUser(r)
		if user == nil {
			writeJSON(w, http.StatusUnauthorized, JD{"error": "unauthorized"})
			return
		}
		if user.Username != link.CreatedBy {
			writeJSON(w, http.StatusForbidden, JD{"error": "permission denied"})
			return
		}

		if err := h.Store.DeleteLink(r.Context(), link.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, JD{"error": "internal error"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) UserList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := h.Store.RetrieveAllUsers(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, JD{"error": "internal error"})
			return
		}

		results := make([]JD, len(users))
		for i, user := range users {
			results[i] = JD{
				"username":   user.Username,
				"password":   "<secret>",
				"is_admin":   user.IsAdmin,
				"created_at": user.CreatedAt,
			}
		}

		writeJSON(w, http.StatusOK, JD{"results": results})
	}
}

func (h *Handler) UserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil || data.Username == "" || data.Password == "" {
			writeJSON(w, http.StatusBadRequest, JD{"error": "missing required fields"})
			return
		}

		if validityErr := db.CheckUsernameValidity(data.Username); validityErr != nil {
			writeJSON(w, http.StatusBadRequest, JD{"error": validityErr.Error()})
			return
		}

		if validityErr := db.CheckPasswordStrengthValidity(data.Password); validityErr != nil {
			writeJSON(w, http.StatusBadRequest, JD{"error": validityErr.Error()})
			return
		}

		user, err := h.Store.CreateUser(r.Context(), data.Username, data.Password)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, JD{"error": "internal error"})
			return
		}

		writeJSON(w, http.StatusCreated, JD{
			"username":   user.Username,
			"password":   "<secret>",
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt,
		})
	}
}

func (h *Handler) UserDetailsOrEdit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Password string `json:"password"`
		}

		user, err := h.Store.RetrieveUser(r.Context(), r.PathValue("username"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, JD{"error": "user not found"})
			return
		}

		if r.Method == "GET" {
			goto respondWithUserDetails
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeJSON(w, http.StatusBadRequest, JD{"error": "internal error"})
			return
		}

		if data.Password != "" {
			if validityErr := db.CheckPasswordStrengthValidity(data.Password); validityErr != nil {
				writeJSON(w, http.StatusBadRequest, JD{"error": validityErr.Error()})
				return
			}
			if err := h.Store.UpdatePassword(r.Context(), user.Username, data.Password); err != nil {
				writeJSON(w, http.StatusInternalServerError, JD{"error": "internal error"})
				return
			}
		}

	respondWithUserDetails:
		writeJSON(w, http.StatusOK, JD{
			"username":   user.Username,
			"password":   "<secret>",
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt,
		})
	}
}

func (h *Handler) UserDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := h.Store.RetrieveUser(r.Context(), r.PathValue("username"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, JD{"error": "user not found"})
			return
		}

		if user.IsAdmin {
			writeJSON(w, http.StatusForbidden, JD{"error": "target user is admin"})
			return
		}

		if err := h.Store.DeleteUser(r.Context(), user.Username); err != nil {
			writeJSON(w, http.StatusInternalServerError, JD{"error": "internal error"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
