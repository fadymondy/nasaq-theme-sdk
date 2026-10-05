import type { Project, Text, ThemeCard, ThemeConnectView, ThemeStatus } from "./types.js"

/**
 * How the client reaches the app's API. `path` is relative to the app's themes base: "" for the
 * status, "/catalog?locale=ar", "/use", ... `body` is JSON. Throw on a failed answer: the Themes
 * page shows the error's `message`.
 */
export type Fetcher = <T>(path: string, init?: { method?: string; body?: unknown }) => Promise<T>

/** The app's themes endpoints, the same on every app (see the README). */
export interface ThemesClient {
  status(): Promise<ThemeStatus>
  catalog(locale: string): Promise<ThemeCard[]>
  use(id: string): Promise<ThemeStatus>
  reset(): Promise<ThemeStatus>
  connectStart(redirectUri: string): Promise<ThemeConnectView>
  connectFinish(code: string, state: string): Promise<ThemeConnectView>
  disconnect(): Promise<ThemeStatus>
  projects(): Promise<Project[]>
  bind(projectId: string): Promise<ThemeStatus>
  pull(): Promise<ThemeStatus>
}

export function createThemesClient(f: Fetcher): ThemesClient {
  const post = <T>(path: string, body?: unknown) => f<T>(path, { method: "POST", body })
  return {
    status: () => f<ThemeStatus>(""),
    catalog: (locale) =>
      f<{ items: ThemeCard[] | null }>(`/catalog?locale=${encodeURIComponent(locale)}`).then((r) => r.items ?? []),
    use: (id) => post<ThemeStatus>("/use", { id }),
    reset: () => post<ThemeStatus>("/reset"),
    connectStart: (redirect_uri) => post<ThemeConnectView>("/circlexo", { action: "start", redirect_uri }),
    connectFinish: (code, state) => post<ThemeConnectView>("/circlexo", { action: "finish", code, state }),
    disconnect: () => f<ThemeStatus>("/circlexo", { method: "DELETE" }),
    projects: () => f<{ items: Project[] | null }>("/projects").then((r) => r.items ?? []),
    bind: (project_id) => post<ThemeStatus>("/bind", { project_id }),
    pull: () => post<ThemeStatus>("/pull"),
  }
}

/** Copy in locale, falling back to English, then Arabic. */
export function pick(t: Text | undefined, locale: string): string {
  return t?.[locale] || t?.en || t?.ar || ""
}
