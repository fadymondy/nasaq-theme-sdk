// The Studio product contract (mirror of the Go package product and product/matjar).
import { PACKS } from "./packs.js"

export type Text = Record<string, string>

export type ProductTemplate = { key: string; route: string; label: Text }

/** The page templates every theme for a product must render. */
export const TEMPLATES: Record<string, ProductTemplate[]> = {
  matjar: [
    { key: "home", route: "/", label: { en: "Home", ar: "الرئيسية" } },
    { key: "collection", route: "/shop", label: { en: "Collection", ar: "المجموعة" } },
    { key: "product", route: "/shop/{slug}", label: { en: "Product", ar: "المنتج" } },
    { key: "cart", route: "/cart", label: { en: "Cart", ar: "السلة" } },
    { key: "checkout", route: "/checkout", label: { en: "Checkout", ar: "الدفع" } },
    { key: "account", route: "/account", label: { en: "Account", ar: "الحساب" } },
  ],
}

export const PRODUCTS = [{ id: "matjar", name: { en: "Matjar", ar: "متجر" }, coverage: "matjar" }] as const

export type PackKey = keyof typeof PACKS

/** A seed pack by product and key ("matjar", "sahar"), or undefined. */
export function seedPack(product: string, key: string) {
  return (PACKS as Record<string, unknown>)[`${product}/${key}`] as (typeof PACKS)[PackKey] | undefined
}

/** The packs of a product. */
export function seedPacks(product: string) {
  return Object.values(PACKS).filter((p) => p.product === product)
}

/** The published theme as a product receives it (PUT /api/nasaq/theme). */
export type ThemePackage = {
  schema: 1
  product: string
  theme: string
  version: string
  locale?: string
  brand: { name?: Text; tagline?: Text; logoLight?: string; logoDark?: string; favicon?: string }
  tokens: {
    colors: { light: Record<string, string>; dark?: Record<string, string> }
    fonts?: Record<string, string>
    radius?: string
    density?: string
  }
  layout?: Record<string, unknown>
  css?: string
  content: { hero: { heading?: Text; sub?: Text; cta?: Text; image?: string }; announcement?: Text }
}

export const THEME_PATH = "/api/nasaq/theme"

const hex = (v: string | undefined) => (v && /^#[0-9a-fA-F]{6}$/.test(v.trim()) ? v.trim().toUpperCase() : "")

/** Matjar's corner style for a Nasaq radius. */
export function matjarRadius(r?: string): "sharp" | "rounded" | "pill" {
  if (r === "none" || r === "sm") return "sharp"
  if (r === "pill" || r === "xl") return "pill"
  return "rounded"
}

/** Matjar's font for a theme's fonts: Inter when the Latin face is Inter, Lusail otherwise. */
export function matjarFont(fonts?: Record<string, string>): "inter" | "lusail" {
  return ["latin", "headingLatin", "sans"].some((k) => (fonts?.[k] ?? "").toLowerCase().includes("inter")) ? "inter" : "lusail"
}

/** The Matjar colours, corners, font and mode of a package (see Go matjar.BrandOf). */
export function matjarLook(p: ThemePackage) {
  const primary = hex(p.tokens.colors.light.primary)
  const accent = hex(p.tokens.colors.light.accent) || primary
  const lightOnly = !p.tokens.colors.dark || Object.keys(p.tokens.colors.dark).length === 0
  const tokens: Record<string, string> = {}
  if (lightOnly) {
    const map: Record<string, string> = { background: "--nq-bg", foreground: "--nq-fg", muted: "--nq-surface-soft", border: "--nq-line" }
    for (const [k, v] of Object.entries(map)) {
      const c = hex(p.tokens.colors.light[k])
      if (c) tokens[v] = c
    }
  }
  return { primary, accent, radius: matjarRadius(p.tokens.radius), font: matjarFont(p.tokens.fonts), mode: lightOnly ? "light" : "system", tokens }
}
