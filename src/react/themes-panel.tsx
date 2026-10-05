"use client"

import { useEffect, useMemo, useState } from "react"
import {
  Alert,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
  ConfirmButton,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Skeleton,
  Spinner,
} from "@fadymondy/nasaq/web"

import type { ThemesClient } from "../client.js"
import { pick } from "../client.js"
import type { ColorPair, Project, ThemeCard, ThemeStatus } from "../types.js"
import { GRANT_LOST } from "../types.js"
import { themeLabels, type ThemeLabels } from "./labels.js"

export interface ThemesPanelProps {
  /** The app's themes API (createThemesClient). */
  client: ThemesClient
  /** "ar" or "en". */
  locale: string
  /** The app's name, in the default theme's card and the copy ("SeatFor", "Matjar"). */
  appName: string
  /** Where CircleXO sends the owner back: the page that renders <CircleXOReturn>. */
  redirectUri: string | (() => string)
  /** Runs before the browser leaves for CircleXO (for example to remember which site it was). */
  onBeforeConnect?: () => void
  /** Overrides for the built-in labels. */
  labels?: Partial<ThemeLabels>
  /** The message for a failed call. Default: the error's `message`. */
  errorMessage?: (e: unknown) => string
}

/**
 * The Themes page body: the theme in use, Nasaq Studio through CircleXO, and the catalog. The
 * host renders the page frame (title and subtitle are in `themeLabels`).
 */
export function ThemesPanel({ client, locale, appName, redirectUri, onBeforeConnect, labels, errorMessage }: ThemesPanelProps) {
  const t = useMemo(() => themeLabels(locale, appName, labels), [locale, appName, labels])
  const [status, setStatus] = useState<ThemeStatus | null>(null)
  const [catalog, setCatalog] = useState<ThemeCard[] | null>(null)
  const [catalogErr, setCatalogErr] = useState<string | null>(null)
  const [projects, setProjects] = useState<Project[] | null>(null)
  const [project, setProject] = useState("")
  const [busy, setBusy] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)

  const message = (e: unknown) => errorMessage?.(e) || (e instanceof Error && e.message) || t.error
  const fail = (e: unknown) => setErr(message(e))
  const linked = !!status?.circlexo.link

  useEffect(() => {
    client.status().then(setStatus).catch(fail)
    client
      .catalog(locale)
      .then(setCatalog)
      .catch((e) => (setCatalog([]), setCatalogErr(message(e))))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client, locale])

  useEffect(() => {
    if (!linked) return
    client
      .projects()
      .then(setProjects)
      .catch((e) => (setProjects([]), fail(e)))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client, linked])

  const run = (key: string, p: () => Promise<ThemeStatus>) => {
    setBusy(key)
    setErr(null)
    return p()
      .then(setStatus)
      .catch((e) => {
        fail(e)
        throw e
      })
      .finally(() => setBusy(null))
  }
  const act = (key: string, p: () => Promise<ThemeStatus>) => void run(key, p).catch(() => {})

  const connect = () => {
    setBusy("connect")
    setErr(null)
    onBeforeConnect?.()
    client
      .connectStart(typeof redirectUri === "function" ? redirectUri() : redirectUri)
      .then((r) => {
        if (r.auth_url) window.location.assign(r.auth_url)
      })
      .catch((e) => (fail(e), setBusy(null)))
  }

  if (!status) {
    return (
      <div className="grid gap-4">
        {err && <Alert tone="danger">{err}</Alert>}
        {!err && <Rows />}
      </div>
    )
  }

  const theme = status.theme
  const onNasaq = theme.source === "nasaq"
  const link = status.circlexo.link

  return (
    <div className="grid gap-8">
      {err && <Alert tone="danger">{err}</Alert>}

      <section className="flex flex-wrap items-center gap-4">
        <div className="grid gap-1">
          <p className="text-xs text-muted-foreground">{t.current}</p>
          <p className="font-medium">{onNasaq ? pick(theme.name, locale) || theme.id : t.defaultName}</p>
          {onNasaq && theme.project_id && <p className="text-xs text-muted-foreground">{t.fromProject}</p>}
        </div>
        <Swatches label={t.colours} pairs={[status.action, theme.accent, theme.background]} />
        {onNasaq && (
          <div className="ms-auto">
            <ConfirmButton
              variant="secondary"
              title={t.resetConfirm}
              description={status.default_theme ? t.resetToPlatform : t.resetToOwn}
              confirmLabel={t.reset}
              onConfirm={() => run("reset", client.reset)}
            >
              {t.reset}
            </ConfirmButton>
          </div>
        )}
      </section>

      <section className="grid gap-4">
        <h2 className="text-sm font-medium">{t.studioTitle}</h2>
        {!status.circlexo.configured ? (
          <p className="text-sm text-muted-foreground">{t.notConfigured}</p>
        ) : !link ? (
          <div className="flex flex-wrap items-center gap-4">
            <p className="max-w-prose text-sm">{t.pitch}</p>
            <Button className="ms-auto" loading={busy === "connect"} disabled={busy !== null} onClick={connect}>
              {t.connect}
            </Button>
          </div>
        ) : (
          <>
            <div className="flex flex-wrap items-center gap-3">
              <Badge variant="success">{t.connected}</Badge>
              <span className="font-medium">{link.org_name || <bdi dir="ltr">{link.org_id}</bdi>}</span>
              <div className="ms-auto">
                <ConfirmButton
                  title={t.disconnectConfirm}
                  description={t.disconnectNote}
                  confirmLabel={t.disconnect}
                  onConfirm={() => run("disconnect", client.disconnect)}
                >
                  {t.disconnect}
                </ConfirmButton>
              </div>
            </div>

            <div className="flex flex-wrap items-end gap-3">
              <div className="grid gap-2">
                <p className="text-sm">{t.project}</p>
                {projects === null ? (
                  <Spinner className="size-4 text-muted-foreground" />
                ) : projects.length === 0 ? (
                  <p className="text-sm text-muted-foreground">{t.noProjects}</p>
                ) : (
                  <Select
                    items={projects.map((p) => ({ value: p.id, label: pick(p.name, locale) || p.slug }))}
                    value={project || link.project_id || ""}
                    onValueChange={(v) => setProject((v as string | null) ?? "")}
                  >
                    <SelectTrigger className="w-64" aria-label={t.project}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {projects.map((p) => (
                        <SelectItem key={p.id} value={p.id}>
                          {pick(p.name, locale) || p.slug}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </div>
              {projects && projects.length > 0 && (
                <Button
                  variant="secondary"
                  loading={busy === "bind"}
                  disabled={busy !== null || !project || project === link.project_id}
                  onClick={() => act("bind", () => client.bind(project))}
                >
                  {t.useProject}
                </Button>
              )}
              {link.project_id && (
                <Button variant="secondary" loading={busy === "pull"} disabled={busy !== null} onClick={() => act("pull", client.pull)}>
                  {t.pull}
                </Button>
              )}
            </div>
            {link.project_id && <p className="text-xs text-muted-foreground">{t.auto}</p>}
            {link.last_pull_at && (
              <p className="text-xs text-muted-foreground">
                {t.lastPull} {new Date(link.last_pull_at).toLocaleString(locale)}
              </p>
            )}
            {link.last_error && (
              <Alert tone="danger">
                {link.last_error === GRANT_LOST ? t.grantLost : t.pullFailed}
                {link.last_error !== GRANT_LOST && (
                  <bdi dir="ltr" className="mt-1 block text-xs">
                    {link.last_error}
                  </bdi>
                )}
              </Alert>
            )}
          </>
        )}
      </section>

      <section className="grid gap-4">
        <h2 className="text-sm font-medium">{t.catalogTitle}</h2>
        {catalogErr && <Alert tone="warning">{catalogErr}</Alert>}
        {catalog === null ? (
          <Rows />
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle>{t.defaultName}</CardTitle>
                <CardDescription>{t.defaultBody}</CardDescription>
              </CardHeader>
              <CardFooter className="mt-auto flex flex-wrap gap-2">
                {!onNasaq ? (
                  <Badge variant="brand">{t.inUse}</Badge>
                ) : (
                  <Button variant="secondary" loading={busy === "reset"} disabled={busy !== null} onClick={() => act("reset", client.reset)}>
                    {t.use}
                  </Button>
                )}
              </CardFooter>
            </Card>
            {catalog.map((c) => {
              const inUse = onNasaq && !theme.project_id && theme.id === c.id
              return (
                <Card key={c.id}>
                  {c.thumbnail && (
                    <img
                      src={c.thumbnail}
                      alt=""
                      loading="lazy"
                      className="aspect-video w-full rounded-[var(--nq-radius-card)] border border-border object-cover"
                    />
                  )}
                  <CardHeader>
                    <CardTitle className="flex flex-wrap items-center gap-2">
                      {pick(c.name, locale) || c.id}
                      {!c.free && <Badge variant="accent">{t.premium}</Badge>}
                    </CardTitle>
                    {c.description && <CardDescription>{pick(c.description, locale)}</CardDescription>}
                  </CardHeader>
                  <CardContent className="text-xs text-muted-foreground">
                    {c.category && <span>{c.category} · </span>}
                    <bdi dir="ltr">v{c.version}</bdi>
                  </CardContent>
                  <CardFooter className="mt-auto flex flex-wrap gap-2">
                    {inUse ? (
                      <Badge variant="brand">{t.inUse}</Badge>
                    ) : (
                      <Button loading={busy === c.id} disabled={busy !== null} onClick={() => act(c.id, () => client.use(c.id))}>
                        {t.use}
                      </Button>
                    )}
                    <Button variant="secondary" nativeButton={false} render={<a href={c.customize_url} target="_blank" rel="noopener noreferrer" />}>
                      {t.customize}
                    </Button>
                    {c.demo_url && (
                      <Button variant="ghost" nativeButton={false} render={<a href={c.demo_url} target="_blank" rel="noopener noreferrer" />}>
                        {t.demo}
                      </Button>
                    )}
                  </CardFooter>
                </Card>
              )
            })}
          </div>
        )}
        <p className="text-xs text-muted-foreground">{t.customizeNote}</p>
      </section>
    </div>
  )
}

function Rows() {
  return (
    <div className="grid gap-3">
      <Skeleton className="h-6 w-1/3" />
      <Skeleton className="h-6 w-2/3" />
      <Skeleton className="h-6 w-1/2" />
    </div>
  )
}

/** The theme's colours, light then dark, so the owner sees what the pages will wear. */
function Swatches({ label, pairs }: { label: string; pairs: (ColorPair | undefined)[] }) {
  const shown = pairs.filter((p): p is ColorPair => !!p?.light)
  return (
    <div className="flex items-center gap-1.5" role="img" aria-label={label}>
      {shown.flatMap((p, i) => [
        <span key={`${i}l`} className="size-5 rounded-full border border-border" style={{ background: p.light }} />,
        <span key={`${i}d`} className="size-5 rounded-full border border-border" style={{ background: p.dark || p.light }} />,
      ])}
    </div>
  )
}
