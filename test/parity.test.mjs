import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { test } from "node:test"
import { contrast, themeCSS } from "../dist/index.js"

const fixture = JSON.parse(readFileSync(new URL("../testdata/parity.json", import.meta.url), "utf8"))

test("contrast follows the shared Go/npm fixture in both directions", () => {
  assert.ok(fixture.contrast.length > 0)
  for (const { a, b, ratio } of fixture.contrast) {
    for (const [first, second] of [[a, b], [b, a]]) {
      assert.ok(Math.abs(contrast(first, second) - ratio) <= 1e-10, `${first} / ${second}`)
    }
  }
})

test("the Go public theme fixture produces the expected CSS variables", () => {
  assert.deepEqual(themeCSS(fixture.theme.public), fixture.theme.css)
  assert.deepEqual(themeCSS(fixture.theme.input), fixture.theme.css)
})
