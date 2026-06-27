---
name: sectors-company-analysis
description: Deep-dive on a single company using the `sectors` CLI — profile, valuation, quarterly financials, revenue segments, dividends, corporate actions, ownership, and IPO performance across IDX/SGX/KLSE. Use when the user asks to analyze, research, or summarize one company, or asks about its financials, valuation, dividends, ownership, earnings, or "is X a good investment" (e.g. "analyze BBCA", "DBS valuation", "ADRO dividends"). Builds on sectors-cli.
---

# Single-company analysis

Everything about one ticker. The hero endpoint is `company report`, whose
`--sections` flag lets you fetch exactly what the question needs.

## Quick snapshot

For a compact profile + valuation (and optional recent financials) in one small
object, use the helper:

```bash
scripts/snapshot.sh BBCA              # profile + valuation
scripts/snapshot.sh BBCA --quarters 4 # + last 4 quarters of financials
scripts/snapshot.sh D05 --market sgx
```

## Pick sections by question (don't fetch the whole report)

`sectors idx company report <symbol> --sections <list>`:

| Question | sections |
|---|---|
| What is this company? | `overview` |
| Cheap/expensive? intrinsic value, P/E, peer avg | `valuation` |
| Profit/revenue/balance sheet history | `financials` |
| Dividend history & yield | `dividend` |
| Who owns it? | `ownership` |
| Analyst targets / forecasts | `future` |
| Comparable companies | `peers` |
| Management | `management` |

Default (no `--sections`) returns everything (~56 KB) — almost always pull only
what you need, and add `--select` to trim further.

## Other single-company endpoints (IDX)

```bash
sectors idx company financials BBCA --n-quarters 8   # quarterly income/balance items
sectors idx company segments TLKM                    # revenue/cost breakdown (Sankey-ready)
sectors idx company corporate-actions BBCA           # dividends, splits, rights issues
sectors idx company shareholders BBCA --year 2024    # monthly cap-table snapshots
sectors idx company ipo-performance BREN             # return since listing
```

## Quarterly financials: read the right fields

- Flow items (`revenue`, `earnings`, `operating_cash_flow`) are **per quarter** —
  sum 4 quarters for a trailing-twelve-month figure.
- Balance-sheet items (`total_assets`, `total_equity`) are point-in-time — use
  the latest quarter, don't sum.
- **Banks/insurers** carry extra fields under `financials_sector_metrics`
  (`net_interest_income`, `gross_loan`, `total_deposit`); non-financials won't.
- Need a specific quarter date? `sectors idx company quarterly-dates <symbol>`
  first, then pass `--report-date`. Or just use `--n-quarters N`.

## Notes

- SGX/KLSE have `report` only (4 sections: overview, valuation, financials,
  dividend) — no segments/financials-quarterly/corporate-actions/shareholders.
- To compare this company against peers, use the **sectors-peer-benchmarking** skill.
