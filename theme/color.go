// Package theme is the site-theme model shared by every app that renders Nasaq
// UI Studio themes: the theme an app stores per site, the checks it must pass,
// and Map, which turns a Nasaq theme's tokens into the colours a page paints.
//
// It has no I/O. The nasaq package reads themes from Nasaq; the circlexo
// package keeps a site in step with the theme its owner publishes there.
package theme

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Text is copy in several locales: {"ar": "...", "en": "..."}.
type Text map[string]string

// In returns the copy for locale, falling back to fallback, then to "".
func (t Text) In(locale, fallback string) string {
	if v := t[locale]; v != "" {
		return v
	}
	return t[fallback]
}

// ColorPair is one colour for each theme, as #RRGGBB.
type ColorPair struct {
	Light string `json:"light"`
	Dark  string `json:"dark"`
}

var hexRE = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// IsHex reports whether v is a #RRGGBB colour.
func IsHex(v string) bool { return hexRE.MatchString(v) }

// Hex is v as an upper-case #RRGGBB colour, or "" when it is not one.
func Hex(v string) string {
	v = strings.TrimSpace(v)
	if !hexRE.MatchString(v) {
		return ""
	}
	return strings.ToUpper(v)
}

// Contrast floors.
const (
	// MinActionContrast is the hard floor for text on the action colour: WCAG's
	// 3:1 for large text and UI components.
	MinActionContrast = 3.0
	// MinInkContrast is the floor for a theme's ink on its ground: WCAG AA for body text.
	MinInkContrast = 4.5
)

// Contrast is the WCAG 2 contrast ratio between two #RRGGBB colours.
func Contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func luminance(hex string) float64 {
	if !hexRE.MatchString(hex) {
		return 0
	}
	ch := func(i int) float64 {
		v, _ := strconv.ParseUint(hex[i:i+2], 16, 8)
		c := float64(v) / 255
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(1) + 0.7152*ch(3) + 0.0722*ch(5)
}

// OnActionCandidates are the text colours tried on an action colour; the one
// with the better contrast wins. An app may set its own before it maps a theme.
var OnActionCandidates = []string{"#FFFFFF", "#0E1A3C"}

// OnActionFor picks the text colour for an action colour: whichever of the
// candidates contrasts more. action must be #RRGGBB.
func OnActionFor(action string) string {
	best, bestRatio := OnActionCandidates[0], 0.0
	for _, c := range OnActionCandidates {
		if r := Contrast(action, c); r > bestRatio {
			best, bestRatio = c, r
		}
	}
	return best
}
