# Go/npm parity fixture

`parity.json` is read by `theme/parity_test.go` and `test/parity.test.mjs`.
Contrast ratios use the WCAG relative-luminance formula, independently calculated
from the listed sRGB pairs; the tolerance is `1e-10`. Cases cover equal colours,
both luminance branches, non-neutral colours, and a ratio just below 4.5.

The theme input and expected public projection check the Go server's JSON
contract. Node consumes that same projection and checks the complete CSS output.
CSS generation is npm-only; the Go SDK does not expose a CSS formatter.

Run `go test ./theme` and `npm run build && npm test`. A contract change should
update both sides deliberately, rather than regenerate expected values from one
implementation during a test run.
