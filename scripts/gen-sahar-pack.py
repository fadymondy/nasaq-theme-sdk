"""One-off: build the SDK's Matjar seed pack (product/matjar/packs/sahar.json) from the
store theme's sample catalogue in nasaq-studio, plus the pages, menus and the cart,
checkout and account samples a storefront needs."""
import json, sys

src = sys.argv[1]
out = sys.argv[2]
d = json.load(open(src, encoding="utf8"))

cats = [
    {"key": "face", "name": {"en": "Face", "ar": "الوجه"}},
    {"key": "body", "name": {"en": "Body", "ar": "الجسم"}},
    {"key": "hair", "name": {"en": "Hair and oils", "ar": "الشعر والزيوت"}},
    {"key": "home", "name": {"en": "Home and candles", "ar": "المنزل والشموع"}},
]

keep = ["title", "excerpt", "body", "price", "compare_price", "currency", "category", "badge", "rating",
        "reviews", "images", "image_alts", "size", "sizes", "sku", "stock", "details", "assure"]
products = []
for x in d["content"]:
    if x["type"] != "product":
        continue
    f = {k: x["fields"][k] for k in keep if k in x["fields"]}
    products.append({"key": x["slug"]["en"], "slug": x["slug"], "fields": f})

def has_badge(p, word):
    b = p["fields"].get("badge")
    return isinstance(b, dict) and word in b.get("en", "")

collections = [
    {"key": "bestsellers", "slug": {"en": "bestsellers", "ar": "bestsellers"},
     "title": {"en": "Bestsellers", "ar": "الأكثر مبيعا"},
     "description": {"en": "The oils, serums and candles our customers come back for.", "ar": "الزيوت والأمصال والشموع التي يعود إليها عملاؤنا."},
     "products": [p["key"] for p in products if p["fields"].get("rating", 0) >= 4.8 or has_badge(p, "Bestseller")][:8]},
    {"key": "new", "slug": {"en": "new-in", "ar": "new-in"},
     "title": {"en": "New in", "ar": "وصل حديثا"},
     "description": {"en": "Fresh from the workshop this season.", "ar": "جديد الورشة هذا الموسم."},
     "products": [p["key"] for p in products if has_badge(p, "New")] + ["scalp-oil", "pear-oil"]},
    {"key": "gifts", "slug": {"en": "gifts", "ar": "gifts"},
     "title": {"en": "Gifts", "ar": "هدايا"},
     "description": {"en": "Wrapped and ready: candles, soaps and small rituals.", "ar": "مغلفة وجاهزة: شموع وصابون وطقوس صغيرة."},
     "products": [p["key"] for p in products if p["fields"].get("category") == "home"] + ["milk-soap", "hand-cream"]},
]

pages = [
    {"key": "about", "slug": {"en": "about", "ar": "about"}, "title": {"en": "About us", "ar": "من نحن"},
     "body": {"en": "A small workshop in Riyadh pressing botanical oils and pouring candles by hand since 2019.",
              "ar": "ورشة صغيرة في الرياض تعصر الزيوت النباتية وتصب الشموع يدويا منذ 2019."}},
    {"key": "shipping", "slug": {"en": "shipping", "ar": "shipping"}, "title": {"en": "Delivery and returns", "ar": "التوصيل والإرجاع"},
     "body": {"en": "Orders leave Riyadh within two working days. Delivery is free from 250 SAR. Unopened items can be returned within 30 days.",
              "ar": "تغادر الطلبات الرياض خلال يومي عمل. التوصيل مجاني ابتداء من 250 ريال. يمكن إرجاع المنتجات غير المفتوحة خلال 30 يوما."}},
    {"key": "faq", "slug": {"en": "faq", "ar": "faq"}, "title": {"en": "Questions", "ar": "أسئلة شائعة"},
     "body": {"en": "Are the oils fragrance free? Yes. Do you ship across the Gulf? Yes, to every GCC country.",
              "ar": "هل الزيوت خالية من العطور؟ نعم. هل تشحنون إلى دول الخليج؟ نعم، إلى كل دول مجلس التعاون."}},
    {"key": "contact", "slug": {"en": "contact", "ar": "contact"}, "title": {"en": "Contact", "ar": "تواصل"},
     "body": {"en": "Write to hello@sahar-demo.example or message us on WhatsApp, Sunday to Thursday.",
              "ar": "راسلنا على hello@sahar-demo.example أو عبر واتساب من الأحد إلى الخميس."}},
]

menus = [
    {"key": "main", "items": [
        {"label": {"en": "Shop", "ar": "تسوق"}, "to": {"template": "collection"}},
        {"label": {"en": "Bestsellers", "ar": "الأكثر مبيعا"}, "to": {"template": "collection", "collection": "bestsellers"}},
        {"label": {"en": "New in", "ar": "وصل حديثا"}, "to": {"template": "collection", "collection": "new"}},
        {"label": {"en": "Gifts", "ar": "هدايا"}, "to": {"template": "collection", "collection": "gifts"}},
        {"label": {"en": "About", "ar": "من نحن"}, "to": {"page": "about"}},
    ]},
    {"key": "footer", "items": [
        {"label": {"en": "Delivery and returns", "ar": "التوصيل والإرجاع"}, "to": {"page": "shipping"}},
        {"label": {"en": "Questions", "ar": "أسئلة شائعة"}, "to": {"page": "faq"}},
        {"label": {"en": "Contact", "ar": "تواصل"}, "to": {"page": "contact"}},
        {"label": {"en": "My account", "ar": "حسابي"}, "to": {"template": "account"}},
    ]},
]

cart = {
    "currency": "SAR",
    "lines": [
        {"product": "night-oil", "size": "30 ml", "qty": 1},
        {"product": "candle-trio", "size": "", "qty": 1},
        {"product": "milk-soap", "size": "", "qty": 2},
    ],
    "coupon": {"code": "WELCOME10", "percent": 10},
}

checkout = {
    "shipping": [
        {"key": "standard", "name": {"en": "Standard delivery, 2 to 4 days", "ar": "توصيل عادي، من 2 إلى 4 أيام"}, "price": 25, "freeFrom": 250},
        {"key": "express", "name": {"en": "Express, next day in Riyadh", "ar": "سريع، اليوم التالي في الرياض"}, "price": 45},
    ],
    "payment": [
        {"key": "card", "name": {"en": "Card (mada, Visa, Mastercard)", "ar": "بطاقة (مدى، فيزا، ماستركارد)"}},
        {"key": "apple_pay", "name": {"en": "Apple Pay", "ar": "Apple Pay"}},
        {"key": "cod", "name": {"en": "Cash on delivery", "ar": "الدفع عند الاستلام"}},
    ],
    "vatPercent": 15,
    "address": {"name": "Noura Al Harbi", "line1": {"en": "12 Al Olaya Street", "ar": "12 شارع العليا"},
                "city": {"en": "Riyadh", "ar": "الرياض"}, "country": "SA", "phone": "+966 50 000 0000"},
}

account = {
    "customer": {"name": "Noura Al Harbi", "email": "noura@sahar-demo.example"},
    "orders": [
        {"number": "SH-1042", "date": "2026-09-28", "status": "delivered", "lines": [{"product": "radiance-serum", "qty": 1}, {"product": "hand-cream", "qty": 2}]},
        {"number": "SH-1077", "date": "2026-10-09", "status": "shipped", "lines": [{"product": "candle-gold", "qty": 1}]},
    ],
}

pack = {
    "schema": 1,
    "product": "matjar",
    "key": "sahar",
    "version": "1.0.0",
    "name": {"en": "Sahar botanicals", "ar": "سحر للعناية النباتية"},
    "description": {"en": "A botanical skincare and candle store from Riyadh: 13 products in 4 categories, 3 collections, pages, menus and a cart, checkout and account sample.",
                    "ar": "متجر عناية نباتية وشموع من الرياض: 13 منتجا في 4 فئات، و3 مجموعات، وصفحات وقوائم، وعينة للسلة والدفع والحساب."},
    "store": {"name": {"en": "Sahar", "ar": "سحر"}, "tagline": {"en": "Botanical skincare, body care and home candles from Riyadh.", "ar": "عناية نباتية بالبشرة والجسم وشموع منزلية من الرياض."},
              "currency": "SAR", "email": "hello@sahar-demo.example"},
    "hero": {"heading": {"en": "Slow botanicals for skin and home", "ar": "عناية نباتية هادئة للبشرة والمنزل"},
             "sub": {"en": "Cold-pressed oils, milk soaps and hand-poured candles, made in small batches in Riyadh.", "ar": "زيوت معصورة على البارد وصابون بالحليب وشموع مصبوبة يدويا، تصنع على دفعات صغيرة في الرياض."},
             "cta": {"en": "Shop the collection", "ar": "تسوق المجموعة"}, "image": "night-oil"},
    "categories": cats,
    "products": products,
    "collections": collections,
    "pages": pages,
    "menus": menus,
    "cart": cart,
    "checkout": checkout,
    "account": account,
}
with open(out, "w", encoding="utf8", newline="\n") as fh:
    json.dump(pack, fh, ensure_ascii=False, indent=2)
    fh.write("\n")
print("products", len(products), "collections", [len(c["products"]) for c in collections])
