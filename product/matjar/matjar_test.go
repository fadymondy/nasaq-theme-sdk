package matjar_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/fadymondy/nasaq-theme-sdk/product"
	"github.com/fadymondy/nasaq-theme-sdk/product/matjar"
)

func TestRegistered(t *testing.T) {
	a, err := product.Get("matjar")
	if err != nil {
		t.Fatal(err)
	}
	p := a.Product()
	for _, k := range []string{"home", "collection", "product", "cart", "checkout", "account"} {
		if _, ok := p.Template(k); !ok {
			t.Errorf("template %s missing", k)
		}
	}
	if len(a.Packs()) == 0 || a.Packs()[0].Key != "sahar" {
		t.Fatalf("packs = %+v", a.Packs())
	}
}

func TestSaharPack(t *testing.T) {
	p, err := matjar.LoadPack("sahar")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Products) != 13 || len(p.Categories) != 4 || len(p.Collections) < 3 {
		t.Fatalf("pack has %d products, %d categories, %d collections", len(p.Products), len(p.Categories), len(p.Collections))
	}
	for _, x := range p.Products {
		if x.Price() <= 0 || x.Title("ar") == "" || x.Title("en") == "" {
			t.Errorf("product %s: price %v, titles %q / %q", x.Key, x.Price(), x.Title("en"), x.Title("ar"))
		}
	}
	if _, err := matjar.LoadPack("../sahar"); !errors.Is(err, product.ErrUnknownPack) {
		t.Fatalf("path pack: %v", err)
	}
}

func TestPreview(t *testing.T) {
	a := matjar.Adapter{}
	for _, tc := range []struct {
		tpl    string
		params map[string]string
		key    string
	}{
		{"home", nil, "collections"},
		{"collection", nil, "products"},
		{"collection", map[string]string{"slug": "gifts"}, "collection"},
		{"collection", map[string]string{"cat": "face"}, "products"},
		{"product", map[string]string{"slug": "night-oil"}, "related"},
		{"cart", nil, "totals"},
		{"checkout", nil, "payment"},
		{"account", nil, "orders"},
	} {
		v, err := a.Preview(tc.tpl, "sahar", tc.params)
		if err != nil {
			t.Fatalf("%s %v: %v", tc.tpl, tc.params, err)
		}
		if v[tc.key] == nil {
			t.Errorf("%s %v: no %s", tc.tpl, tc.params, tc.key)
		}
		if _, err := json.Marshal(v); err != nil {
			t.Errorf("%s: %v", tc.tpl, err)
		}
	}
	v, _ := a.Preview("collection", "sahar", map[string]string{"cat": "face"})
	if n := len(v["products"].([]map[string]any)); n != 2 {
		t.Errorf("face has %d products, want 2", n)
	}
	if _, err := a.Preview("product", "sahar", map[string]string{"slug": "nope"}); !errors.Is(err, product.ErrNotFound) {
		t.Errorf("unknown product: %v", err)
	}
	if _, err := a.Preview("blog", "sahar", nil); !errors.Is(err, product.ErrUnknownTemplate) {
		t.Errorf("unknown template: %v", err)
	}
}

func TestCartTotals(t *testing.T) {
	p, _ := matjar.LoadPack("sahar")
	// night-oil 189 + candle-trio 245 + 2 x milk-soap 36 = 506; 10% off = 50.60; free shipping from 250
	got := p.CartTotals("")
	if got.Subtotal != 506 || got.Discount != 50.6 || got.Shipping != 0 || got.Total != 455.4 {
		t.Fatalf("totals = %+v", got)
	}
	// VAT is the included share: 455.40 * 15 / 115
	if got.VAT != 59.4 {
		t.Fatalf("vat = %v", got.VAT)
	}
}

func TestBrandOf(t *testing.T) {
	var p product.Package
	p.Schema, p.Product, p.Theme, p.Version = 1, "matjar", "store", "3"
	p.Tokens.Colors.Light = map[string]string{"primary": "#2c4a38", "accent": "#d9a441", "background": "#fbfaf7", "foreground": "#1b1f1a"}
	p.Tokens.Radius = "sm"
	p.Tokens.Fonts = map[string]string{"latin": "Inter"}
	p.Content.Hero.Heading = product.Text{"en": "Slow botanicals", "ar": "عناية هادئة"}
	b := matjar.BrandOf(p)
	if b.Primary != "#2C4A38" || b.Accent != "#D9A441" || b.Radius != "sharp" || b.Font != "inter" || b.Mode != "light" {
		t.Fatalf("brand = %+v", b)
	}
	if b.Tokens["--nq-bg"] != "#FBFAF7" || b.Hero.HeadingAR != "عناية هادئة" || b.Preset != "nasaq:store@3" {
		t.Fatalf("brand = %+v", b)
	}
	// with a dark palette Matjar derives both modes itself
	p.Tokens.Colors.Dark = map[string]string{"primary": "#a8c9b1"}
	if b := matjar.BrandOf(p); b.Mode != "system" || len(b.Tokens) != 0 {
		t.Fatalf("dark brand = %+v", b)
	}
}
