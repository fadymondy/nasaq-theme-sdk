// Package nasaqtest is a fake Nasaq UI Studio for tests: a catalog with one
// theme covering the app ("clinic") and one that does not ("shop"), and one
// project ("p1") in the signed-in organisation.
package nasaqtest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/fadymondy/nasaq-theme-sdk/nasaq"
	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// Clinic is the "clinic" theme's tokens.
func Clinic() theme.Tokens {
	var t theme.Tokens
	t.Colors.Light = theme.Palette{Primary: "#0F7C78", Accent: "#E8A87C", Background: "#FBFEFD", Foreground: "#10302F"}
	t.Colors.Dark = theme.Palette{Primary: "#4FD1C5", Accent: "#E8A87C", Background: "#0B1F1E", Foreground: "#E6F4F2"}
	t.Radius, t.Density = "lg", "spacious"
	return t
}

// Published is the primary colour of the theme published on project p1: the
// clinic theme, customised.
const Published = "#7A1F5C"

// Server is the fake. Guard changes during requests with Mu.
type Server struct {
	*httptest.Server
	Mu sync.Mutex
	// Bound is every binding received.
	Bound []nasaq.Binding
	// Live is the theme published on p1; "" means nothing is published.
	Live string
	// Bearers is the Authorization header of every app-API call.
	Bearers []string
	// Forbid answers 403 to the app API (Nasaq not installed in the org).
	Forbid bool
}

// New starts a fake Nasaq for app. It closes when the test ends.
func New(t testing.TB, app string) *Server {
	n := NewServer(app)
	t.Cleanup(n.Close)
	return n
}

// NewServer starts a fake without test-owned cleanup. The caller must close it.
func NewServer(app string) *Server {
	n := &Server{Live: "clinic"}
	cover := map[string]any{app: map[string]any{"pages": []string{"book"}}}
	clinic := map[string]any{"id": "clinic", "version": "1.0.0", "name": map[string]string{"en": "Clinic", "ar": "عيادة"},
		"free": true, "coverage": cover, "tokens": Clinic()}
	shop := map[string]any{"id": "shop", "version": "1.0.0", "name": map[string]string{"en": "Shop"},
		"coverage": map[string]any{"another-app": true}, "tokens": Clinic()}
	n.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		switch {
		case r.URL.Path == "/api/themes":
			_ = enc.Encode(map[string]any{"items": []any{clinic, shop}})
			return
		case r.URL.Path == "/api/themes/clinic":
			_ = enc.Encode(clinic)
			return
		case r.URL.Path == "/api/themes/shop":
			_ = enc.Encode(shop)
			return
		case strings.HasPrefix(r.URL.Path, "/api/themes/"):
			w.WriteHeader(http.StatusNotFound)
			return
		}
		n.Mu.Lock()
		defer n.Mu.Unlock()
		n.Bearers = append(n.Bearers, r.Header.Get("Authorization"))
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if n.Forbid {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch {
		case r.URL.Path == "/api/cms/v1/projects":
			_ = enc.Encode(map[string]any{"items": []any{map[string]any{"id": "p1", "slug": "spa", "name": map[string]string{"en": "My spa"}}}})
		case r.URL.Path == "/api/cms/v1/projects/p1/bind" && r.Method == http.MethodPost:
			var b nasaq.Binding
			_ = json.NewDecoder(r.Body).Decode(&b)
			n.Bound = append(n.Bound, b)
			w.WriteHeader(http.StatusCreated)
			_ = enc.Encode(b)
		case r.URL.Path == "/api/cms/v1/projects/p1/theme/core" && r.URL.Query().Get("app") == app:
			if n.Live == "" {
				_ = enc.Encode(map[string]any{"theme": nil})
				return
			}
			tok := Clinic()
			tok.Colors.Light.Primary = Published
			_ = enc.Encode(map[string]any{"theme": n.Live, "version": "1.0.0", "tokens": tok})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return n
}
