package theme

import (
	"errors"
	"math"
	"testing"
	"time"
)

func clinic() Tokens {
	var t Tokens
	t.Colors.Light = Palette{Primary: "#0F7C78", Accent: "#E8A87C", Background: "#FBFEFD", Foreground: "#10302F"}
	t.Colors.Dark = Palette{Primary: "#4FD1C5", Accent: "#E8A87C", Background: "#0B1F1E", Foreground: "#E6F4F2"}
	t.Radius, t.Density = "lg", "spacious"
	return t
}

func TestMapWritesTheTheme(t *testing.T) {
	cur := Look{Action: ColorPair{Light: "#123456", Dark: "#654321"}}
	got, err := Map(cur, Source{ID: "clinic", Version: "1.0.0"}, clinic(), time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got.Action.Light != "#0F7C78" || got.Action.Dark != "#4FD1C5" {
		t.Fatalf("action = %+v", got.Action)
	}
	if got.Theme.Source != SourceNasaq || got.Theme.ID != "clinic" || got.Theme.Radius != "lg" || got.Theme.Density != "spacious" {
		t.Fatalf("theme = %+v", got.Theme)
	}
	if got.Theme.Background.Light != "#FBFEFD" || got.Theme.Foreground.Dark != "#E6F4F2" {
		t.Fatalf("ground/ink = %+v %+v", got.Theme.Background, got.Theme.Foreground)
	}
	if got.Theme.OwnAction.Light != "#123456" {
		t.Fatalf("own action not kept: %+v", got.Theme.OwnAction)
	}
	if p := got.Theme.Problems(); len(p) > 0 {
		t.Fatalf("mapped theme has problems: %v", p)
	}
	// A second theme keeps the site's own colour, not the first theme's.
	again, err := Map(got, Source{ID: "other"}, clinic(), time.Unix(0, 0))
	if err != nil || again.Theme.OwnAction.Light != "#123456" {
		t.Fatalf("own action after a second theme = %+v (%v)", again.Theme.OwnAction, err)
	}
	// Default puts it back.
	back := Default(again, time.Unix(0, 0))
	if back.Action.Light != "#123456" || back.Action.Dark != "#654321" || back.Theme.Source != SourceDefault {
		t.Fatalf("default = %+v / %+v", back.Action, back.Theme)
	}
	if back.OnAction.Light != OnActionFor("#123456") {
		t.Fatalf("on action = %+v", back.OnAction)
	}
	if !Same(got, Look{Action: got.Action, OnAction: got.OnAction, Theme: func() SiteTheme { th := got.Theme; th.AppliedAt = "x"; return th }()}) {
		t.Fatal("applied_at alone made two looks differ")
	}
	if Same(got, back) {
		t.Fatal("different looks are the same")
	}
}

func TestMapDropsWhatWouldNotRender(t *testing.T) {
	tok := clinic()
	tok.Colors.Light.Foreground = "#F0F0F0" // unreadable on its ground
	tok.Colors.Dark.Accent = "teal"
	tok.Radius, tok.Density = "huge", "airy"
	got, err := Map(Look{}, Source{ID: "clinic"}, tok, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme.Background.Light != "" || got.Theme.Foreground.Light != "" {
		t.Fatalf("light ground kept: %+v", got.Theme)
	}
	if got.Theme.Background.Dark == "" {
		t.Fatal("dark ground dropped too")
	}
	if got.Theme.Accent.Dark != "#E8A87C" || got.Theme.Radius != "" || got.Theme.Density != "" {
		t.Fatalf("theme = %+v", got.Theme)
	}
	if p := got.Theme.Problems(); len(p) > 0 {
		t.Fatalf("mapped theme has problems: %v", p)
	}

	tok = clinic()
	tok.Colors.Light.Primary = ""
	if _, err := Map(Look{}, Source{ID: "x"}, tok, time.Now()); !errors.Is(err, ErrUnusable) {
		t.Fatalf("no primary: %v", err)
	}
}

func TestProblems(t *testing.T) {
	bad := SiteTheme{Source: "other", ID: "../x", Version: "v 1", ProjectID: "a b",
		Accent: ColorPair{Light: "red"}, Background: ColorPair{Light: "#FFFFFF"},
		Foreground: ColorPair{Dark: "#000000"}, Radius: "huge", Density: "airy"}
	if p := bad.Problems(); len(p) < 8 {
		t.Fatalf("problems = %v", p)
	}
	lowInk := SiteTheme{Background: ColorPair{Light: "#FFFFFF"}, Foreground: ColorPair{Light: "#EEEEEE"}}
	if p := lowInk.Problems(); len(p) != 1 {
		t.Fatalf("low ink = %v", p)
	}
	if p := (SiteTheme{}).Problems(); len(p) != 0 {
		t.Fatalf("zero theme = %v", p)
	}
}

func TestPublic(t *testing.T) {
	if (SiteTheme{}).PublicOrNil() != nil {
		t.Fatal("a theme never applied projects")
	}
	p := SiteTheme{Source: SourceNasaq, ID: "clinic", Radius: "lg", OwnAction: ColorPair{Light: "#123456"}}.PublicOrNil()
	if p == nil || p.Source != SourceNasaq || p.Radius != "lg" {
		t.Fatalf("public = %+v", p)
	}
	if (SiteTheme{ID: "x"}).Public().Source != SourceDefault {
		t.Fatal("empty source is not the default")
	}
}

func TestContrast(t *testing.T) {
	if r := Contrast("#FFFFFF", "#000000"); math.Abs(r-21) > 0.01 {
		t.Fatalf("white on black = %v", r)
	}
	if OnActionFor("#FFFF00") != "#0E1A3C" || OnActionFor("#0E1A3C") != "#FFFFFF" {
		t.Fatal("on-action pick wrong")
	}
	if Hex(" #abcdef ") != "#ABCDEF" || Hex("#abc") != "" {
		t.Fatal("hex")
	}
}

func TestEveryRefusalIsBilingual(t *testing.T) {
	for _, e := range All {
		if e.EN == "" || e.AR == "" || e.Status < 400 || e.In("ar") != e.AR || e.In("en") != e.EN {
			t.Errorf("%+v", e)
		}
	}
}
