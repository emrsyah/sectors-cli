---
name: sectors-screening
description: Find and rank companies on the Indonesian (IDX) and Singapore (SGX) markets with the `sectors` CLI screener — by sector, financial metric, growth, or index membership, using natural-language or SQL-like filters. Use when the user wants to discover/screen/find/rank companies by criteria ("top N by X", "which stocks…", "cheapest/biggest/fastest-growing companies in <sector>", "companies in LQ45"). Builds on sectors-cli.
---

# Screening & ranking companies

`sectors idx companies` / `sectors sgx companies` is the discovery entry point.
It returns a list of `symbol` + `company_name` — the *starting point* for deeper
analysis, not the final answer.

## Two ways to query

```bash
# 1) Natural language — easiest; OVERRIDES every other flag when set
sectors idx companies --q "top 5 growing banks by revenue in 2024"

# 2) SQL-like filter + sort — precise and composable
sectors idx companies --where "sub_sector='banks'" --order-by "-market_cap" --max 10
```

Inspect what an NL query actually did via the response's `llm_translation`.

## `--where` / `--order-by` syntax (get this right)

- Operators: `= != > >= < <= like in`, combined with `and` / `or`.
- **Time-series fields need a year bracket**: `revenue[2024] > 1e12`,
  `order_by "-earnings[2023]"`. A bare `revenue` (no bracket) fails.
- **Static metrics** take no bracket: `market_cap`, `pe_ttm`, `dividend_yield`,
  `yoy_quarter_revenue_growth`.
- Arithmetic is allowed both sides and in `order_by`:
  `--where "revenue[2024]/total_assets[2024] > 0.1"`, `--order-by "-(earnings[2024]/earnings[2023])"`.
- List membership: `sub_sector in ['banks','insurance']`, `indices in ['lq45','idxbumn20']`.
- `--order-by` prefix `-` = descending; paginate with `--max`/`--offset`.

## Resolve slugs first (the #1 cause of empty results)

Sector/sub_sector/industry/tag values are kebab-case. Guessing fails silently.
Use the helper:

```bash
scripts/find-slug.sh bank                 # -> banks
scripts/find-slug.sh coal --type subsectors
scripts/find-slug.sh tech --market sgx --type sectors
```

(Or read the full lists: `sectors idx list subsectors|industries|subindustries|tags`.)

## The screen → enrich pattern

Screener rows are minimal. To rank/compare on the metric itself, fan out to
`company report` per symbol — that's the **sectors-peer-benchmarking** skill and
its `benchmark.sh`. Two common shapes:

```bash
# rank by a metric directly (no per-company calls): sort, then read top symbols
sectors idx companies --where "sub_sector='banks'" --order-by "-revenue[2024]" \
  --select "results[].symbol" --max 10

# index membership → constituents → drill in
sectors idx companies --where "indices in ['lq45']" --select "results[].symbol"
```

## Notes

- KLSE has **no** screener — use `sectors klse companies --sector <slug>` (sector
  is required) and `sectors klse top`.
- Empty list ≠ error: widen the filter or fix the slug/date.
