# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project aims to
follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-06-27

Packaging fixes for `go install`. No functional changes to the CLI.

### Changed

- Moved the entrypoint to `cmd/sectors/`, so installing via Go produces a binary
  named `sectors` (not `sectors-cli`):
  `go install github.com/emrsyah/sectors-cli/cmd/sectors@latest`.

### Notes

- `0.1.0` is not installable through the public Go module proxy — the proxy
  cached a fetch failure from before the repository was public. Use `0.1.1`
  (or set `GOPROXY=direct GOSUMDB=off` for `0.1.0`).
- Agent skills (`skills/`) ship alongside the CLI; see `skills/README.md`.

## [0.1.0] - 2026-06-27

First public release. A complete, agent-facing CLI for the
[Sectors Financial API v2](https://docs.sectors.app) covering Indonesia (IDX),
Singapore (SGX), Malaysia (KLSE), and Indonesian mining data.

### Added

- **Full API coverage** — all 63 v2 endpoints across four markets, with a typed
  client generated from the OpenAPI spec via `oapi-codegen`:
  - `idx` — companies screener (`--where` / `--q` / `--order-by`), company
    reports, segments, financials, corporate actions, shareholders, IPO
    performance, brokers, transactions, rankings, news, and helper lists.
  - `sgx` — screener, reports, rankings, news, filings, buybacks, short-sell,
    daily prices.
  - `klse` — companies by sector, rankings, reports.
  - `mining` — companies, commodities & trade, sites & production, reserves,
    licenses, auctions, and contracts.
- **Agent integration**
  - `sectors manifest` — generates tool/function schemas for every command
    (`--format json|anthropic|openai`, `--filter`) so a host agent can load the
    CLI as a callable toolset.
  - Structured JSON error envelope with stable `category` + `retryable`, and
    category-based exit codes (2 invalid_input, 3 auth, 4 not_found,
    5 rate_limited, 6 server).
  - Automatic retries with exponential backoff + jitter (429/5xx/network),
    honoring `Retry-After` (`--retries`, `--retry-max-wait`).
  - Response shaping to save tokens: `--select` (field projection), `--max`
    (truncation), `--count`.
  - On-disk response cache with per-endpoint TTLs (`--no-cache`, `--cache-ttl`,
    `sectors cache path|clear`).
  - `--verbose` request tracing and `--dry-run` (echoes the request, key
    redacted, no network).
- **Output** — raw JSON by default (pretty on a TTY, compact when piped),
  plus `--output table` for list responses.
- **Auth** — `--api-key` flag, `SECTORS_API_KEY` env, or
  `sectors auth login`.
- Cross-platform release binaries (goreleaser) and shell completions.

[Unreleased]: https://github.com/emrsyah/sectors-cli/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/emrsyah/sectors-cli/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/emrsyah/sectors-cli/releases/tag/v0.1.0
