package theme

import (
	"encoding/json"
	"time"
)

// Palette is one theme's colours (Nasaq's tokens.colors.light / .dark).
type Palette struct {
	Primary    string `json:"primary,omitempty"`
	Accent     string `json:"accent,omitempty"`
	Background string `json:"background,omitempty"`
	Foreground string `json:"foreground,omitempty"`
	Muted      string `json:"muted,omitempty"`
	Border     string `json:"border,omitempty"`
}

// Tokens are a Nasaq theme's design tokens.
type Tokens struct {
	Colors struct {
		Light Palette `json:"light"`
		Dark  Palette `json:"dark"`
	} `json:"colors"`
	Fonts   map[string]string `json:"fonts,omitempty"`
	Radius  string            `json:"radius,omitempty"`
	Density string            `json:"density,omitempty"`
}

// Look is everything a theme decides about a site: the action colour and the
// text on it (which an app keeps with its brand), and the rest of the theme.
type Look struct {
	Action   ColorPair `json:"action"`
	OnAction ColorPair `json:"on_action"`
	Theme    SiteTheme `json:"theme"`
}

// Source says where a theme came from.
type Source struct {
	ID        string
	Version   string
	Name      Text
	ProjectID string // set when pulled from the site's Nasaq project
}

// Map puts a Nasaq theme on a site's look: the primary colour becomes the
// action colour (with a readable text colour), the rest goes into the theme. A
// piece that would not render well is left out rather than refusing the whole
// theme: ground and ink below AA, a colour that is not #RRGGBB, a radius or
// density the model does not know. Only a missing or unreadable primary refuses
// it (ErrUnusable).
func Map(cur Look, src Source, tok Tokens, now time.Time) (Look, error) {
	l, d := tok.Colors.Light, tok.Colors.Dark
	actL, actD := Hex(l.Primary), Hex(d.Primary)
	if actD == "" {
		actD = actL
	}
	if actL == "" {
		return cur, ErrUnusable
	}
	onL, onD := OnActionFor(actL), OnActionFor(actD)
	if Contrast(actL, onL) < MinActionContrast || Contrast(actD, onD) < MinActionContrast {
		return cur, ErrUnusable
	}

	own := cur.Theme.OwnAction
	if cur.Theme.Source != SourceNasaq {
		own = cur.Action // the site's own colour, kept for "reset"
	}
	t := SiteTheme{
		Source: SourceNasaq, ID: src.ID, Version: src.Version, Name: src.Name, ProjectID: src.ProjectID,
		AppliedAt: now.UTC().Format(time.RFC3339), OwnAction: own,
		Accent: ColorPair{Light: Hex(l.Accent), Dark: Hex(d.Accent)},
	}
	if t.Accent.Dark == "" {
		t.Accent.Dark = t.Accent.Light
	}
	if t.Accent.Light == "" {
		t.Accent = ColorPair{}
	}
	// Ground and ink: both or neither, per theme, and readable.
	ground := func(bg, fg string) (string, string) {
		bg, fg = Hex(bg), Hex(fg)
		if bg == "" || fg == "" || Contrast(bg, fg) < MinInkContrast {
			return "", ""
		}
		return bg, fg
	}
	t.Background.Light, t.Foreground.Light = ground(l.Background, l.Foreground)
	t.Background.Dark, t.Foreground.Dark = ground(d.Background, d.Foreground)
	if oneOf(tok.Radius, Radii) {
		t.Radius = tok.Radius
	}
	if oneOf(tok.Density, Densities) {
		t.Density = tok.Density
	}

	return Look{
		Action:   ColorPair{Light: actL, Dark: actD},
		OnAction: ColorPair{Light: onL, Dark: onD},
		Theme:    t,
	}, nil
}

// Default puts the app's own look back: the site's own action colour (if a
// Nasaq theme had replaced it) and no theme tokens.
func Default(cur Look, now time.Time) Look {
	if own := cur.Theme.OwnAction; cur.Theme.Source == SourceNasaq && Hex(own.Light) != "" {
		dark := Hex(own.Dark)
		if dark == "" {
			dark = Hex(own.Light)
		}
		cur.Action = ColorPair{Light: Hex(own.Light), Dark: dark}
		cur.OnAction = ColorPair{Light: OnActionFor(cur.Action.Light), Dark: OnActionFor(dark)}
	}
	cur.Theme = SiteTheme{Source: SourceDefault, AppliedAt: now.UTC().Format(time.RFC3339)}
	return cur
}

// Same reports whether two looks render the same (when each was applied aside).
func Same(a, b Look) bool {
	a.Theme.AppliedAt, b.Theme.AppliedAt = "", ""
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return string(ja) == string(jb)
}
