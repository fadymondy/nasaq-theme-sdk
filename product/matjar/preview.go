package matjar

import (
	"math"

	"github.com/fadymondy/nasaq-theme-sdk/product"
)

// Preview view models. Money is in major units with the pack currency; totals are computed
// in minor units (halalas) so they add up exactly.

func minor(v float64) int64        { return int64(math.Round(v * 100)) }
func major(v int64) float64        { return float64(v) / 100 }
func pct(v int64, p float64) int64 { return int64(math.Round(float64(v) * p / 100)) }

func card(x Product) map[string]any {
	m := map[string]any{"key": x.Key, "slug": x.Slug}
	for _, k := range []string{"title", "excerpt", "price", "compare_price", "currency", "category", "badge", "rating", "reviews", "images", "image_alts", "size", "stock"} {
		if v, ok := x.Fields[k]; ok {
			m[k] = v
		}
	}
	return m
}

func (p *Pack) cards(keys []string) []map[string]any {
	out := []map[string]any{}
	for _, k := range keys {
		if x, ok := p.Item(k); ok {
			out = append(out, card(x))
		}
	}
	return out
}

func (p *Pack) productKeys(cat string) []string {
	var ks []string
	for _, x := range p.Products {
		if cat == "" || x.Category() == cat {
			ks = append(ks, x.Key)
		}
	}
	return ks
}

// Totals are the money lines of a cart.
type Totals struct {
	Subtotal float64 `json:"subtotal"`
	Discount float64 `json:"discount"`
	Shipping float64 `json:"shipping"`
	VAT      float64 `json:"vat"`
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
}

func (p *Pack) lines(ls []Line) ([]map[string]any, int64) {
	out := []map[string]any{}
	var sub int64
	for _, l := range ls {
		x, ok := p.Item(l.Product)
		if !ok {
			continue
		}
		unit := minor(x.Price())
		sub += unit * int64(l.Qty)
		out = append(out, map[string]any{"product": card(x), "size": l.Size, "qty": l.Qty, "unit": major(unit), "total": major(unit * int64(l.Qty))})
	}
	return out, sub
}

// CartTotals prices the pack's cart with a shipping option ("" = the first). Prices include
// VAT (Saudi retail), so VAT is the included share, not an addition.
func (p *Pack) CartTotals(shipping string) Totals {
	_, sub := p.lines(p.Cart.Lines)
	var disc int64
	if p.Cart.Coupon != nil {
		disc = pct(sub, p.Cart.Coupon.Percent)
	}
	var ship int64
	for i, o := range p.Checkout.Shipping {
		if o.Key == shipping || (shipping == "" && i == 0) {
			ship = minor(o.Price)
			if o.FreeFrom > 0 && sub-disc >= minor(o.FreeFrom) {
				ship = 0
			}
			break
		}
	}
	total := sub - disc + ship
	var vat int64
	if p.Checkout.VATPercent > 0 {
		vat = int64(math.Round(float64(total) * p.Checkout.VATPercent / (100 + p.Checkout.VATPercent)))
	}
	return Totals{Subtotal: major(sub), Discount: major(disc), Shipping: major(ship), VAT: major(vat), Total: major(total), Currency: p.Cart.Currency}
}

func (Adapter) Preview(template, pack string, params map[string]string) (map[string]any, error) {
	p, err := LoadPack(pack)
	if err != nil {
		return nil, err
	}
	if _, ok := Def.Template(template); !ok {
		return nil, product.ErrUnknownTemplate
	}
	base := map[string]any{"template": template, "store": p.Store, "categories": p.Categories, "menus": p.Menus}
	switch template {
	case "home":
		var featured []map[string]any
		for _, c := range p.Collections {
			featured = append(featured, map[string]any{"key": c.Key, "slug": c.Slug, "title": c.Title, "description": c.Description, "products": p.cards(c.Products)})
		}
		base["hero"] = p.Hero
		base["collections"] = featured
		base["products"] = p.cards(p.productKeys(""))
	case "collection":
		if slug := params["slug"]; slug != "" {
			c, ok := p.FindCollection(slug)
			if !ok {
				return nil, product.ErrNotFound
			}
			base["collection"] = map[string]any{"key": c.Key, "slug": c.Slug, "title": c.Title, "description": c.Description}
			base["products"] = p.cards(c.Products)
		} else {
			cat := params["cat"]
			base["category"] = cat
			base["products"] = p.cards(p.productKeys(cat))
		}
	case "product":
		var x Product
		ok := false
		for _, y := range p.Products {
			if y.Key == params["slug"] || y.Slug["en"] == params["slug"] || y.Slug["ar"] == params["slug"] {
				x, ok = y, true
				break
			}
		}
		if !ok {
			return nil, product.ErrNotFound
		}
		base["product"] = map[string]any{"key": x.Key, "slug": x.Slug, "fields": x.Fields}
		var rel []string
		for _, k := range p.productKeys(x.Category()) {
			if k != x.Key && len(rel) < 4 {
				rel = append(rel, k)
			}
		}
		base["related"] = p.cards(rel)
	case "cart":
		ls, _ := p.lines(p.Cart.Lines)
		base["lines"] = ls
		base["coupon"] = p.Cart.Coupon
		base["totals"] = p.CartTotals(params["shipping"])
	case "checkout":
		ls, _ := p.lines(p.Cart.Lines)
		base["lines"] = ls
		base["shipping"] = p.Checkout.Shipping
		base["payment"] = p.Checkout.Payment
		base["address"] = p.Checkout.Address
		base["totals"] = p.CartTotals(params["shipping"])
	case "account":
		var orders []map[string]any
		for _, o := range p.Account.Orders {
			ls, sub := p.lines(o.Lines)
			orders = append(orders, map[string]any{"number": o.Number, "date": o.Date, "status": o.Status, "lines": ls, "total": major(sub)})
		}
		base["customer"] = p.Account.Customer
		base["orders"] = orders
	}
	return base, nil
}
