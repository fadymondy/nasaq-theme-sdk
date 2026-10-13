// Package product is the contract between Nasaq Studio and the products a Studio theme can
// be built for (Matjar first). A product declares the data a theme can bind to (entities
// and their fields), the page templates every theme for it must cover, and the seed packs
// that give a theme realistic content while it is designed. An Adapter turns a pack into
// the view model of one template, so Studio can preview a template with product data, and
// the product's own storefront renders the same shape.
//
// The published theme travels from Studio to the product as a Package (package.go); the
// product receives it with a Receiver (receiver.go).
package product

import (
	"encoding/json"
	"errors"
	"sort"
	"sync"

	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// Text is a bilingual string keyed by locale ("en", "ar").
type Text = theme.Text

// Field is one bindable field of an entity. Type is one of text, richtext, money, number,
// int, bool, image, images, ref, refs, list.
type Field struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Label    Text   `json:"label"`
	Required bool   `json:"required,omitempty"`
	// Localized fields hold one value per locale ({"en": ..., "ar": ...}).
	Localized bool `json:"localized,omitempty"`
	// Ref is the entity a ref/refs field points to.
	Ref string `json:"ref,omitempty"`
}

// Entity is a kind of product data a theme can show (product, category, collection...).
type Entity struct {
	Key    string  `json:"key"`
	Label  Text    `json:"label"`
	Fields []Field `json:"fields"`
}

// Field returns the entity field with key k.
func (e Entity) Field(k string) (Field, bool) {
	for _, f := range e.Fields {
		if f.Key == k {
			return f, true
		}
	}
	return Field{}, false
}

// Template is a page every theme for the product must render. Route is the storefront path
// with {param} placeholders; Binds lists the entities the template's view model carries.
type Template struct {
	Key   string   `json:"key"`
	Route string   `json:"route"`
	Label Text     `json:"label"`
	Binds []string `json:"binds,omitempty"`
}

// Product is a product Studio can build themes for.
type Product struct {
	ID          string     `json:"id"`
	Name        Text       `json:"name"`
	Description Text       `json:"description,omitempty"`
	Entities    []Entity   `json:"entities"`
	Templates   []Template `json:"templates"`
	// Coverage is the manifest key a theme sets to say it supports the product
	// (manifest.coverage[Coverage] = "full" | "partial").
	Coverage string `json:"coverage"`
}

// Entity returns the product's entity with key k.
func (p Product) Entity(k string) (Entity, bool) {
	for _, e := range p.Entities {
		if e.Key == k {
			return e, true
		}
	}
	return Entity{}, false
}

// Template returns the product's template with key k.
func (p Product) Template(k string) (Template, bool) {
	for _, t := range p.Templates {
		if t.Key == k {
			return t, true
		}
	}
	return Template{}, false
}

// PackInfo describes a seed pack.
type PackInfo struct {
	Product     string `json:"product"`
	Key         string `json:"key"`
	Version     string `json:"version"`
	Name        Text   `json:"name"`
	Description Text   `json:"description,omitempty"`
}

// Adapter is what a product plugs into the registry.
type Adapter interface {
	Product() Product
	Packs() []PackInfo
	// Pack is the raw JSON of a seed pack (the product package documents its shape).
	Pack(key string) (json.RawMessage, error)
	// Preview is the view model of a template rendered with a pack's data. Params are the
	// template's route params ({slug}) and query options (cat, locale).
	Preview(template, pack string, params map[string]string) (map[string]any, error)
}

var (
	ErrUnknownProduct  = errors.New("product: unknown product")
	ErrUnknownPack     = errors.New("product: unknown seed pack")
	ErrUnknownTemplate = errors.New("product: unknown template")
	ErrNotFound        = errors.New("product: not found in pack")
)

// Registry holds the products Studio can target.
type Registry struct {
	mu sync.RWMutex
	m  map[string]Adapter
}

// Default is the process registry; product packages register themselves in it on import.
var Default = &Registry{}

// Register adds (or replaces) a product adapter.
func (r *Registry) Register(a Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.m == nil {
		r.m = map[string]Adapter{}
	}
	r.m[a.Product().ID] = a
}

// Get returns the adapter of product id.
func (r *Registry) Get(id string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if a, ok := r.m[id]; ok {
		return a, nil
	}
	return nil, ErrUnknownProduct
}

// All lists the registered products by id.
func (r *Registry) All() []Product {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Product, 0, len(r.m))
	for _, a := range r.m {
		out = append(out, a.Product())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Register adds a product to the default registry.
func Register(a Adapter) { Default.Register(a) }

// Get returns a product adapter from the default registry.
func Get(id string) (Adapter, error) { return Default.Get(id) }

// All lists the products of the default registry.
func All() []Product { return Default.All() }
