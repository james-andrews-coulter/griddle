# Griddle

A tiny, self-hosted RSS filter proxy with a visual rule builder.

## Development

```bash
go run main.go                    # Start on :4080 (override with PORT env)
go test ./...                     # Run tests
go build -o rss-griddle .         # Build binary
DATA_FILE=./feeds.json go run .   # Use a local data file
```

## Architecture

Single-file Go app (`main.go`, ~1000 lines):

- **Data model** — `Feed`, `FilterGroup`, `Rule` structs persisted as a single JSON file.
- **Filter engine** — builds an [expr-lang](https://expr-lang.org/) expression from each feed's nested rule groups, compiles once per request, evaluates per item. Fail-open on compile error.
- **HTTP handlers** — CRUD for feeds, filtered RSS XML output (`/api/feed?name=...`), HTMX partials for dynamic form interactions.
- **Output preservation** — feeds are filtered by mutating the original XML in place via [etree](https://github.com/beevik/etree), so custom XML fields, namespaces, and attributes survive into the output. Standard reconstruction libraries (gorilla/feeds and friends) only emit a fixed set of fields.
- **Live preview** (`POST /api/dryrun`) — debounced dry-run that returns each upstream `<item>` labeled pass/filter, rendered as XML beside the rule editor. Backed by a 5-minute in-process URL-keyed cache.
- **Templates** — inline Go templates with htmx for swap-driven editing. terminal.css 0.7.5 + ~30 lines of custom CSS for the workspace layout and live preview pane.

## Key patterns

- `buildExpr()` converts nested rule groups into a single expr-lang expression string with proper grouping (`(group1) && (group2)` etc.).
- `filterItems()` compiles the expression once, evaluates per item.
- All string comparisons are case-insensitive (lowercased on both sides).
- Missing fields default to empty string. A `contains` rule on a missing field doesn't match; a `not_contains` rule does.
- Namespaced XML tags (`dc:creator`, `media:thumbnail`) are flattened to `<prefix>_<tag>` keys (since `expr-lang` parses `:` as a type op). Repeated tags join by comma; tags with empty `Value` fall back to common URL-bearing attributes (`url`, `href`).

## Design system

`PRODUCT.md` and `DESIGN.md` capture the strategic and visual systems. The aesthetic is committed and intentional, not generic. Read both before making UI changes. PRs that introduce a second font, rounded corners, gradients, or shadows will be asked to adjust.
