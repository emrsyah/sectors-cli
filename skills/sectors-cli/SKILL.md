---
name: sectors-cli
description: Foundation for the `sectors` command-line tool — Indonesian (IDX), Singapore (SGX), Malaysian (KLSE), and Indonesian mining market data (company financials, valuation, screeners, rankings, broker activity, news). Covers auth, command discovery, output/error handling, and token discipline. Use when the user asks about IDX/SGX/KLSE stocks, Indonesian or Southeast Asian equities, company fundamentals, stock screening, market movers, broker/foreign flows, IPOs, or mining data and the `sectors` binary is available — and as the base for the specialized sectors-* skills.
---

# Sectors financial data (`sectors` CLI)

`sectors` is a CLI over the Sectors Financial API v2. Run it with the shell and
parse its JSON. It is built for agents: JSON by default, non-zero exit + a JSON
error on failure, and `--help` on every command.

## Setup

Auth comes from `$SECTORS_API_KEY` (preferred) or `sectors auth login --api-key <key>`.
Check with `sectors auth status`. If you get `{"category":"auth"}` / exit 3, the
key is missing or invalid — tell the user; do not retry.

## Discover commands

Don't guess flags. The surface is four market groups — `idx`, `sgx`, `klse`,
`mining` — plus `auth`, `manifest`, `cache`.

```bash
sectors --help                 # top-level groups
sectors idx --help             # a group's commands
sectors idx companies --help   # one command's flags
sectors manifest               # machine-readable schema of every command
```

## The screener (most powerful tool)

`idx companies` / `sgx companies` filter and rank companies two ways:

```bash
# natural language (easiest; overrides every other flag)
sectors idx companies --q "top 5 growing banks by revenue in 2024"

# SQL-like filter + sort
sectors idx companies --where "sub_sector='banks'" --order-by "-market_cap" --max 10
```

Sector/tag values are kebab-case slugs — list valid ones with `sectors idx list
subsectors` (also `industries`, `subindustries`, `tags`).

## Save tokens — always

Responses can be large. Trim before reading:

- `--select <paths>` keep only fields, e.g. `--select overview.market_cap,valuation`
  or `--select "results[].symbol,results[].company_name"`.
- `--max N` cap a result list. `--count` return only the count.
- Reports support `--sections` (e.g. `--sections overview,valuation`) to fetch less.

Keep output as JSON (default). Use `-o table` **only** when showing a human.

## Reading results & errors

Success = exit 0 and a JSON body on stdout. Failure = non-zero exit and a JSON
envelope on stderr: `{"error","status","category","retryable"}`. Categories:
`invalid_input`(2) `auth`(3) `not_found`(4) `rate_limited`(5) `server`(6).
Transient errors are retried automatically — don't build your own retry loop.
A `404`/`not_found` usually means no data for that symbol/slug, not a bug.

## Gotchas that bite everyone

- **Symbols in, suffix out.** Pass bare symbols (`BBCA`, `GOTO`, SGX `D05`, KLSE
  `1155`); responses come back suffixed (`BBCA.JK`, `D05.SI`, `1155.KL`). Never
  pass the suffix.
- **Slugs are kebab-case** (`banks`, `oil-gas-coal`) — resolve real ones from
  `sectors idx list subsectors|industries|tags` (SGX/KLSE: `... sectors`/`tags`).
  A wrong slug returns an empty list, not an error.
- **`--q` overrides `--where`/`--order-by`/etc.** on the screener — never combine
  them. Empty results usually mean a too-narrow filter or stale date.
- **Screener rows are minimal** — `idx companies` returns only `symbol` +
  `company_name`, *not* the metric you filtered/sorted on. To get metrics, fetch
  a `company report` per symbol (see the sectors-screening / sectors-peer-benchmarking skills).
- **Time-series fields need a year bracket**: `revenue[2024]`, `-earnings[2023]`;
  bare `revenue` fails. Static metrics (`market_cap`, `pe_ttm`) take no bracket.
- **Dates** are `YYYY-MM-DD`; flow/price ranges cap to ~90 days (some intraday
  ≤14d); future dates → `invalid_input`. Inject today's date to resolve "last N days".

## Specialized skills

For deeper workflows, load the matching skill:
`sectors-screening` (find/rank), `sectors-company-analysis` (one-company deep
dive), `sectors-peer-benchmarking` (score N peers), `sectors-market-pulse`
(movers/flows/news), `sectors-mining` (commodity domain).

Full command map, query syntax, and recipes: see [REFERENCE.md](REFERENCE.md).
