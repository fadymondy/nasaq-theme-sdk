package theme

import (
	"fmt"
	"regexp"
)

// SiteTheme is the theme a site uses: the app's own default look, or a theme
// published from Nasaq UI Studio. The action colour itself is not here: an app
// keeps it with its brand (see Look), so a theme and a hand-picked brand colour
// are one and the same thing to every page.
//
// The rest is what a Nasaq theme carries beyond the action colour: the ground
// and ink per theme, the accent, the corner radius and the density.
type SiteTheme struct {
	// Source is "default" (the app's own look) or "nasaq". Empty reads as default.
	Source string `json:"source,omitempty"`
	// ID and Version name the Nasaq theme (e.g. "clinic", "1.2.0").
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
	Name    Text   `json:"name,omitempty"`
	// ProjectID is the Nasaq project the theme was published from, when the
	// owner customised it there.
	ProjectID string `json:"project_id,omitempty"`
	// AppliedAt is RFC 3339.
	AppliedAt string `json:"applied_at,omitempty"`

	Accent     ColorPair `json:"accent,omitempty"`
	Background ColorPair `json:"background,omitempty"`
	Foreground ColorPair `json:"foreground,omitempty"`
	// Radius is Nasaq's scale: none, sm, md, lg, pill.
	Radius string `json:"radius,omitempty"`
	// Density is compact, comfortable or spacious.
	Density string `json:"density,omitempty"`

	// OwnAction is the site's own action colour from before its first Nasaq
	// theme: "reset to default" puts it back.
	OwnAction ColorPair `json:"own_action,omitempty"`
}

// Theme sources.
const (
	SourceDefault = "default"
	SourceNasaq   = "nasaq"
)

// Radii and Densities are the values a theme may carry.
var (
	Radii     = []string{"none", "sm", "md", "lg", "pill"}
	Densities = []string{"compact", "comfortable", "spacious"}
)

var (
	idRE        = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	projectIDRE = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)
	versionRE   = regexp.MustCompile(`^[0-9A-Za-z.+-]{1,32}$`)
)

// ValidID reports whether id can name a Nasaq theme (and so is safe in a URL path).
func ValidID(id string) bool { return idRE.MatchString(id) }

func oneOf(v string, set []string) bool {
	for _, s := range set {
		if v == s {
			return true
		}
	}
	return false
}

// Problems is every reason the theme cannot be stored, in English, worded for
// the field ("theme.radius must be ..."). Empty means it is fine.
func (t SiteTheme) Problems() []string {
	var out []string
	errf := func(f string, a ...any) { out = append(out, fmt.Sprintf(f, a...)) }
	switch t.Source {
	case "", SourceDefault, SourceNasaq:
	default:
		errf(`theme.source must be "default" or "nasaq"`)
	}
	if t.Source == SourceNasaq && !idRE.MatchString(t.ID) {
		errf("theme.id must be a Nasaq theme id")
	}
	if t.ID != "" && !idRE.MatchString(t.ID) {
		errf("theme.id must be lowercase letters, digits, '.', '_' or '-'")
	}
	if t.Version != "" && !versionRE.MatchString(t.Version) {
		errf("theme.version is not a version")
	}
	if t.ProjectID != "" && !projectIDRE.MatchString(t.ProjectID) {
		errf("theme.project_id is not an id")
	}
	for _, pair := range []struct {
		field string
		c     ColorPair
	}{{"theme.accent", t.Accent}, {"theme.background", t.Background}, {"theme.foreground", t.Foreground}, {"theme.own_action", t.OwnAction}} {
		if pair.c.Light != "" && !IsHex(pair.c.Light) {
			errf("%s.light must be #RRGGBB", pair.field)
		}
		if pair.c.Dark != "" && !IsHex(pair.c.Dark) {
			errf("%s.dark must be #RRGGBB", pair.field)
		}
	}
	// Ground and ink go together: one without the other cannot be checked.
	for _, g := range []struct{ name, bg, fg string }{
		{"light", t.Background.Light, t.Foreground.Light},
		{"dark", t.Background.Dark, t.Foreground.Dark},
	} {
		if (g.bg == "") != (g.fg == "") {
			errf("theme: the %s background and foreground are set together", g.name)
			continue
		}
		if IsHex(g.bg) && IsHex(g.fg) && Contrast(g.bg, g.fg) < MinInkContrast {
			errf("theme: text on the %s background is %.2f:1; at least %.1f:1 is required", g.name, Contrast(g.bg, g.fg), MinInkContrast)
		}
	}
	if t.Radius != "" && !oneOf(t.Radius, Radii) {
		errf("theme.radius must be none, sm, md, lg or pill")
	}
	if t.Density != "" && !oneOf(t.Density, Densities) {
		errf("theme.density must be compact, comfortable or spacious")
	}
	return out
}

// PublicSiteTheme is the part of the theme the pages render.
type PublicSiteTheme struct {
	Source     string    `json:"source"`
	ID         string    `json:"id,omitempty"`
	Accent     ColorPair `json:"accent,omitempty"`
	Background ColorPair `json:"background,omitempty"`
	Foreground ColorPair `json:"foreground,omitempty"`
	Radius     string    `json:"radius,omitempty"`
	Density    string    `json:"density,omitempty"`
}

// Public projects the theme onto what a page needs.
func (t SiteTheme) Public() PublicSiteTheme {
	src := t.Source
	if src == "" {
		src = SourceDefault
	}
	return PublicSiteTheme{
		Source: src, ID: t.ID, Accent: t.Accent, Background: t.Background,
		Foreground: t.Foreground, Radius: t.Radius, Density: t.Density,
	}
}

// PublicOrNil is Public for an applied theme and nil for none, so a site that
// never chose a theme projects exactly as before themes existed.
func (t SiteTheme) PublicOrNil() *PublicSiteTheme {
	if !t.Applied() {
		return nil
	}
	p := t.Public()
	return &p
}

// Applied reports whether a theme was ever applied (the default counts).
func (t SiteTheme) Applied() bool { return t.Source != "" || t.ID != "" }
