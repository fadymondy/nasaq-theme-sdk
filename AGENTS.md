# AGENTS.md

Notes for AI coding agents working in this repository.

- **What it is:** Nasaq themes for any app, in two halves that share one HTTP contract (see the README):
  the Go module (`theme`, `nasaq`, `circlexo`, `nasaqtest`, `product`) and the npm package (`src/`,
  with `src/react` for the Themes page).
- **Checks:** `gofmt -l .` (must print nothing), `go vet ./...`, `go test ./...`, then
  `npm ci && npm run typecheck && npm run build && npm test`.
- **Contract changes** touch both halves in the same pull request, with tests on each side.
- **Generated files:** product packs come from `npm run gen` (`scripts/gen-packs.mjs`); regenerate
  them rather than editing by hand.
- **Secrets:** none belong in this repo. Tests use the fakes in `nasaqtest` and `circlexotest`.
