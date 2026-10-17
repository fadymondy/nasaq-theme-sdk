package circlexo_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	cxo "github.com/circlexo/circlexo-go"
	"github.com/circlexo/circlexo-go/circlexotest"

	"github.com/fadymondy/nasaq-theme-sdk/circlexo"
	"github.com/fadymondy/nasaq-theme-sdk/nasaq"
	"github.com/fadymondy/nasaq-theme-sdk/nasaqtest"
	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// vault "seals" by prefixing; good enough to see a token never lands in plain text.
type vault struct{}

func (vault) Configured() bool { return true }
func (vault) Seal(s string) (string, error) {
	return "sealed:" + s, nil
}
func (vault) Unseal(s string) (string, error) {
	if !strings.HasPrefix(s, "sealed:") {
		return "", errors.New("not sealed")
	}
	return strings.TrimPrefix(s, "sealed:"), nil
}

// sites keeps each site's look in memory.
type sites struct {
	mu    sync.Mutex
	looks map[string]theme.Look
	saves int
}

func (s *sites) Look(_ context.Context, id string) (theme.Look, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.looks[id], nil
}

func (s *sites) SetLook(_ context.Context, id string, l theme.Look) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.looks[id] = l
	s.saves++
	return nil
}

func TestConnectBindPull(t *testing.T) {
	hub := circlexotest.New(t, "seatfor")
	n := nasaqtest.New(t, "seatfor")
	store := &circlexo.MemStore{}
	look := &sites{looks: map[string]theme.Look{"s1": {Action: theme.ColorPair{Light: "#123456", Dark: "#654321"}}}}
	c := &circlexo.Connector{
		Hub: func(context.Context) (cxo.Config, error) {
			return cxo.Config{Issuer: hub.URL, AppID: "seatfor", ClientSecret: hub.ClientSecret}, nil
		},
		NasaqAudience: func(context.Context) string { return "nasaq" },
		Nasaq:         nasaq.New("seatfor", nasaq.Static(n.URL)),
		Store:         store,
		Vault:         vault{},
		Site:          look,
	}
	ctx := context.Background()
	if !c.Configured(ctx) {
		t.Fatal("not configured")
	}
	if _, err := c.Pull(ctx, "s1"); !errors.Is(err, theme.ErrNotLinked) {
		t.Fatalf("pull before connect: %v", err)
	}

	const back = "https://app.example.test/panel/themes/circlexo"
	auth, err := c.ConnectStart(ctx, "s1", "owner-1", back)
	if err != nil {
		t.Fatal(err)
	}
	noFollow := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := noFollow.Get(auth)
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Body.Close(); err != nil {
		t.Fatal(err)
	}
	loc, err := url.Parse(res.Header.Get("Location"))
	if err != nil || !strings.HasPrefix(loc.String(), back) {
		t.Fatalf("hub sent us to %q (%v)", res.Header.Get("Location"), err)
	}
	code, state := loc.Query().Get("code"), loc.Query().Get("state")

	if _, err := c.ConnectFinish(ctx, "s1", "someone-else", code, state); !errors.Is(err, theme.ErrState) {
		t.Fatalf("another account finished the sign-in: %v", err)
	}
	if _, err := c.ConnectFinish(ctx, "s2", "owner-1", code, state); !errors.Is(err, theme.ErrState) {
		t.Fatalf("finished on another site: %v", err)
	}
	link, err := c.ConnectFinish(ctx, "s1", "owner-1", code, state)
	if err != nil {
		t.Fatal(err)
	}
	if link.OrgID != "org-1" || link.ProjectID != "" || link.LinkedBy != "owner-1" {
		t.Fatalf("link = %+v", link)
	}
	sealed := store.Sealed("s1")
	if !strings.HasPrefix(sealed, "sealed:") {
		t.Fatalf("refresh token stored as %q", sealed)
	}

	if _, err := c.Pull(ctx, "s1"); !errors.Is(err, theme.ErrNoProject) {
		t.Fatalf("pull before bind: %v", err)
	}
	ps, err := c.Projects(ctx, "s1")
	if err != nil || len(ps) != 1 || ps[0].ID != "p1" {
		t.Fatalf("projects = %+v %v", ps, err)
	}
	if _, err := c.Bind(ctx, "s1", "nope", nasaq.Binding{}); !errors.Is(err, theme.ErrUnknownProject) {
		t.Fatalf("bind unknown: %v", err)
	}
	link2, err := c.Bind(ctx, "s1", "p1", nasaq.Binding{AppRef: "acme", Role: "booking", MountPath: "/book", RenderURL: "https://app.example.test"})
	if err != nil {
		t.Fatal(err)
	}
	if link2.ProjectID != "p1" || link2.ProjectName != "My spa" || link2.LastPullAt == nil || link2.LastError != "" {
		t.Fatalf("after bind = %+v", link2)
	}
	if len(n.Bound) != 1 || n.Bound[0].App != "seatfor" || n.Bound[0].RenderURL != "https://app.example.test" {
		t.Fatalf("binding = %+v", n.Bound)
	}
	if store.Sealed("s1") == sealed {
		t.Fatal("refresh token did not rotate")
	}
	got := look.looks["s1"]
	if got.Action.Light != nasaqtest.Published || got.Theme.ProjectID != "p1" || got.Theme.OwnAction.Light != "#123456" {
		t.Fatalf("pulled look = %+v", got)
	}
	// Nothing changed upstream: a second pull is a no-op.
	if changed, err := c.Pull(ctx, "s1"); err != nil || changed || look.saves != 1 {
		t.Fatalf("second pull = %v %v (saves %d)", changed, err, look.saves)
	}

	// The sweep finds the bound site.
	if ids, _ := store.Bound(ctx); len(ids) != 1 || ids[0] != "s1" {
		t.Fatalf("bound = %v", ids)
	}
	var swept []string
	n.Mu.Lock()
	n.Live = ""
	n.Mu.Unlock()
	c.Sweep(ctx, nil, func(id string) { swept = append(swept, id) })
	if len(swept) != 0 {
		t.Fatalf("nothing published, yet swept %v", swept)
	}
	if l, _ := store.Link(ctx, "s1"); l.LastError != "" {
		t.Fatalf("nothing published is not an error: %q", l.LastError)
	}

	// Nasaq uninstalled from the org.
	n.Mu.Lock()
	n.Forbid = true
	n.Mu.Unlock()
	if _, err := c.Pull(ctx, "s1"); !errors.Is(err, theme.ErrNotInstalled) {
		t.Fatalf("forbidden: %v", err)
	}
	if l, _ := store.Link(ctx, "s1"); l == nil || l.LastError == "" {
		t.Fatalf("last_error not recorded: %+v", l)
	}

	// The grant is lost: the store says so, and the sweep skips the site.
	_ = store.WithRefresh(ctx, "s1", func(string) (string, error) { return "sealed:revoked", nil })
	if _, err := c.Pull(ctx, "s1"); !errors.Is(err, theme.ErrGrantLost) {
		t.Fatalf("lost grant: %v", err)
	}
	if l, _ := store.Link(ctx, "s1"); l.LastError != circlexo.GrantLost {
		t.Fatalf("last_error = %q", l.LastError)
	}
	if ids, _ := store.Bound(ctx); len(ids) != 0 {
		t.Fatalf("a lost grant is still swept: %v", ids)
	}

	if err := c.Disconnect(ctx, "s1"); err != nil {
		t.Fatal(err)
	}
	if l, _ := c.Link(ctx, "s1"); l != nil {
		t.Fatal("link survived disconnect")
	}
}

func TestNotConfigured(t *testing.T) {
	ctx := context.Background()
	c := &circlexo.Connector{}
	if _, err := c.ConnectStart(ctx, "s", "u", "https://x"); !errors.Is(err, theme.ErrNotConfigured) {
		t.Fatalf("no hub: %v", err)
	}
	c.Hub = func(context.Context) (cxo.Config, error) {
		return cxo.Config{Issuer: "https://hub", AppID: "a", ClientSecret: "s"}, nil
	}
	if _, err := c.ConnectStart(ctx, "s", "u", "https://x"); !errors.Is(err, theme.ErrNoVault) {
		t.Fatalf("no vault: %v", err)
	}
}

func TestLimiter(t *testing.T) {
	var l circlexo.Limiter
	now := time.Now()
	if !l.Allow("a", now) || l.Allow("a", now.Add(time.Second)) || !l.Allow("b", now) || !l.Allow("a", now.Add(circlexo.PullGap)) {
		t.Fatal("rate limit wrong")
	}
}
