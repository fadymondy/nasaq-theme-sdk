package product

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// PackageSchema is the version of the Package wire format.
const PackageSchema = 1

// MaxPackageCSS caps the custom CSS a package may carry.
const MaxPackageCSS = 64 << 10

// Package is a published Studio theme as a product receives it: Studio's theme core (brand,
// tokens, layout, css) plus the product id and the theme's editable content (the hero).
// It is sent as JSON with PUT <app>/api/nasaq/theme.
type Package struct {
	Schema   int            `json:"schema"`
	Product  string         `json:"product"`
	Theme    string         `json:"theme"`
	Version  string         `json:"version"`
	Locale   string         `json:"locale,omitempty"`
	Brand    Brand          `json:"brand"`
	Tokens   Tokens         `json:"tokens"`
	Layout   map[string]any `json:"layout,omitempty"`
	CSS      string         `json:"css,omitempty"`
	Content  Content        `json:"content"`
	Coverage map[string]any `json:"coverage,omitempty"`
}

// Brand is the store identity a theme carries.
type Brand struct {
	Name      Text   `json:"name,omitempty"`
	Tagline   Text   `json:"tagline,omitempty"`
	LogoLight string `json:"logoLight,omitempty"`
	LogoDark  string `json:"logoDark,omitempty"`
	Favicon   string `json:"favicon,omitempty"`
}

// Tokens are the theme's design tokens. Colours are open sets (primary, accent, background,
// foreground, muted, border, onPrimary...) per mode.
type Tokens struct {
	Colors struct {
		Light map[string]string `json:"light"`
		Dark  map[string]string `json:"dark,omitempty"`
	} `json:"colors"`
	Fonts   map[string]string `json:"fonts,omitempty"`
	Radius  string            `json:"radius,omitempty"`
	Density string            `json:"density,omitempty"`
}

// Content is the theme content a product shows outside the token layer.
type Content struct {
	Hero Hero `json:"hero"`
	// Announcement is the bar above the header.
	Announcement Text `json:"announcement,omitempty"`
}

// Hero is the home page hero.
type Hero struct {
	Heading Text   `json:"heading,omitempty"`
	Sub     Text   `json:"sub,omitempty"`
	CTA     Text   `json:"cta,omitempty"`
	Image   string `json:"image,omitempty"`
}

var themeIDRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

var knownRadius = map[string]bool{"": true, "none": true, "sm": true, "md": true, "lg": true, "xl": true, "pill": true}

// Validate checks a package before a product applies it. It refuses what cannot be applied
// safely (no product or theme, no readable primary colour, oversized or escaping CSS) and
// leaves the rest to the product's own mapping.
func (p Package) Validate() error {
	if p.Schema != PackageSchema {
		return fmt.Errorf("product: package schema %d, want %d", p.Schema, PackageSchema)
	}
	if p.Product == "" {
		return fmt.Errorf("product: package has no product")
	}
	if !themeIDRE.MatchString(p.Theme) {
		return fmt.Errorf("product: bad theme id %q", p.Theme)
	}
	if theme.Hex(p.Tokens.Colors.Light["primary"]) == "" {
		return fmt.Errorf("product: package needs a #RRGGBB light primary colour")
	}
	if !knownRadius[p.Tokens.Radius] {
		return fmt.Errorf("product: unknown radius %q", p.Tokens.Radius)
	}
	if len(p.CSS) > MaxPackageCSS {
		return fmt.Errorf("product: css is %d bytes, max %d", len(p.CSS), MaxPackageCSS)
	}
	if l := strings.ToLower(p.CSS); strings.Contains(l, "</style") || strings.Contains(l, "@import") {
		return fmt.Errorf("product: css may not close the style element or import")
	}
	return nil
}

// Color is a validated colour of the package (light mode unless dark), or "".
func (p Package) Color(key string, dark bool) string {
	if dark {
		if c := theme.Hex(p.Tokens.Colors.Dark[key]); c != "" {
			return c
		}
	}
	return theme.Hex(p.Tokens.Colors.Light[key])
}
