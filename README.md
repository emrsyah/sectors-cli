# sectors-cli

A command-line client for the [Sectors Financial API v2](https://docs.sectors.app)
covering Indonesia (IDX), Singapore (SGX), Malaysia (KLSE), and Indonesian mining
data.

It is **built to be driven by AI agents as well as humans**: output defaults to
the API's raw JSON (pretty-printed in a terminal, compact when piped), failed
requests exit non-zero with a JSON error on stderr, and `--help` on any command
documents its parameters straight from the API spec — so an agent can discover
capabilities by reading help.

## Install

```bash
# Released binary (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/emrsyah/sectors-cli/main/install.sh | sh

# Or with Go
go install github.com/emrsyah/sectors-cli/cmd/sectors@latest
```

Or download a prebuilt archive for your platform from the
[releases page](https://github.com/emrsyah/sectors-cli/releases).

## Quick start

```bash
# Build from source
go build -o sectors ./cmd/sectors

# Authenticate (get a key at https://sectors.app/api)
./sectors auth login --api-key <your-key>      # or: export SECTORS_API_KEY=<key>
./sectors auth status

# Use it
./sectors idx company report BBCA
./sectors idx company report BREN --sections overview,valuation
./sectors idx company report BBCA -o pretty
```

## Output formats (`-o` / `--output`)

| value | behaviour |
|---|---|
| `auto` (default) | pretty JSON on an interactive terminal, compact JSON when piped (best for agents) |
| `json` | compact single-line JSON |
| `pretty` | indented JSON |
| `table` | ASCII table for list responses (arrays or `{results:[…]}`); falls back to pretty JSON for nested/single-object responses |

```bash
sectors idx transaction daily BBCA --start 2026-06-20 --end 2026-06-26 -o table
sectors idx companies --q "top 5 banks by revenue" | jq '.results[].symbol'
```

Failed requests print a JSON error to **stderr** and exit non-zero, so scripts
and agents can branch on the result.

## Agent integration

This CLI is designed to be driven by AI agents.

**Tool manifest.** Generate tool/function-calling schemas for every command,
straight from the command tree (so they never drift from the real flags):

```bash
sectors manifest --format anthropic        # tool-use definitions
sectors manifest --format openai           # function-calling
sectors manifest --format json             # neutral
sectors manifest --filter "idx_*"          # export a subset
```

Each leaf command becomes one tool (e.g. `idx_company_report`) with a JSON-Schema
`input_schema`: positional args and flags become typed properties, required flags
land in `required`, and `a|b|c` choice flags become `enum`s.

**Structured errors.** Failures emit a JSON envelope on stderr with a stable
`category` and `retryable` so an agent can branch without scraping text:

```json
{"error":"request failed","status":429,"category":"rate_limited","retryable":true}
```

**Exit codes** mirror the category, so shell/agent logic can branch without
parsing JSON at all:

| code | meaning |
|---|---|
| 0 | success |
| 1 | client-side / config / transport error |
| 2 | invalid input (400/422) |
| 3 | auth (401/403) |
| 4 | not found (404) |
| 5 | rate limited (429) |
| 6 | server error (5xx) |

**Automatic retries.** Transient failures (429, 500, 502, 503, 504, and network
errors) are retried with exponential backoff + jitter, honoring `Retry-After`.
Tune with `--retries` (default 3) and `--retry-max-wait` (default 10s); set
`--retries 0` to disable.

**Response shaping (token economy).** Shrink responses *before* they reach the
model's context window:

```bash
# keep only the fields you need (dotted paths; key[] maps over an array)
sectors idx company report BBCA --select overview.market_cap,valuation
sectors idx companies --where "sub_sector='banks'" --select "results[].symbol"

# cap a list (adds "_truncated": true to paginated responses)
sectors idx news list --max 5

# just the count
sectors idx list subsectors --count        # -> {"count": 33}
```

These run client-side after the request, so they compose with the server-side
`--sections` / `--limit` filters. In practice `--select` cuts a full company
report from ~56 KB to a few dozen bytes.

## Caching & observability

**Caching.** GET responses are cached on disk to cut latency, cost, and
rate-limit pressure for the repetitive reads agents make. TTLs vary by endpoint
volatility (reference lists ~24h, intraday data ~1m, reports ~5m). Override or
inspect:

```bash
sectors idx list subsectors --no-cache       # bypass the cache
sectors idx daily BBCA --cache-ttl 30s        # uniform TTL override
sectors cache path                            # where the cache lives
sectors cache clear                           # wipe it
```

**Tracing.** `--verbose` (`-v`) logs each request to stderr; `--dry-run` prints
the request that *would* be sent (API key redacted) without calling the API:

```bash
$ sectors idx list subsectors -v
GET https://api.sectors.app/v2/subsectors/ -> 200 1ms (cache hit)

$ sectors idx company report BBCA --dry-run
{"dry_run":true,"method":"GET","url":".../v2/company/report/BBCA/","headers":{"Authorization":"***redacted***"}}
```

## Shell completions

Cobra generates completions for bash, zsh, fish, and PowerShell:

```bash
sectors completion zsh > "${fpath[1]}/_sectors"     # zsh
sectors completion bash | sudo tee /etc/bash_completion.d/sectors   # bash
sectors completion fish > ~/.config/fish/completions/sectors.fish   # fish
```

Once installed, completing the `-o`/`--output` flag suggests the valid formats
with descriptions (`auto`, `json`, `pretty`, `table`) — e.g. `sectors … -o <TAB>`.

## Releasing

Versioned cross-platform binaries are built with [goreleaser](https://goreleaser.com)
(`.goreleaser.yaml`). The version is injected via ldflags and surfaced by
`sectors --version`.

```bash
goreleaser release --snapshot --clean   # local snapshot build
git tag v0.1.0 && goreleaser release --clean   # tagged release
```

## Tests

```bash
go test ./...
```

## Architecture

| Layer | What |
|---|---|
| **API client** | Generated by [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) from `sectors-schema.json` (the upstream OpenAPI 3.0 spec). 63 typed operations in `internal/sectors/sectors.gen.go`. |
| **CLI framework** | [`spf13/cobra`](https://github.com/spf13/cobra) — command tree, flags, auto `--help` and shell completions. |
| **Config / auth** | `internal/config` — resolves the API key with precedence: `--api-key` flag → `$SECTORS_API_KEY` → config file (`<user-config-dir>/sectors/config.yaml`, 0600). |
| **Output** | `internal/output` — raw-JSON renderer with TTY auto-detection (pretty on a terminal, compact when piped) and structured JSON errors on stderr. |

### Regenerating the client

`tools/fixspec` sanitizes the upstream spec into `sectors-schema.fixed.json`
before generation (the pristine `sectors-schema.json` is never mutated). It:

1. drops malformed paths (a required path parameter missing from the URL template), which `oapi-codegen` rejects; and
2. empties every **response** body schema so response bodies generate as `interface{}`. The upstream response schemas are frequently inaccurate (arrays typed as objects, string-vs-number, offset-less timestamps), which would make the generated response parser reject otherwise-valid payloads. The CLI emits the raw response body and never reads the typed fields, so untyped bodies are strictly safer. Request **parameter** types are untouched.

```bash
go generate ./...   # runs fixspec, then oapi-codegen
```

## Command tree

- `auth login | status`
- `idx` — Indonesia (complete): ✅
  - `companies` (screener: `--where` / `--q` / `--order-by` / `--limit`) · `free-float`
  - `company report | segments | financials | corporate-actions | shareholders | ipo-performance | quarterly-dates`
  - `subsector report`
  - `brokers activity | activity-top | summary | summary-top | registry | top | foreign-flow`
  - `transaction daily | idx-total | index-daily`
  - `ranking most-traded | top-changes`
  - `news list | filings | suspensions`
  - `list industries | subindustries | subsectors | tags | segments-companies`
- `sgx` — Singapore (complete): ✅ `companies` (screener) · `top` · `report` · `sectors` · `tags` · `news` · `filings` · `buybacks` · `short-sell` · `daily`
- `klse` — Malaysia (complete): ✅ `companies --sector` · `top` · `report` · `sectors`
- `mining` — Mining extension (complete): ✅
  - `companies list | get | financials | ownership | performance`
  - `commodities list | price | exports | global | sales-destination`
  - `sites list | get` · `production total` · `reserves index | get`
  - `licenses list` · `auctions list | get` · `contracts list`

All **63 API operations** are wired to commands.

### Code layout

Each market is its own package exposing a `NewCmd()` constructor; `cmd/cmdutil`
holds the shared client factory, response/error rendering, and optional-flag →
pointer helpers. Adding a market is a new `cmd/<market>/` package wired into
`cmd/root.go`.

```
cmd/
├── root.go            # global flags, Execute, wires market packages
├── auth.go
├── cmdutil/           # NewClient, Emit, Fail, Do, Opt{Str,Int,Bool,Float,Enum}
├── idx/               # company.go companies.go brokers.go market.go news.go list.go
├── sgx/               # companies.go reference.go news.go transaction.go
├── klse/              # klse.go
└── mining/            # companies.go commodities.go sites.go licenses.go
```

See `sectors-docs/_ENDPOINT_INVENTORY.md` for the full endpoint → command map.
