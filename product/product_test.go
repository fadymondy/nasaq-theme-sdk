package product_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fadymondy/nasaq-theme-sdk/product"
)

func pkg() product.Package {
	var p product.Package
	p.Schema, p.Product, p.Theme, p.Version = 1, "matjar", "store", "1"
	p.Tokens.Colors.Light = map[string]string{"primary": "#2c4a38"}
	return p
}

func TestValidate(t *testing.T) {
	if err := pkg().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mut := range map[string]func(*product.Package){
		"schema":  func(p *product.Package) { p.Schema = 2 },
		"product": func(p *product.Package) { p.Product = "" },
		"theme":   func(p *product.Package) { p.Theme = "../x" },
		"primary": func(p *product.Package) { p.Tokens.Colors.Light["primary"] = "red" },
		"radius":  func(p *product.Package) { p.Tokens.Radius = "huge" },
		"css":     func(p *product.Package) { p.CSS = "a{}</style><script>" },
	} {
		p := pkg()
		p.Tokens.Colors.Light = map[string]string{"primary": "#2c4a38"}
		mut(&p)
		if p.Validate() == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestReceiverAndPush(t *testing.T) {
	hash := product.HashToken("tok-1")
	var got product.Package
	rc := product.Receiver{
		Product: "matjar",
		Verify: func(_ context.Context, _ *http.Request, tok string) (string, error) {
			if !product.TokenMatches(tok, hash) {
				return "", product.ErrBadToken
			}
			return "tenant-a", nil
		},
		Apply: func(_ context.Context, tenant string, p product.Package) (product.Applied, error) {
			if tenant != "tenant-a" {
				t.Errorf("tenant = %s", tenant)
			}
			got = p
			return product.Applied{Theme: p.Theme, Version: p.Version}, nil
		},
	}
	mux := http.NewServeMux()
	mux.Handle(product.ThemePath, rc)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	a, err := product.Push(context.Background(), nil, srv.URL, "tok-1", pkg())
	if err != nil || a.Theme != "store" || got.Theme != "store" {
		t.Fatalf("push: %+v %v", a, err)
	}
	var pe *product.PushError
	if _, err := product.Push(context.Background(), nil, srv.URL, "wrong", pkg()); !errors.As(err, &pe) || pe.Status != 401 {
		t.Fatalf("bad token: %v", err)
	}
	other := pkg()
	other.Product = "qaima"
	if _, err := product.Push(context.Background(), nil, srv.URL, "tok-1", other); !errors.As(err, &pe) || pe.Status != 422 {
		t.Fatalf("other product: %v", err)
	}
	empty := httptest.NewServer(http.NotFoundHandler())
	defer empty.Close()
	if _, err := product.Push(context.Background(), nil, empty.URL, "tok-1", pkg()); !errors.Is(err, product.ErrNotLinked) {
		t.Fatalf("no receiver: %v", err)
	}
	if !strings.HasPrefix(product.ThemePath, "/api/") {
		t.Fatal(product.ThemePath)
	}
}
