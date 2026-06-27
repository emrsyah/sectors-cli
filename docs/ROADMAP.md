# sectors-cli — Roadmap (agent-product foundation)

The CLI is feature-complete against the Sectors v2 API (all 63 endpoints, verified
live). This roadmap hardens it for its real job: **the programmatic interface an
agent product drives at scale, unattended.** Five workstreams, sequenced by
leverage. Each is independently shippable and written to be handed to one
engineer/agent.

Recommended order: **1 → 2 → 3** unlock the product; **4 → 5** make it durable.

Architecture reminders (so each task slots in cleanly):
- Command tree: `cmd/root.go` wires `cmd/<market>/NewCmd()`; shared logic in `cmd/cmdutil`.
- HTTP: generated client in `internal/sectors` (auth via `RequestEditorFn` in `client.go`).
- Output/errors: `internal/output` (raw-body emit, `EmitError`); the CLI never reads typed response bodies.
- Spec hygiene: `tools/fixspec` rewrites the upstream OpenAPI before codegen.

---

## 1. Agent capability manifest — `sectors manifest` ✅ DONE

> Shipped: `cmd/manifest.go` generates `--format json|anthropic|openai` from the
> Cobra tree (63 tools), `--filter` glob, args+flags→typed JSON-Schema, required
> flags and `a|b|c` enums detected. Tests in `cmd/manifest_test.go`.

**Goal.** Emit machine-readable tool definitions for every command, generated from
the Cobra tree, so a host agent loads the CLI as a callable toolset.

**Why (agent-product).** This is the bridge that makes "CLI instead of MCP
tool-calling" real. Generated from source → never drifts from the actual flags.
One catalog → the model can call any command (or a curated subset).

**Design.**
- New `cmd/manifest.go`: walk `rootCmd` recursively; for each leaf command emit
  `{name, description, input_schema}`.
  - `name`: dotted/underscored path, e.g. `idx_company_report`.
  - `description`: `cmd.Short` (+ first paragraph of `Long`).
  - `input_schema`: positional args (from `Use` tokens like `<symbol>`) → required
    string params; flags (`cmd.Flags()`) → typed properties (string/int/bool/number),
    pulling help text and enum hints from flag usage.
- `--format anthropic|openai|json` (default `json`). Anthropic = `input_schema`
  JSON-Schema; OpenAI = `function` wrapper; json = neutral.
- `--filter <glob>` to export a curated subset (e.g. `idx.*`).
- Reuse: factor the flag→type mapping so it can later feed an MCP server (#hedge).

**Key decision (ask product).** All ~62 commands as tools (max capability, big tool
list) vs curated/dynamic subset per agent task. Recommend: emit all, let the host
filter with `--filter`.

**Files.** `cmd/manifest.go`, test `cmd/manifest_test.go`.
**Acceptance.** `sectors manifest --format anthropic` produces valid tool schemas
for every leaf command; round-trips through the Anthropic SDK tool validator; a
golden test pins the output so flag changes are visible in diffs.
**Effort.** ~1 day.

---

## 2. Resilience: retry/backoff + typed error envelope ✅ DONE

> Shipped: `internal/sectors/retry.go` (RoundTripper; retries 429/500/502/503/504
> + network on GET, exp backoff + jitter, honors Retry-After; `--retries`,
> `--retry-max-wait`). Error envelope gains `category`+`retryable`
> (`internal/output`); `cmdutil.HandledError` carries category exit codes 2–6.
> Tests in `internal/sectors/retry_test.go`. Exit-code table in README.

**Goal.** Survive parallel, unattended traffic: auto-retry transient failures and
return errors agents can branch on.

**Why (agent-product).** Agents fan out and will hit 429s and transient 5xx
constantly. Today a 429 just fails. This is the line between demo and production.

**Design.**
- **Retry** as an `http.RoundTripper` wrapper in `internal/sectors` (compose with
  the existing client; keep the generated code untouched). Retry on 429, 502, 503,
  504, and network errors. Exponential backoff + jitter, honor `Retry-After`.
  Cap attempts (default 3) and total time; make both configurable via global flags
  `--retries` / `--retry-max-wait`. Never retry non-idempotent calls (all current
  endpoints are GET, so safe — but gate on method).
- **Typed error envelope.** Extend `output.EmitError` to include a `category` and
  `retryable`: `{error, status, category, retryable, response}`.
  Categories: `invalid_input` (400/422), `auth` (401/403), `not_found` (404),
  `rate_limited` (429), `server` (5xx), `network` (transport). Map in `cmdutil`.
- Stable **exit codes** by category (e.g. 2=invalid_input, 3=auth, 4=not_found,
  5=rate_limited, 6=server, 7=network) so shell/agent logic can branch without
  parsing JSON. Document the table.

**Files.** `internal/sectors/retry.go`, `internal/sectors/client.go` (wire the
transport), `internal/output/output.go` (envelope), `cmd/cmdutil/cmdutil.go`
(categorize from status), `cmd/root.go` (flags), tests.
**Acceptance.** A stubbed 429-then-200 server yields one transparent retry and a
success; a persistent 429 returns `category:rate_limited, retryable:true` and exit
5; `Retry-After` is honored. Unit-tested against an `httptest` server.
**Effort.** ~1–1.5 days.

---

## 3. Token & response economy ✅ DONE

> Shipped: `internal/output/project.go` — `--select` (dotted paths + `key[]`
> array wildcards, deep-merged), `--max` (truncates list, adds `_truncated` on
> paginated), `--count`. Applied in `cmdutil.Emit` via `shape()`; `--count`
> short-circuits. Tests in `internal/output/project_test.go`. Verified live:
> full company report ~56 KB → ~44 bytes with `--select`.

**Goal.** Shrink responses before they reach the model's context window.

**Why (agent-product).** Direct $ + context savings on every call. A full company
report is large; agents rarely need all of it.

**Design.**
- Global `--select <paths>`: comma-separated dotted JSON paths
  (`overview.market_cap,valuation`). Apply client-side in `internal/output`
  after fetch, before emit. Support array wildcards (`results[].symbol`).
  Prefer a tiny dependency-free path filter over pulling in full jq.
- Global `--max <n>`: truncate top-level arrays / `results[]` to n items and add a
  `_truncated` marker so the agent knows there's more.
- Lean on what the API already does first: `--sections` (reports) and `--limit`
  (lists) reduce server-side payload — document the precedence (server-side
  filters first, then `--select`/`--max` client-side).
- Optional `--count`: emit only the length of the result set.

**Files.** `internal/output/project.go` (+ tests), `cmd/root.go` (flags),
`cmd/cmdutil` (apply before `Emit`).
**Acceptance.** `--select` returns only requested paths for both object and
`results[]` shapes; `--max` truncates and marks; non-matching paths are omitted
(not error). Golden tests on representative payloads.
**Effort.** ~1 day.

---

## 4. CI/CD + spec-drift smoke tests + LICENSE ✅ DONE

> Shipped: `.github/workflows/ci.yml` (gofmt + build + vet + test, and a
> generated-fresh job: `go generate` then `git diff --exit-code` on the gen'd
> client), `release.yml` (goreleaser on `v*` tags), `smoke.yml` (daily live run).
> `scripts/smoke.sh` exercises ~62 commands with tolerant id discovery (404 on a
> discovered id = benign). `LICENSE` = MIT (holder "Supertype" — CONFIRM/CHANGE).
> Needs repo secret `SECTORS_API_KEY` for the smoke job. Verified locally: smoke
> 62/0, all CI gates green.

**Goal.** Reproducible, trustworthy releases; catch upstream spec drift before
agents do.

**Why (agent-product).** We already proved the upstream OpenAPI lies about response
shapes (fixed in `fixspec`). A scheduled live test catches the next drift early.
Pinned binaries make agent runtimes/containers reproducible.

**Design.**
- `.github/workflows/ci.yml`: on PR/push → `go build`, `go vet`, `go test ./...`,
  and `go generate` + `git diff --exit-code` (fails if generated client is stale).
- `.github/workflows/release.yml`: on tag `v*` → goreleaser (config already exists).
- `.github/workflows/smoke.yml`: scheduled (e.g. daily) → run a representative
  command per endpoint against the live API using a repo secret `SECTORS_API_KEY`;
  fail/alert on any non-2xx or client-side decode error. Reuse the QA command list.
- Add a `LICENSE` (the goreleaser archive already globs `LICENSE*`).
- Optional: `install.sh` for `curl | sh` pinned installs.

**Files.** `.github/workflows/*.yml`, `LICENSE`, optional `scripts/smoke.sh`,
`install.sh`.
**Acceptance.** PR CI green on a clean checkout; tagging `v0.0.1` produces release
artifacts; smoke job catches an intentionally broken endpoint.
**Effort.** ~0.5–1 day (needs the API key as a CI secret + license choice).

---

## 5. Response caching + observability ✅ DONE

> Shipped: `internal/sectors/cache.go` (on-disk GET cache as a RoundTripper;
> TTL by endpoint class — ref ~24h / intraday ~1m / reports ~5m; `--no-cache`,
> `--cache-ttl`; `sectors cache path|clear`) and `internal/sectors/observe.go`
> (`-v/--verbose` request tracing to stderr w/ cache hit/miss + duration;
> `--dry-run` echoes the request, key redacted, no network). Transport chain in
> `cmdutil.buildTransport`: verbose → cache → retry → base; dry-run
> short-circuits. Tests in `internal/sectors/cache_test.go`. Verified live:
> 273ms miss → 1ms cache hit.
>
> Deferred sub-item: surfacing the API request-id header in the error envelope
> (would require threading response headers through ~63 call sites; verbose
> logging covers the debugging need for now).

**Goal.** Cut latency, token-gen cost, and rate-limit pressure for repetitive agent
reads; make agent runs debuggable.

**Why (agent-product).** Agents re-fetch slow-changing reference data (`list
subsectors`, `klse sectors`, company reports) relentlessly. Caching is near-free
reliability + speed.

**Design.**
- **Cache** as another `http.RoundTripper` (composes with the retry transport).
  On-disk under the config dir; key = method+URL+`Authorization` hash; TTL by
  endpoint class (helper-lists long, reports medium, transactions short, news
  short). Flags: `--cache-ttl`, `--no-cache`, and `sectors cache clear`.
- **Observability:** `--verbose` logs method/URL/status/duration to **stderr**
  (never stdout — keeps JSON clean); `--dry-run` prints the would-be request
  without sending; surface the API request-id header if present in errors.

**Files.** `internal/sectors/cache.go`, `cmd/cache.go`, wire transports in
`client.go`, flags in `root.go`, tests.
**Acceptance.** Second identical call within TTL is served from cache (verified via
`--verbose` timing / a hit counter); `--no-cache` bypasses; `cache clear` empties it.
**Effort.** ~1–1.5 days.

---

## Sequencing & milestones

| Milestone | Includes | Outcome |
|---|---|---|
| **M1 — Loadable & durable** | #1, #2 | Host agent can load the toolset; survives real traffic |
| **M2 — Cheap** | #3 | Every call is token/cost-lean |
| **M3 — Trustworthy** | #4 | Pinned releases + drift alarms |
| **M4 — Fast & debuggable** | #5 | Cache + tracing for production agent runs |

Cross-cutting decisions to settle first:
1. Manifest scope: all commands vs curated subset (affects #1).
2. License choice (affects #4).
3. Whether to also ship an **MCP server mode** (`sectors mcp`, stdio) reusing #1's
   flag→schema mapping — a hedge to meet agents that expect MCP. Small once #1 exists.

## Explicitly out of scope (note, don't silently skip)
- OAuth bearer flow (the spec defines `OAuthBearerAuth`; only API-key is wired) —
  add if the product needs per-end-user auth.
- Write/mutating endpoints — the API is read-only today.
- Per-endpoint typed response models — intentionally avoided; the CLI emits raw
  bodies and the upstream response schemas are unreliable.
