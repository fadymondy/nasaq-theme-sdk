// Package matjar is Matjar's side of the Studio product contract: its bindable schema, its
// page templates, the seed packs a Matjar theme is designed with, the preview view models,
// and the mapping of a published theme package onto a Matjar store brand.
//
// Importing the package registers Matjar in product.Default.
package matjar

import (
	"github.com/fadymondy/nasaq-theme-sdk/product"
)

// ID is Matjar's product id.
const ID = "matjar"

type t = product.Text

func f(key, typ, en, ar string, opts ...func(*product.Field)) product.Field {
	fl := product.Field{Key: key, Type: typ, Label: t{"en": en, "ar": ar}}
	for _, o := range opts {
		o(&fl)
	}
	return fl
}

func required(fl *product.Field)  { fl.Required = true }
func localized(fl *product.Field) { fl.Localized = true }
func ref(e string) func(*product.Field) {
	return func(fl *product.Field) { fl.Ref = e }
}

// Def is Matjar's product definition. Product field keys are the keys Studio's store theme
// reads, so seeded entries render as they are; prices are in major units (SAR).
var Def = product.Product{
	ID:          ID,
	Name:        t{"en": "Matjar", "ar": "متجر"},
	Description: t{"en": "Online stores: catalogue, cart, checkout and customer accounts.", "ar": "المتاجر الإلكترونية: الكتالوج والسلة والدفع وحسابات العملاء."},
	Coverage:    ID,
	Entities: []product.Entity{
		{Key: "store", Label: t{"en": "Store", "ar": "المتجر"}, Fields: []product.Field{
			f("name", "text", "Name", "الاسم", required, localized),
			f("tagline", "text", "Tagline", "الشعار النصي", localized),
			f("currency", "text", "Currency", "العملة", required),
			f("email", "text", "Email", "البريد"),
		}},
		{Key: "category", Label: t{"en": "Category", "ar": "فئة"}, Fields: []product.Field{
			f("key", "text", "Key", "المعرف", required),
			f("name", "text", "Name", "الاسم", required, localized),
		}},
		{Key: "product", Label: t{"en": "Product", "ar": "منتج"}, Fields: []product.Field{
			f("title", "text", "Title", "العنوان", required, localized),
			f("excerpt", "text", "Summary", "الملخص", localized),
			f("body", "richtext", "Description", "الوصف", localized),
			f("price", "money", "Price", "السعر", required),
			f("compare_price", "money", "Compare-at price", "السعر قبل الخصم"),
			f("currency", "text", "Currency", "العملة"),
			f("category", "ref", "Category", "الفئة", ref("category")),
			f("badge", "text", "Badge", "الشارة", localized),
			f("rating", "number", "Rating", "التقييم"),
			f("reviews", "int", "Reviews", "المراجعات"),
			f("images", "images", "Images", "الصور", required),
			f("image_alts", "list", "Image descriptions", "وصف الصور", localized),
			f("size", "text", "Size", "الحجم"),
			f("sizes", "list", "Sizes", "الأحجام"),
			f("sku", "text", "SKU", "رمز المنتج"),
			f("stock", "int", "Stock", "المخزون"),
			f("details", "list", "Details", "التفاصيل", localized),
			f("assure", "list", "Assurances", "الضمانات", localized),
		}},
		{Key: "collection", Label: t{"en": "Collection", "ar": "مجموعة"}, Fields: []product.Field{
			f("title", "text", "Title", "العنوان", required, localized),
			f("description", "text", "Description", "الوصف", localized),
			f("products", "refs", "Products", "المنتجات", required, ref("product")),
		}},
		{Key: "page", Label: t{"en": "Page", "ar": "صفحة"}, Fields: []product.Field{
			f("title", "text", "Title", "العنوان", required, localized),
			f("body", "richtext", "Body", "المحتوى", localized),
		}},
		{Key: "menu", Label: t{"en": "Menu", "ar": "قائمة"}, Fields: []product.Field{
			f("items", "list", "Items", "العناصر", required),
		}},
		{Key: "cart", Label: t{"en": "Cart", "ar": "السلة"}, Fields: []product.Field{
			f("lines", "list", "Lines", "البنود", required),
			f("coupon", "text", "Coupon", "القسيمة"),
		}},
		{Key: "order", Label: t{"en": "Order", "ar": "طلب"}, Fields: []product.Field{
			f("number", "text", "Number", "الرقم", required),
			f("date", "text", "Date", "التاريخ"),
			f("status", "text", "Status", "الحالة"),
			f("lines", "list", "Lines", "البنود", required),
		}},
		{Key: "customer", Label: t{"en": "Customer", "ar": "العميل"}, Fields: []product.Field{
			f("name", "text", "Name", "الاسم", required),
			f("email", "text", "Email", "البريد"),
		}},
	},
	Templates: []product.Template{
		{Key: "home", Route: "/", Label: t{"en": "Home", "ar": "الرئيسية"}, Binds: []string{"store", "collection", "product", "category"}},
		{Key: "collection", Route: "/shop", Label: t{"en": "Collection", "ar": "المجموعة"}, Binds: []string{"collection", "product", "category"}},
		{Key: "product", Route: "/shop/{slug}", Label: t{"en": "Product", "ar": "المنتج"}, Binds: []string{"product", "category"}},
		{Key: "cart", Route: "/cart", Label: t{"en": "Cart", "ar": "السلة"}, Binds: []string{"cart", "product"}},
		{Key: "checkout", Route: "/checkout", Label: t{"en": "Checkout", "ar": "الدفع"}, Binds: []string{"cart", "customer"}},
		{Key: "account", Route: "/account", Label: t{"en": "Account", "ar": "الحساب"}, Binds: []string{"customer", "order"}},
	},
}

// Adapter is Matjar's product.Adapter.
type Adapter struct{}

func (Adapter) Product() product.Product { return Def }

func init() { product.Register(Adapter{}) }
