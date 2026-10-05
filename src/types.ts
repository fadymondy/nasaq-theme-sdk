/** Copy in several locales: { ar: "...", en: "..." }. */
export type Text = Partial<Record<string, string>>

export interface ColorPair {
  light: string
  dark: string
}

/** A site's applied theme, as the server stores it. */
export interface SiteTheme {
  /** "nasaq" or "default". */
  source?: string
  id?: string
  version?: string
  name?: Text
  project_id?: string
  accent?: ColorPair
  background?: ColorPair
  foreground?: ColorPair
  radius?: string
  density?: string
  own_action?: ColorPair
  applied_at?: string
}

/** The theme fields public pages render (theme.SiteTheme.Public in Go). */
export interface PublicSiteTheme {
  source: string
  id?: string
  accent?: ColorPair
  background?: ColorPair
  foreground?: ColorPair
  radius?: string
  density?: string
}

/** The site's link to its CircleXO organisation. */
export interface Link {
  org_id: string
  org_name?: string
  project_id?: string
  project_name?: string
  linked_by?: string
  linked_at: string
  last_pull_at?: string
  /** GRANT_LOST when the CircleXO grant expired; otherwise the last pull's error. */
  last_error?: string
}

export interface ThemeCircleXO {
  configured: boolean
  link?: Link
}

export interface ThemeStatus {
  theme: SiteTheme
  action: ColorPair
  /** The platform's default Nasaq theme; empty means "reset" goes back to the app's own look. */
  default_theme?: string
  circlexo: ThemeCircleXO
}

export interface ThemeCard {
  id: string
  version: string
  name: Text
  description?: Text
  category?: string
  free: boolean
  thumbnail?: string
  thumbnail_dark?: string
  demo_url?: string
  customize_url: string
}

export interface ThemeConnectView {
  auth_url?: string
  status?: ThemeStatus
}

export interface ProjectBinding {
  app: string
  app_ref: string
}

export interface Project {
  id: string
  slug: string
  name: Text
  default_locale?: string
  is_primary?: boolean
  status?: string
  bindings?: ProjectBinding[] | null
}

/** Link.last_error when the CircleXO grant expired: the owner has to connect again. */
export const GRANT_LOST = "grant_lost"
