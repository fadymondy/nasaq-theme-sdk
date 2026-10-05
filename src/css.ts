import type { PublicSiteTheme } from "./types.js"

/** Nasaq's radius scale per theme radius: surface, card, control, floating. */
export const RADII: Record<string, [string, string, string, string]> = {
  none: ["0px", "0px", "0px", "0px"],
  sm: ["0px", "4px", "4px", "8px"],
  md: ["0px", "10px", "8px", "12px"],
  lg: ["0px", "14px", "12px", "18px"],
  pill: ["0px", "18px", "999px", "22px"],
}

/** v when it is a #RRGGBB colour, else "": nothing else reaches a style sheet. */
export function safeHex(v: string | undefined): string {
  const s = (v ?? "").trim()
  return /^#[0-9a-fA-F]{6}$/.test(s) ? s : ""
}

export interface ThemeCSS {
  /** Declarations for the light scheme. */
  light: string
  /** Declarations for the dark scheme. */
  dark: string
  /** Declarations for both schemes. */
  radius: string
}

/**
 * A Nasaq theme's ground and ink as Nasaq's own variables: the surfaces, lines and quieter inks
 * are mixed from those two, so every Nasaq component follows the theme without new tokens.
 * Empty for a site on the default look. `safe` vets each colour; a colour it empties is skipped.
 */
export function themeCSS(th: PublicSiteTheme | undefined | null, safe: (v: string | undefined) => string = safeHex): ThemeCSS {
  const out = { light: "", dark: "", radius: "" }
  if (!th || th.source !== "nasaq") return out
  const mix = (a: string, b: string, p: number) => `color-mix(in oklab,${a},${b} ${p}%)`
  const ground = (bg: string, fg: string, dark: boolean) => {
    const lift = dark ? fg : "white"
    return (
      `--nq-bg:${bg};--nq-fg:${fg};` +
      `--nq-surface:${mix(bg, lift, dark ? 5 : 45)};--nq-surface-raised:${mix(bg, lift, dark ? 9 : 65)};` +
      `--nq-surface-overlay:${mix(bg, lift, dark ? 14 : 85)};--nq-surface-soft:${mix(bg, fg, dark ? 8 : 4)};` +
      `--nq-line:${mix(bg, fg, dark ? 16 : 12)};--nq-line-strong:${mix(bg, fg, dark ? 45 : 25)};` +
      `--nq-fg-body:${mix(fg, bg, dark ? 18 : 25)};--nq-fg-muted:${mix(fg, bg, dark ? 40 : 42)};`
    )
  }
  const bl = safe(th.background?.light), fl = safe(th.foreground?.light)
  if (bl && fl) out.light += ground(bl, fl, false)
  const bd = safe(th.background?.dark), fd = safe(th.foreground?.dark)
  if (bd && fd) out.dark += ground(bd, fd, true)
  const al = safe(th.accent?.light)
  if (al) out.light += `--nq-accent:${al};--nq-accent-text:${mix(al, fl || "black", 40)};`
  const ad = safe(th.accent?.dark)
  if (ad) out.dark += `--nq-accent:${ad};--nq-accent-text:${ad};`
  const r = th.radius ? RADII[th.radius] : undefined
  if (r) out.radius = `--nq-radius-surface:${r[0]};--nq-radius-card:${r[1]};--nq-radius-control:${r[2]};--nq-radius-floating:${r[3]};`
  return out
}

/**
 * The theme as one style sheet: radius on `root`, the light ground on `light`, the dark ground on
 * `dark`. The defaults fit Nasaq's `.dark` class; pass your own selectors otherwise.
 */
export function themeStyleSheet(
  th: PublicSiteTheme | undefined | null,
  sel: { root?: string; light?: string; dark?: string } = {},
): string {
  const css = themeCSS(th)
  const root = sel.root ?? ":root", light = sel.light ?? ":root", dark = sel.dark ?? ".dark"
  let out = ""
  if (css.radius) out += `${root}{${css.radius}}`
  if (css.light) out += `${light}{${css.light}}`
  if (css.dark) out += `${dark}{${css.dark}}`
  return out
}
