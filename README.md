<div align="center">

<img src="https://raw.githubusercontent.com/fadymondy/nasaq/main/packages/brands/assets/nasaq-mark.svg" alt="Nasaq" width="80" />

# nasaq-theme-sdk

**Nasaq themes for any app: the theme catalog, live themes from Nasaq Studio, and a ready-made Themes page (Go + npm).**

[![CI](https://github.com/fadymondy/nasaq-theme-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/fadymondy/nasaq-theme-sdk/actions/workflows/ci.yml)
[![npm](https://img.shields.io/npm/v/@fadymondy/nasaq-theme-sdk.svg)](https://www.npmjs.com/package/@fadymondy/nasaq-theme-sdk)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/fadymondy/nasaq-theme-sdk?style=social)](https://github.com/fadymondy/nasaq-theme-sdk/stargazers)

[Nasaq](https://github.com/fadymondy/nasaq) · [Go docs](https://pkg.go.dev/github.com/fadymondy/nasaq-theme-sdk) · [Discussions](https://github.com/fadymondy/nasaq-theme-sdk/discussions) · [Contributing](CONTRIBUTING.md)

</div>

Nasaq themes for any app. An app built on it gets:

- **Theme switching.** The app's businesses (sites) pick a theme from the Nasaq catalog.
- **Live themes from Nasaq Studio.** A site connects its CircleXO organisation, chooses a Nasaq
  project, and from then on follows every theme that project publishes.
- **A ready-made Themes page,** built from stock Nasaq components.

SeatFor uses it, and Matjar is next.

It has two halves that share one contract:

| Half | Module | What it holds |
|---|---|---|
| Go | `github.com/fadymondy/nasaq-theme-sdk` | Theme model and mapping (`theme`), Nasaq client (`nasaq`), CircleXO connector with storage (`circlexo`), fake Nasaq for tests (`nasaqtest`) |
| npm | `@fadymondy/nasaq-theme-sdk` | Types, `themeCSS`, contrast helpers, API client; `/react` adds the Themes page and the CircleXO return |

## Go

```go
import (
	"github.com/fadymondy/nasaq-theme-sdk/circlexo"
	"github.com/fadymondy/nasaq-theme-sdk/nasaq"
	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

catalog := nasaq.New("matjar", nasaq.Static("https://ui.nasaq.app")) // the app's slug in Nasaq coverage

conn := &circlexo.Connector{
	Hub:           func(ctx context.Context) (cxo.Config, error) { return hubConfig(ctx) }, // the app's CircleXO client
	NasaqAudience: func(context.Context) string { return "nasaq" },
	Nasaq:         catalog,
	Store:         &circlexo.SQLStore{DB: pool},   // or your own circlexo.Store
	Vault:         vault,                          // seals the refresh token at rest
	Site:          sites,                          // reads and writes a site's theme.Look
}

conn.StartSweep(ctx, log, 5*time.Minute, func(site string) { /* flush the site's caches */ })
```

### The pieces

- **`theme`** holds the model.
  - Types: `SiteTheme`, `Look` (action colour, its label colour, theme) and `ColorPair`.
  - `Map` turns a Nasaq theme's tokens into a look. It drops any colour that would not render
    readably.
  - `Default` puts the site's own colour back.
  - `Problems` validates a stored theme.
  - Every refusal is a `*theme.Error` with an HTTP status and both English and Arabic text.
- **`nasaq.Client`** reads the public catalog. The catalog is cached per URL and still answers
  from cache while Nasaq is down. The client also calls Nasaq's app API: projects, bind and the
  published core theme. `Cards` builds the catalog that the Themes page shows.
- **`circlexo.Connector`** handles the CircleXO side:
  - It connects a site to its CircleXO organisation. The OIDC state is sealed, bound to both the
    site and the user, and lasts 15 minutes.
  - It exchanges the site's refresh token for a Nasaq token.
  - It binds a project and pulls the project's published theme.
  - A sweep pulls every bound site. When a grant is lost, the site is marked `grant_lost` and
    the sweep stops visiting it.
  - `Status` builds the Themes page's `ThemeStatus`.
- **`circlexo.SQLStore`** is the Postgres store.
  - `circlexo.Schema` creates its table.
  - You can point it at your own table with `Table`, `Key` and `KeyType`, for example
    `Table: "tenant_circlexo_links", Key: "tenant_id"`.
  - `MemStore` is an in-memory store for tests.

### The HTTP contract

The app serves these routes under one base, for example `/api/tenant/theme`. The npm client
calls them.

| Method | Path | Answer |
|---|---|---|
| GET | `` | `ThemeStatus` |
| GET | `/catalog?locale=` | `{ items: ThemeCard[] }` |
| POST | `/use` `{id}` | `ThemeStatus` |
| POST | `/reset` | `ThemeStatus` |
| POST | `/circlexo` `{action:"start", redirect_uri}` | `ThemeConnectView` (`auth_url`) |
| POST | `/circlexo` `{action:"finish", code, state}` | `ThemeConnectView` |
| DELETE | `/circlexo` | `ThemeStatus` |
| GET | `/projects` | `{ items: Project[] }` |
| POST | `/bind` `{project_id}` | `ThemeStatus` |
| POST | `/pull` | `ThemeStatus` |

Errors are JSON with a `message`. Use `(*theme.Error).In(locale)` for the text in the caller's
language.

## npm

```bash
npm i @fadymondy/nasaq-theme-sdk @fadymondy/nasaq
```

```ts
import { createThemesClient, themeStyleSheet } from "@fadymondy/nasaq-theme-sdk"

// Public pages: the site's theme as Nasaq variables.
const css = themeStyleSheet(site.theme) // ":root{...}.dark{...}"

// The admin: the themes API over your own fetch.
const client = createThemesClient((path, init) => api(`/tenant/theme${path}`, init))
```

```tsx
import { ThemesPanel, CircleXOReturn, themeLabels } from "@fadymondy/nasaq-theme-sdk/react"

<ThemesPanel client={client} locale="ar" appName="Matjar" redirectUri={() => `${location.origin}/admin/themes/circlexo`} />

// On the redirect page:
<CircleXOReturn client={client} locale="ar" appName="Matjar" params={searchParams} onDone={() => router.replace("/admin/themes")} />
```

The components use `@fadymondy/nasaq/web` and Tailwind utilities. Add the package to Tailwind's
sources next to Nasaq's:

```css
@source "../node_modules/@fadymondy/nasaq-theme-sdk/dist";
```

The copy is built in, in Arabic and English. `{app}` becomes `appName`. To change any string,
pass `labels`.

## Products (Nasaq Studio themes)

Nasaq Studio builds themes *for* a product. The `product` package is that contract:

- `product.Product`: the product id, the entities and fields a theme can bind to, and the page
  templates every theme must render (Matjar: home, collection, product, cart, checkout, account).
- Seed packs: realistic data a theme is designed with (`matjar.LoadPack("sahar")`: 13 products,
  4 categories, 3 collections, pages, menus, a cart, checkout and account sample). Keys are stable,
  so seeding is idempotent and a reset removes exactly what a pack added.
- `Adapter.Preview(template, pack, params)`: the view model of a template filled with a pack.
- `product.Package`: the published theme Studio sends with `PUT /api/nasaq/theme`.
- `product.Receiver`: the product's handler for that route (bearer link token, validation,
  bilingual errors); `product.Push` is the Studio side.

```go
import (
	"github.com/fadymondy/nasaq-theme-sdk/product"
	"github.com/fadymondy/nasaq-theme-sdk/product/matjar" // registers Matjar
)

mux.Handle(product.ThemePath, product.Receiver{
	Product: matjar.ID,
	Verify:  func(ctx context.Context, r *http.Request, tok string) (string, error) { return store.TenantForLinkToken(ctx, tok) },
	Apply: func(ctx context.Context, tenant string, p product.Package) (product.Applied, error) {
		return store.ApplyBrand(ctx, tenant, matjar.BrandOf(p)) // colours, corners, font, hero
	},
})
```

npm: `@fadymondy/nasaq-theme-sdk/products` has the same packs (generated from the Go JSON by
`npm run gen`), the templates, the `ThemePackage` type and `matjarLook`.

## Develop

```bash
go test ./...                    # Go
npm install && npm run build && npm test   # npm
```

## Contributing

Contributions are welcome: bug reports, theme mapping fixes, and new product packs. Read [CONTRIBUTING.md](CONTRIBUTING.md) for setup, conventions
and the DCO sign-off, then look for
[`good first issue`](https://github.com/fadymondy/nasaq-theme-sdk/labels/good%20first%20issue).

## Community

- [GitHub Discussions](https://github.com/fadymondy/nasaq-theme-sdk/discussions): questions, ideas, show and tell
- [Issues](https://github.com/fadymondy/nasaq-theme-sdk/issues/new/choose): bugs and feature requests
- Security reports: see [SECURITY.md](SECURITY.md)

Please follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

MIT, © 2026 Fady Mondy. See [LICENSE](LICENSE).
