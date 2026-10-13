package matjar

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/fadymondy/nasaq-theme-sdk/product"
)

//go:embed packs/*.json
var packFS embed.FS

// Pack is a Matjar seed pack: a realistic store a theme is designed with. Keys are stable,
// so seeding the same pack twice updates instead of duplicating, and a reset removes
// exactly what the pack added.
type Pack struct {
	Schema      int          `json:"schema"`
	Product     string       `json:"product"`
	Key         string       `json:"key"`
	Version     string       `json:"version"`
	Name        t            `json:"name"`
	Description t            `json:"description"`
	Store       Store        `json:"store"`
	Hero        product.Hero `json:"hero"`
	Categories  []Category   `json:"categories"`
	Products    []Product    `json:"products"`
	Collections []Collection `json:"collections"`
	Pages       []Page       `json:"pages"`
	Menus       []Menu       `json:"menus"`
	Cart        Cart         `json:"cart"`
	Checkout    Checkout     `json:"checkout"`
	Account     Account      `json:"account"`
}

type Store struct {
	Name     t      `json:"name"`
	Tagline  t      `json:"tagline"`
	Currency string `json:"currency"`
	Email    string `json:"email"`
}

type Category struct {
	Key  string `json:"key"`
	Name t      `json:"name"`
}

// Product keeps its fields as the store theme reads them (see Def's product entity).
type Product struct {
	Key    string         `json:"key"`
	Slug   t              `json:"slug"`
	Fields map[string]any `json:"fields"`
}

// Price is the product price in major units.
func (p Product) Price() float64 { n, _ := p.Fields["price"].(float64); return n }

// Category is the product's category key.
func (p Product) Category() string { s, _ := p.Fields["category"].(string); return s }

// Title is the product title in a locale.
func (p Product) Title(locale string) string { return text(p.Fields["title"]).In(locale, "en") }

type Collection struct {
	Key         string   `json:"key"`
	Slug        t        `json:"slug"`
	Title       t        `json:"title"`
	Description t        `json:"description"`
	Products    []string `json:"products"`
}

type Page struct {
	Key   string `json:"key"`
	Slug  t      `json:"slug"`
	Title t      `json:"title"`
	Body  t      `json:"body"`
}

// Link is where a menu item goes: a template (with an optional collection), a page, or a URL.
type Link struct {
	Template   string `json:"template,omitempty"`
	Collection string `json:"collection,omitempty"`
	Page       string `json:"page,omitempty"`
	URL        string `json:"url,omitempty"`
}

type MenuItem struct {
	Label t    `json:"label"`
	To    Link `json:"to"`
}

type Menu struct {
	Key   string     `json:"key"`
	Items []MenuItem `json:"items"`
}

type Line struct {
	Product string `json:"product"`
	Size    string `json:"size,omitempty"`
	Qty     int    `json:"qty"`
}

type Cart struct {
	Currency string `json:"currency"`
	Lines    []Line `json:"lines"`
	Coupon   *struct {
		Code    string  `json:"code"`
		Percent float64 `json:"percent"`
	} `json:"coupon,omitempty"`
}

type Option struct {
	Key      string  `json:"key"`
	Name     t       `json:"name"`
	Price    float64 `json:"price,omitempty"`
	FreeFrom float64 `json:"freeFrom,omitempty"`
}

type Checkout struct {
	Shipping   []Option       `json:"shipping"`
	Payment    []Option       `json:"payment"`
	VATPercent float64        `json:"vatPercent"`
	Address    map[string]any `json:"address"`
}

type Order struct {
	Number string `json:"number"`
	Date   string `json:"date"`
	Status string `json:"status"`
	Lines  []Line `json:"lines"`
}

type Account struct {
	Customer struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"customer"`
	Orders []Order `json:"orders"`
}

// Item returns the pack product with key k.
func (p *Pack) Item(k string) (Product, bool) {
	for _, x := range p.Products {
		if x.Key == k {
			return x, true
		}
	}
	return Product{}, false
}

// FindCollection returns the pack collection with key or slug k.
func (p *Pack) FindCollection(k string) (Collection, bool) {
	for _, c := range p.Collections {
		if c.Key == k || c.Slug["en"] == k || c.Slug["ar"] == k {
			return c, true
		}
	}
	return Collection{}, false
}

// Validate checks that every reference in the pack resolves and keys are unique.
func (p *Pack) Validate() error {
	if p.Product != ID {
		return fmt.Errorf("matjar pack %s: product %q", p.Key, p.Product)
	}
	cats := map[string]bool{}
	for _, c := range p.Categories {
		if cats[c.Key] {
			return fmt.Errorf("matjar pack %s: duplicate category %s", p.Key, c.Key)
		}
		cats[c.Key] = true
	}
	prods := map[string]bool{}
	pe, _ := Def.Entity("product")
	for _, x := range p.Products {
		if prods[x.Key] {
			return fmt.Errorf("matjar pack %s: duplicate product %s", p.Key, x.Key)
		}
		prods[x.Key] = true
		for _, fl := range pe.Fields {
			if _, ok := x.Fields[fl.Key]; fl.Required && !ok {
				return fmt.Errorf("matjar pack %s: product %s has no %s", p.Key, x.Key, fl.Key)
			}
		}
		if !cats[x.Category()] {
			return fmt.Errorf("matjar pack %s: product %s in unknown category %q", p.Key, x.Key, x.Category())
		}
	}
	lines := func(where string, ls []Line) error {
		for _, l := range ls {
			if !prods[l.Product] || l.Qty < 1 {
				return fmt.Errorf("matjar pack %s: %s line %+v", p.Key, where, l)
			}
		}
		return nil
	}
	cols := map[string]bool{}
	for _, c := range p.Collections {
		if cols[c.Key] {
			return fmt.Errorf("matjar pack %s: duplicate collection %s", p.Key, c.Key)
		}
		cols[c.Key] = true
		seen := map[string]bool{}
		for _, k := range c.Products {
			if !prods[k] || seen[k] {
				return fmt.Errorf("matjar pack %s: collection %s product %q", p.Key, c.Key, k)
			}
			seen[k] = true
		}
	}
	pages := map[string]bool{}
	for _, pg := range p.Pages {
		pages[pg.Key] = true
	}
	for _, m := range p.Menus {
		for _, it := range m.Items {
			switch {
			case it.To.Page != "" && !pages[it.To.Page]:
				return fmt.Errorf("matjar pack %s: menu %s links to unknown page %s", p.Key, m.Key, it.To.Page)
			case it.To.Collection != "" && !cols[it.To.Collection]:
				return fmt.Errorf("matjar pack %s: menu %s links to unknown collection %s", p.Key, m.Key, it.To.Collection)
			case it.To.Template != "":
				if _, ok := Def.Template(it.To.Template); !ok {
					return fmt.Errorf("matjar pack %s: menu %s links to unknown template %s", p.Key, m.Key, it.To.Template)
				}
			}
		}
	}
	if err := lines("cart", p.Cart.Lines); err != nil {
		return err
	}
	for _, o := range p.Account.Orders {
		if err := lines("order "+o.Number, o.Lines); err != nil {
			return err
		}
	}
	return nil
}

var (
	packMu    sync.Mutex
	packCache = map[string]*Pack{}
)

// LoadPack returns an embedded seed pack (validated, cached; callers must not modify it).
func LoadPack(key string) (*Pack, error) {
	packMu.Lock()
	defer packMu.Unlock()
	if p, ok := packCache[key]; ok {
		return p, nil
	}
	b, err := RawPack(key)
	if err != nil {
		return nil, err
	}
	var p Pack
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("matjar pack %s: %w", key, err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	packCache[key] = &p
	return &p, nil
}

// RawPack is the embedded JSON of a pack.
func RawPack(key string) ([]byte, error) {
	if key == "" || strings.ContainsAny(key, "/\\.") {
		return nil, product.ErrUnknownPack
	}
	b, err := packFS.ReadFile(path.Join("packs", key+".json"))
	if err != nil {
		return nil, product.ErrUnknownPack
	}
	return b, nil
}

// PackKeys lists the embedded packs.
func PackKeys() []string {
	es, _ := packFS.ReadDir("packs")
	var out []string
	for _, e := range es {
		if k, ok := strings.CutSuffix(e.Name(), ".json"); ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func (Adapter) Packs() []product.PackInfo {
	var out []product.PackInfo
	for _, k := range PackKeys() {
		if p, err := LoadPack(k); err == nil {
			out = append(out, product.PackInfo{Product: ID, Key: p.Key, Version: p.Version, Name: p.Name, Description: p.Description})
		}
	}
	return out
}

func (Adapter) Pack(key string) (json.RawMessage, error) {
	if _, err := LoadPack(key); err != nil {
		return nil, err
	}
	return RawPack(key)
}

func text(v any) t {
	switch x := v.(type) {
	case map[string]any:
		out := t{}
		for k, s := range x {
			if str, ok := s.(string); ok {
				out[k] = str
			}
		}
		return out
	case string:
		return t{"en": x}
	}
	return t{}
}
