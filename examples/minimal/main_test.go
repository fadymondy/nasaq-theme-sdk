package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fadymondy/nasaq-theme-sdk/nasaq"
	"github.com/fadymondy/nasaq-theme-sdk/nasaqtest"
)

func TestCatalogAndSelectedCSS(t *testing.T) {
	fake := nasaqtest.New(t, app)
	h := handler(nasaq.New(app, nasaq.Static(fake.URL)))
	for _, tc := range []struct {
		path, contentType, contains string
	}{
		{"/", "text/html", "Clinic"},
		{"/?theme=clinic", "text/html", "theme.css?id=clinic"},
		{"/theme.css?id=clinic", "text/css", "--primary: #0F7C78"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != http.StatusOK || !strings.HasPrefix(response.Header().Get("Content-Type"), tc.contentType) || !strings.Contains(response.Body.String(), tc.contains) {
				t.Fatalf("response = %d %v %s", response.Code, response.Header(), response.Body.String())
			}
		})
	}
	for _, id := range []string{"shop", "unknown", ""} {
		response := httptest.NewRecorder()
		h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/theme.css?id="+id, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("theme %q: status %d", id, response.Code)
		}
	}
	for _, id := range []string{"shop", "unknown"} {
		response := httptest.NewRecorder()
		h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/?theme="+id, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("selection %q: status %d", id, response.Code)
		}
	}
}
