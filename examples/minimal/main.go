package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/fadymondy/nasaq-theme-sdk/nasaq"
	"github.com/fadymondy/nasaq-theme-sdk/nasaqtest"
	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

const app = "minimal-example"

var page = template.Must(template.New("catalog").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>Nasaq themes</title>
<link rel="stylesheet" href="/theme.css?id={{.Selected}}"></head>
<body><h1>Choose a theme</h1><ul>{{range .Themes}}
<li><a href="/?theme={{.ID}}">{{index .Name "en"}}</a></li>{{end}}</ul>
<p>This example uses a local fake catalog.</p></body></html>`))

func handler(client *nasaq.Client) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		themes, err := client.Catalog(r.Context())
		if err != nil || len(themes) == 0 {
			http.Error(w, "The theme catalog is unavailable.", http.StatusServiceUnavailable)
			return
		}
		selected := r.URL.Query().Get("theme")
		if selected == "" {
			selected = themes[0].ID
		}
		found := false
		for _, available := range themes {
			if available.ID == selected {
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Choose a theme from the catalog.", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(w, struct {
			Themes   []nasaq.CatalogTheme
			Selected string
		}{themes, selected}); err != nil {
			log.Printf("render catalog: %v", err)
		}
	})
	mux.HandleFunc("GET /theme.css", func(w http.ResponseWriter, r *http.Request) {
		selected, err := client.Theme(r.Context(), r.URL.Query().Get("id"))
		if errors.Is(err, theme.ErrUnknownTheme) || errors.Is(err, theme.ErrNotForApp) {
			http.NotFound(w, r)
			return
		}
		if err != nil || selected.Tokens == nil {
			http.Error(w, "The selected theme is unavailable.", http.StatusServiceUnavailable)
			return
		}
		colors := selected.Tokens.Colors.Light
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		fmt.Fprintf(w, ":root { --primary: %s; --background: %s; --foreground: %s; }\nbody { background: var(--background); color: var(--foreground); }\na { color: var(--primary); }\n",
			theme.Hex(colors.Primary), theme.Hex(colors.Background), theme.Hex(colors.Foreground))
	})
	return mux
}

func run() error {
	fake := nasaqtest.NewServer(app)
	defer fake.Close()
	client := nasaq.New(app, nasaq.Static(fake.URL))
	server := &http.Server{Addr: "127.0.0.1:8080", Handler: handler(client), ReadHeaderTimeout: 5 * time.Second}
	log.Print("Open http://127.0.0.1:8080 (local fake catalog; no credentials needed)")
	return server.ListenAndServe()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
