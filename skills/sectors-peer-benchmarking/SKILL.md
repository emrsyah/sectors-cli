---
name: sectors-peer-benchmarking
description: Compare and rank a set of companies (peers, a subsector, or an index) on multiple financial metrics using the `sectors` CLI — build scorecards, normalize metrics, and rank by a composite. Use when the user asks to compare/benchmark/rank several companies, score a peer group, find the best/cheapest/strongest among a set, or build a comparison table (e.g. "compare BBCA, BMRI, BBRI on P/E and market cap", "rank IDX banks by ROE"). Builds on sectors-cli.
---

# Peer benchmarking & scorecards

Comparing N companies on M metrics. The core procedure (and the hidden math) is
encoded in `scripts/benchmark.sh` — prefer it over re-deriving by hand.

## Why a script: the screener won't give you metrics

`sectors idx companies` returns only `symbol` + `company_name` — **not** the
metric you filtered/sorted on. So a comparison table always requires a
**fan-out**: screen to get the symbols, then fetch a `company report` per symbol
and pull the fields. `benchmark.sh` does exactly this (and the CLI's cache +
auto-retry make the fan-out fast and safe — no manual rate-limiting).

## Use the helper

```bash
# explicit peers, sorted by market cap
scripts/benchmark.sh --symbols BBCA,BMRI,BBRI \
  --metrics overview.market_cap,valuation.forward_pe --sort market_cap

# resolve the universe from a screener, normalize + composite-score
scripts/benchmark.sh --where "sub_sector='banks'" --top 8 \
  --metrics overview.market_cap,valuation.forward_pe --normalize --sort _score
```

Output is a JSON array (one row per company). `--metrics` are report JSON paths
(`section.field`); `--sections` is derived from them automatically. `--normalize`
adds min-max `_norm_*` columns in [0,1] and a `_score` = mean of available norms.
Discover field names with `sectors idx company report <sym> --sections valuation -o pretty`.

## The hidden math (when you must compute by hand)

- **TTM (trailing twelve months)** = sum of the **4 most recent quarters** for
  flow items (revenue, earnings, net interest income). Get them with
  `sectors idx company financials <sym> --n-quarters 4`.
- **Balance-sheet items** (total_assets, total_equity, total_deposit) = the
  **latest quarter only**, never summed.
- **Null-safe**: valuation/financial fields are often `null` — treat as missing
  (skip in averages), never as 0. `benchmark.sh` already does this.
- **Normalize before composing**: metrics have different scales — min-max or
  percentile-rank each to [0,1] before averaging into a score.
- **Direction matters**: for "lower is better" metrics (P/E, cost-to-income,
  debt) invert the normalized value (`1 - norm`) before scoring, or sort
  ascending. Don't blindly average raw norms across mixed-direction metrics.
- **Peer-average baselining**: a company's `valuation.historical_valuation[]`
  rows include `pe_peer_avg` / `pb_peer_avg` — use them to show "vs sector avg".

## Bank / financial-sector metrics

Banks expose extra fields under `financials_sector_metrics` in
`sectors idx company financials` (e.g. `net_interest_income`, `gross_loan`,
`total_deposit`). Compute ROE = TTM earnings ÷ latest total_equity, NIM ≈ TTM
net_interest_income ÷ latest total_assets. Don't expect these fields on non-banks.

## Tips

- Diversify a shortlist by de-duplicating on `sub_sector` / `industry` before
  ranking, so a "top 10" isn't all banks.
- For cross-market comparisons run the screener per market (IDX and SGX are
  separate); there is no unified cross-market screen.
