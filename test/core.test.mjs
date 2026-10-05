import assert from "node:assert/strict"
import { test } from "node:test"

import { contrast, createThemesClient, hex, onActionFor, pick, themeCSS, themeStyleSheet } from "../dist/index.js"
import { themeLabels } from "../dist/react/labels.js"

const nasaq = {
  source: "nasaq",
  id: "clinic",
  accent: { light: "#E8A87C", dark: "#E8A87C" },
  background: { light: "#FBFEFD", dark: "#0B1F1E" },
  foreground: { light: "#10302F", dark: "#E6F4F2" },
  radius: "lg",
}

test("themeCSS writes Nasaq's variables", () => {
  const css = themeCSS(nasaq)
  assert.match(css.light, /--nq-bg:#FBFEFD;--nq-fg:#10302F;/)
  assert.match(css.dark, /--nq-bg:#0B1F1E;/)
  assert.match(css.light, /--nq-accent:#E8A87C;--nq-accent-text:color-mix\(in oklab,#E8A87C,#10302F 40%\);/)
  assert.equal(css.radius, "--nq-radius-surface:0px;--nq-radius-card:14px;--nq-radius-control:12px;--nq-radius-floating:18px;")
})

test("themeCSS is empty for the default look and drops what is not a colour", () => {
  assert.deepEqual(themeCSS({ source: "default" }), { light: "", dark: "", radius: "" })
  assert.deepEqual(themeCSS(undefined), { light: "", dark: "", radius: "" })
  const css = themeCSS({ ...nasaq, background: { light: "red;}body{", dark: "#0B1F1E" }, radius: "huge" })
  assert.equal(css.light.includes("--nq-bg"), false)
  assert.equal(css.light.includes("}"), false)
  assert.match(css.dark, /--nq-bg:#0B1F1E;/)
  assert.equal(css.radius, "")
})

test("themeStyleSheet wraps the declarations", () => {
  const sheet = themeStyleSheet(nasaq, { dark: "[data-theme=dark]" })
  assert.match(sheet, /^:root\{--nq-radius-surface/)
  assert.match(sheet, /\[data-theme=dark\]\{--nq-bg:#0B1F1E/)
})

test("contrast matches the Go SDK", () => {
  assert.ok(Math.abs(contrast("#FFFFFF", "#000000") - 21) < 0.01)
  assert.equal(contrast("red", "#000000"), 1)
  assert.equal(onActionFor("#FFFF00"), "#0E1A3C")
  assert.equal(onActionFor("#0E1A3C"), "#FFFFFF")
  assert.equal(hex(" #abcdef "), "#ABCDEF")
  assert.equal(hex("#abc"), "")
})

test("the client calls the app's themes endpoints", async () => {
  const calls = []
  const c = createThemesClient(async (path, init) => {
    calls.push([path, init?.method ?? "GET", init?.body])
    return path.startsWith("/catalog") || path === "/projects" ? { items: null } : {}
  })
  assert.deepEqual(await c.catalog("ar"), [])
  await c.use("clinic")
  await c.connectStart("https://app/back")
  await c.disconnect()
  assert.deepEqual(await c.projects(), [])
  assert.deepEqual(calls, [
    ["/catalog?locale=ar", "GET", undefined],
    ["/use", "POST", { id: "clinic" }],
    ["/circlexo", "POST", { action: "start", redirect_uri: "https://app/back" }],
    ["/circlexo", "DELETE", undefined],
    ["/projects", "GET", undefined],
  ])
})

test("labels name the app in both languages", () => {
  const en = themeLabels("en", "Matjar"), ar = themeLabels("ar", "Matjar", { use: "Apply" })
  assert.equal(en.defaultName, "Matjar default")
  assert.equal(ar.defaultName, "Matjar الافتراضي")
  assert.equal(ar.use, "Apply")
  for (const l of [en, ar]) for (const [k, v] of Object.entries(l)) assert.ok(v && !v.includes("{app}"), k)
  assert.equal(pick({ en: "A", ar: "ب" }, "ar"), "ب")
  assert.equal(pick({ ar: "ب" }, "en"), "ب")
})
