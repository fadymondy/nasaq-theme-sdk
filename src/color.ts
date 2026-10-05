/** WCAG floors, the same as the Go SDK: an action colour against its label, and body ink on its ground. */
export const MIN_ACTION_CONTRAST = 3.0
export const MIN_INK_CONTRAST = 4.5

/** The two labels an action colour may carry, light first. */
export const ON_ACTION_CANDIDATES = ["#FFFFFF", "#0E1A3C"] as const

/** v as upper-case #RRGGBB, or "" when it is not one. */
export function hex(v: string | undefined): string {
  const s = (v ?? "").trim()
  return /^#[0-9a-fA-F]{6}$/.test(s) ? s.toUpperCase() : ""
}

function luminance(h: string): number {
  const c = [1, 3, 5].map((i) => {
    const v = parseInt(h.slice(i, i + 2), 16) / 255
    return v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]
}

/** The WCAG contrast ratio of a and b; 1 when either is not a #RRGGBB colour. */
export function contrast(a: string, b: string): number {
  const x = hex(a), y = hex(b)
  if (!x || !y) return 1
  const [hi, lo] = [luminance(x), luminance(y)].sort((p, q) => q - p)
  return (hi + 0.05) / (lo + 0.05)
}

/** The candidate label with the better contrast on action. */
export function onActionFor(action: string): string {
  const [a, b] = ON_ACTION_CANDIDATES
  return contrast(action, a) >= contrast(action, b) ? a : b
}
