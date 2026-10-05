"use client"

import { useEffect, useMemo, useRef, useState, type ReactNode } from "react"
import { Alert, Spinner } from "@fadymondy/nasaq/web"

import type { ThemesClient } from "../client.js"
import { themeLabels, type ThemeLabels } from "./labels.js"

export interface CircleXOReturnProps {
  client: ThemesClient
  locale: string
  appName: string
  /** The query CircleXO returned with (code, state, error). */
  params: URLSearchParams
  /** Runs once the link is made: send the owner back to the Themes page. */
  onDone: () => void
  /** Runs once, before finishing (for example to forget a remembered site). */
  onStart?: () => void
  /** Shown under a refusal: a link or button back to the Themes page. */
  back?: ReactNode
  labels?: Partial<ThemeLabels>
  errorMessage?: (e: unknown) => string
}

/** CircleXO's return address: trades the code and the state for the link, then hands back. */
export function CircleXOReturn({ client, locale, appName, params, onDone, onStart, back, labels, errorMessage }: CircleXOReturnProps) {
  const t = useMemo(() => themeLabels(locale, appName, labels), [locale, appName, labels])
  const [error, setError] = useState<string | null>(null)
  const once = useRef(false)

  useEffect(() => {
    if (once.current) return
    once.current = true
    onStart?.()
    const code = params.get("code")
    const state = params.get("state")
    if (params.get("error") || !code || !state) {
      setError(t.denied)
      return
    }
    client
      .connectFinish(code, state)
      .then(onDone)
      .catch((e) => setError(errorMessage?.(e) || (e instanceof Error && e.message) || t.denied))
  }, [client, params, onDone, onStart, errorMessage, t])

  if (error) {
    return (
      <div className="flex flex-col items-start gap-4">
        <Alert tone="danger">{error}</Alert>
        {back}
      </div>
    )
  }
  return (
    <p className="flex items-center gap-2 text-sm text-muted-foreground">
      <Spinner className="size-4" />
      {t.finishing}
    </p>
  )
}
