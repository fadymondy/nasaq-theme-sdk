package nasaq_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fadymondy/nasaq-theme-sdk/nasaq"
	"github.com/fadymondy/nasaq-theme-sdk/nasaqtest"
	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

func TestCatalogKeepsTheAppsThemes(t *testing.T) {
	n := nasaqtest.New(t, "seatfor")
	c := nasaq.New("seatfor", nasaq.Static(n.URL))
	ctx := context.Background()
	list, err := c.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != "clinic" || list[0].Tokens != nil {
		t.Fatalf("catalog = %+v", list)
	}
	if th, err := c.Theme(ctx, "clinic"); err != nil || th.Tokens == nil || th.Tokens.Colors.Light.Primary != "#0F7C78" {
		t.Fatalf("clinic = %+v %v", th, err)
	}
	if _, err := c.Theme(ctx, "shop"); !errors.Is(err, theme.ErrNotForApp) {
		t.Fatalf("shop: %v", err)
	}
	if _, err := c.Theme(ctx, "nope"); !errors.Is(err, theme.ErrUnknownTheme) {
		t.Fatalf("nope: %v", err)
	}
	if _, err := c.Theme(ctx, "../x"); !errors.Is(err, theme.ErrUnknownTheme) {
		t.Fatalf("bad id: %v", err)
	}
	// Nasaq down: the cached list and theme still answer.
	n.Close()
	if list, err := c.Catalog(ctx); err != nil || len(list) != 1 {
		t.Fatalf("stale catalog = %v %v", list, err)
	}
	if _, err := c.Theme(ctx, "clinic"); err != nil {
		t.Fatalf("stale theme: %v", err)
	}
	c.ResetCache()
	if _, err := c.Catalog(ctx); !errors.Is(err, nasaq.ErrUnavailable) {
		t.Fatalf("down without a cache: %v", err)
	}
}

func TestAppAPI(t *testing.T) {
	n := nasaqtest.New(t, "matjar")
	c := nasaq.New("matjar", nasaq.Static(n.URL+"/"))
	ctx := context.Background()
	if _, err := c.Projects(ctx, ""); nasaq.Refusal(err) == nil {
		t.Fatal("no token answered")
	}
	ps, err := c.Projects(ctx, "tok")
	if err != nil || len(ps) != 1 || ps[0].Label() != "My spa" {
		t.Fatalf("projects = %+v %v", ps, err)
	}
	if err := c.Bind(ctx, "tok", "p1", nasaq.Binding{AppRef: "shop-1"}); err != nil {
		t.Fatal(err)
	}
	if len(n.Bound) != 1 || n.Bound[0].App != "matjar" {
		t.Fatalf("bound = %+v", n.Bound)
	}
	core, err := c.Core(ctx, "tok", "p1")
	if err != nil || core == nil || core.Tokens.Colors.Light.Primary != nasaqtest.Published {
		t.Fatalf("core = %+v %v", core, err)
	}
	n.Live = ""
	if core, err := c.Core(ctx, "tok", "p1"); err != nil || core != nil {
		t.Fatalf("nothing published = %+v %v", core, err)
	}
	if _, err := c.Core(ctx, "tok", "p2"); !errors.Is(nasaq.Refusal(err), theme.ErrUnknownProject) {
		t.Fatalf("unknown project: %v", err)
	}
	n.Forbid = true
	if _, err := c.Projects(ctx, "tok"); !errors.Is(nasaq.Refusal(err), theme.ErrNotInstalled) {
		t.Fatalf("forbidden: %v", err)
	}
	u := c.CustomizeURL(ctx, "fr", "clinic", "org-1")
	if !strings.HasPrefix(u, n.URL+"/ar/customize?") || !strings.Contains(u, "org=org-1") || !strings.Contains(u, "template=clinic") {
		t.Fatalf("customize = %s", u)
	}
}
