package matjar

import (
	"strings"

	"github.com/fadymondy/nasaq-theme-sdk/product"
)

// Brand is the part of a Matjar store brand a Studio theme decides. It mirrors the fields
// of Matjar's commerce.Brand with the same JSON names, so Matjar copies it over its own
// brand (keeping logos, social links and layout) and normalizes the result.
type Brand struct {
	Primary      string            `json:"primary"`
	Accent       string            `json:"accent"`
	Mode         string            `json:"mode"`   // light | system
	Radius       string            `json:"radius"` // sharp | rounded | pill
	Font         string            `json:"font"`   // plex-arabic | inter
	Preset       string            `json:"preset"` // "nasaq:<theme>@<version>"
	Tokens       map[string]string `json:"tokens"`
	Announcement struct {
		Enabled bool   `json:"enabled"`
		EN      string `json:"en"`
		AR      string `json:"ar"`
	} `json:"announcement"`
	Hero struct {
		HeadingEN string `json:"headingEn"`
		HeadingAR string `json:"headingAr"`
		SubEN     string `json:"subEn"`
		SubAR     string `json:"subAr"`
		CtaEN     string `json:"ctaEn"`
		CtaAR     string `json:"ctaAr"`
	} `json:"hero"`
}

// surfaceTokens are the Studio colours that become Matjar's Nasaq variables. Matjar derives
// both modes from primary and accent and applies custom tokens to every mode, so the ground
// and ink colours are carried only for light-only themes (the store then opens in light).
var surfaceTokens = map[string]string{
	"background": "--nq-bg",
	"foreground": "--nq-fg",
	"muted":      "--nq-surface-soft",
	"border":     "--nq-line",
}

// BrandOf maps a published package onto a Matjar brand.
func BrandOf(p product.Package) Brand {
	var b Brand
	b.Primary = p.Color("primary", false)
	b.Accent = p.Color("accent", false)
	if b.Accent == "" {
		b.Accent = b.Primary
	}
	b.Radius = RadiusOf(p.Tokens.Radius)
	b.Font = FontOf(p.Tokens.Fonts)
	b.Preset = "nasaq:" + p.Theme
	if p.Version != "" {
		b.Preset += "@" + p.Version
	}
	if len(b.Preset) > 40 {
		b.Preset = b.Preset[:40]
	}
	b.Tokens = map[string]string{}
	b.Mode = "system"
	if len(p.Tokens.Colors.Dark) == 0 {
		b.Mode = "light"
		for k, v := range surfaceTokens {
			if c := p.Color(k, false); c != "" {
				b.Tokens[v] = c
			}
		}
	}
	h := p.Content.Hero
	b.Hero.HeadingEN, b.Hero.HeadingAR = clip(h.Heading["en"], 120), clip(h.Heading["ar"], 120)
	b.Hero.SubEN, b.Hero.SubAR = clip(h.Sub["en"], 300), clip(h.Sub["ar"], 300)
	b.Hero.CtaEN, b.Hero.CtaAR = clip(h.CTA["en"], 40), clip(h.CTA["ar"], 40)
	if a := p.Content.Announcement; a["en"] != "" || a["ar"] != "" {
		b.Announcement.Enabled = true
		b.Announcement.EN, b.Announcement.AR = clip(a["en"], 200), clip(a["ar"], 200)
	}
	return b
}

// RadiusOf maps a Nasaq radius to Matjar's corner style.
func RadiusOf(r string) string {
	switch r {
	case "none", "sm":
		return "sharp"
	case "pill", "xl":
		return "pill"
	}
	return "rounded"
}

// FontOf picks Matjar's font: Inter when the theme's Latin face is Inter, IBM Plex Sans Arabic
// ("plex-arabic") otherwise (Plex carries all Arabic in every Matjar store).
func FontOf(fonts map[string]string) string {
	for _, k := range []string{"latin", "headingLatin", "sans"} {
		if strings.Contains(strings.ToLower(fonts[k]), "inter") {
			return "inter"
		}
	}
	return "plex-arabic"
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) > n {
		return strings.TrimSpace(string(r[:n]))
	}
	return s
}
