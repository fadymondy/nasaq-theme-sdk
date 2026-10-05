// Package nasaq reads themes from Nasaq UI Studio: the public catalog (no
// account needed) and, with a CircleXO token for Nasaq's audience, the
// organisation's projects, binding an app to one, and the theme published on it.
package nasaq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// CatalogTheme is one theme in Nasaq's public catalog.
type CatalogTheme struct {
	ID          string     `json:"id"`
	Version     string     `json:"version"`
	Name        theme.Text `json:"name"`
	Description theme.Text `json:"description,omitempty"`
	Category    string     `json:"category,omitempty"`
	Free        bool       `json:"free"`
	Previews    struct {
		Thumbnail     string   `json:"thumbnail,omitempty"`
		ThumbnailDark string   `json:"thumbnailDark,omitempty"`
		Screens       []string `json:"screens,omitempty"`
	} `json:"previews"`
	DemoURL  string                     `json:"demoUrl,omitempty"`
	Coverage map[string]json.RawMessage `json:"coverage,omitempty"`
	Tokens   *theme.Tokens              `json:"tokens,omitempty"`
}

// Covers reports whether the theme styles app's pages.
func (t CatalogTheme) Covers(app string) bool { _, ok := t.Coverage[app]; return ok }

// Core is what Nasaq answers for a project's published theme, as one app sees
// it (GET /api/cms/v1/projects/{id}/theme/core?app=...). Theme is empty when
// the project has no live theme yet.
type Core struct {
	Theme   string `json:"theme"`
	Version string `json:"version"`
	Brand   struct {
		Name theme.Text `json:"name,omitempty"`
	} `json:"brand"`
	Tokens theme.Tokens `json:"tokens"`
}

// Project is a Nasaq project (a site) in the owner's organisation.
type Project struct {
	ID            string           `json:"id"`
	Slug          string           `json:"slug"`
	Name          theme.Text       `json:"name"`
	IsPrimary     bool             `json:"is_primary"`
	DefaultLocale string           `json:"default_locale"`
	Status        string           `json:"status"`
	Bindings      []ProjectBinding `json:"bindings,omitempty"`
}

// Label is the project's name for people: English, Arabic, then the slug.
func (p Project) Label() string {
	if n := p.Name["en"]; n != "" {
		return n
	}
	if n := p.Name["ar"]; n != "" {
		return n
	}
	return p.Slug
}

// ProjectBinding is an app bound to a Nasaq project.
type ProjectBinding struct {
	App    string `json:"app"`
	AppRef string `json:"app_ref"`
}

// Binding is an app bound to a Nasaq project: Nasaq learns where the app
// renders the site, so it can say when the theme changes.
type Binding struct {
	App       string `json:"app"`
	AppRef    string `json:"app_ref"`
	Role      string `json:"role"`
	MountPath string `json:"mount_path"`
	RenderURL string `json:"render_url"`
}

// ErrUnavailable: Nasaq did not answer, or answered with an error.
var ErrUnavailable = errors.New("nasaq: unavailable")

// StatusError is Nasaq answering with a status other than 2xx. It unwraps to
// ErrUnavailable; a 403 means the organisation has not installed Nasaq.
type StatusError struct {
	Status int
	Body   string
}

func (e *StatusError) Error() string { return fmt.Sprintf("nasaq: %d %s", e.Status, e.Body) }
func (e *StatusError) Unwrap() error { return ErrUnavailable }

// DefaultTTL is how long the catalog and a theme's detail are cached.
const DefaultTTL = 10 * time.Minute

// Client talks to Nasaq on behalf of one app.
type Client struct {
	// App is the app as Nasaq knows it: the key a theme's coverage names and a
	// project binds (e.g. "seatfor", "matjar").
	App string
	// URL is Nasaq's base URL, read per call so an operator can change it at
	// run time. Use Static for a fixed one.
	URL func(context.Context) string
	// HTTP is used for every call; nil means a client with a 15 s timeout.
	HTTP *http.Client
	// TTL is how long the catalog is cached; 0 means DefaultTTL.
	TTL time.Duration

	mu    sync.Mutex
	list  map[string]cached[[]CatalogTheme] // by Nasaq URL
	items map[string]cached[CatalogTheme]   // by Nasaq URL + id
}

type cached[T any] struct {
	at  time.Time
	val T
}

// Static is a URL func for a fixed base URL.
func Static(base string) func(context.Context) string {
	base = strings.TrimSuffix(base, "/")
	return func(context.Context) string { return base }
}

// New is a client for app against the Nasaq at url.
func New(app string, url func(context.Context) string) *Client {
	return &Client{App: app, URL: url}
}

var defaultHTTP = &http.Client{Timeout: 15 * time.Second}

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return defaultHTTP
}

func (c *Client) ttl() time.Duration {
	if c.TTL > 0 {
		return c.TTL
	}
	return DefaultTTL
}

func (c *Client) base(ctx context.Context) string {
	if c.URL == nil {
		return ""
	}
	return strings.TrimSuffix(c.URL(ctx), "/")
}

// call does one JSON request against Nasaq. token is a hub access token for
// Nasaq's audience, or empty for the public catalog.
func (c *Client) call(ctx context.Context, method, path, token string, body, out any) error {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = strings.NewReader(string(b))
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base(ctx)+path, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := c.http().Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return &StatusError{Status: res.StatusCode, Body: strings.TrimSpace(string(b))}
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(out); err != nil {
		return fmt.Errorf("%w: decode %s: %v", ErrUnavailable, path, err)
	}
	return nil
}

// ---- the public catalog -------------------------------------------------------

// Catalog is Nasaq's themes that cover the app, without their tokens (the list
// is for choosing). A stale copy is served when Nasaq is down.
func (c *Client) Catalog(ctx context.Context) ([]CatalogTheme, error) {
	key := c.base(ctx)
	c.mu.Lock()
	hit, ok := c.list[key]
	c.mu.Unlock()
	if ok && time.Since(hit.at) < c.ttl() {
		return hit.val, nil
	}
	var res struct {
		Items []CatalogTheme `json:"items"`
	}
	if err := c.call(ctx, http.MethodGet, "/api/themes", "", nil, &res); err != nil {
		if ok {
			return hit.val, nil
		}
		return nil, err
	}
	out := make([]CatalogTheme, 0, len(res.Items))
	for _, t := range res.Items {
		if t.Covers(c.App) {
			t.Tokens = nil
			out = append(out, t)
		}
	}
	c.mu.Lock()
	if c.list == nil {
		c.list = map[string]cached[[]CatalogTheme]{}
	}
	c.list[key] = cached[[]CatalogTheme]{at: time.Now(), val: out}
	c.mu.Unlock()
	return out, nil
}

// Theme is one catalog theme with its tokens. It refuses with
// theme.ErrUnknownTheme and theme.ErrNotForApp.
func (c *Client) Theme(ctx context.Context, id string) (CatalogTheme, error) {
	if !theme.ValidID(id) {
		return CatalogTheme{}, theme.ErrUnknownTheme
	}
	key := c.base(ctx) + "|" + id
	c.mu.Lock()
	hit, ok := c.items[key]
	c.mu.Unlock()
	if ok && time.Since(hit.at) < c.ttl() {
		return hit.val, nil
	}
	var t CatalogTheme
	if err := c.call(ctx, http.MethodGet, "/api/themes/"+url.PathEscape(id), "", nil, &t); err != nil {
		var se *StatusError
		if errors.As(err, &se) && se.Status == http.StatusNotFound {
			return CatalogTheme{}, theme.ErrUnknownTheme
		}
		if ok {
			return hit.val, nil
		}
		return CatalogTheme{}, err
	}
	if !t.Covers(c.App) {
		return CatalogTheme{}, theme.ErrNotForApp
	}
	if t.Tokens == nil {
		return CatalogTheme{}, fmt.Errorf("%w: theme %s has no tokens", ErrUnavailable, id)
	}
	c.mu.Lock()
	if c.items == nil {
		c.items = map[string]cached[CatalogTheme]{}
	}
	c.items[key] = cached[CatalogTheme]{at: time.Now(), val: t}
	c.mu.Unlock()
	return t, nil
}

// ResetCache drops the cached catalog.
func (c *Client) ResetCache() {
	c.mu.Lock()
	c.list, c.items = nil, nil
	c.mu.Unlock()
}

// CustomizeURL is Nasaq's customizer for a theme, in the owner's organisation
// when org is set. locale is "ar" or "en".
func (c *Client) CustomizeURL(ctx context.Context, locale, themeID, org string) string {
	if locale != "en" {
		locale = "ar"
	}
	q := url.Values{"template": {themeID}}
	if org != "" {
		q.Set("org", org)
	}
	return c.base(ctx) + "/" + locale + "/customize?" + q.Encode()
}

// ---- the app API (a hub token for Nasaq's audience) ------------------------------

// Projects lists the projects in the token's organisation.
func (c *Client) Projects(ctx context.Context, token string) ([]Project, error) {
	var res struct {
		Items []Project `json:"items"`
	}
	if err := c.call(ctx, http.MethodGet, "/api/cms/v1/projects", token, nil, &res); err != nil {
		return nil, err
	}
	return res.Items, nil
}

// Bind binds the app to a project. b.App defaults to the client's app.
func (c *Client) Bind(ctx context.Context, token, projectID string, b Binding) error {
	if b.App == "" {
		b.App = c.App
	}
	return c.call(ctx, http.MethodPost, "/api/cms/v1/projects/"+url.PathEscape(projectID)+"/bind", token, b, nil)
}

// Core is the theme published on a project for the app, or nil when nothing is
// published yet.
func (c *Client) Core(ctx context.Context, token, projectID string) (*Core, error) {
	var core Core
	path := "/api/cms/v1/projects/" + url.PathEscape(projectID) + "/theme/core?app=" + url.QueryEscape(c.App)
	if err := c.call(ctx, http.MethodGet, path, token, nil, &core); err != nil {
		return nil, err
	}
	if core.Theme == "" {
		return nil, nil
	}
	return &core, nil
}

// Refusal turns a failed app-API call into a refusal the owner can act on: a
// 403 is "Nasaq is not installed", a 404 "no such project", the rest upstream.
func Refusal(err error) error {
	var se *StatusError
	if errors.As(err, &se) {
		switch se.Status {
		case http.StatusForbidden:
			return theme.ErrNotInstalled
		case http.StatusNotFound:
			return theme.ErrUnknownProject
		}
	}
	var te *theme.Error
	if errors.As(err, &te) {
		return err
	}
	return fmt.Errorf("%w: %v", theme.ErrUpstream, err)
}
