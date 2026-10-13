import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { test } from "node:test"

import { matjarLook, seedPack, seedPacks, TEMPLATES } from "../dist/products/index.js"

test("the npm seed packs are the Go SDK's packs", () => {
  const go = JSON.parse(readFileSync(new URL("../product/matjar/packs/sahar.json", import.meta.url), "utf8"))
  assert.deepEqual(seedPack("matjar", "sahar"), go)
  assert.equal(seedPacks("matjar").length, 1)
  assert.equal(seedPack("matjar", "nope"), undefined)
})

test("matjar covers the six storefront templates", () => {
  assert.deepEqual(TEMPLATES.matjar.map((t) => t.key), ["home", "collection", "product", "cart", "checkout", "account"])
})

test("matjarLook maps like the Go BrandOf", () => {
  const p = {
    schema: 1, product: "matjar", theme: "store", version: "1", brand: {}, content: { hero: {} },
    tokens: { colors: { light: { primary: "#2c4a38", accent: "#d9a441", background: "#fbfaf7" } }, radius: "sm", fonts: { latin: "Inter" } },
  }
  assert.deepEqual(matjarLook(p), { primary: "#2C4A38", accent: "#D9A441", radius: "sharp", font: "inter", mode: "light", tokens: { "--nq-bg": "#FBFAF7" } })
  const d = matjarLook({ ...p, tokens: { ...p.tokens, colors: { ...p.tokens.colors, dark: { primary: "#a8c9b1" } } } })
  assert.equal(d.mode, "system")
  assert.deepEqual(d.tokens, {})
})
