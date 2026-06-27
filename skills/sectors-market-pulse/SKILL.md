---
name: sectors-market-pulse
description: Track Indonesian (IDX) and Singapore (SGX) market activity with the `sectors` CLI — top gainers/losers, most-traded stocks, broker and foreign-investor flows, daily price trends, news, insider filings, and suspensions. Use when the user asks what's moving today, top movers/most-traded, who's buying/selling a stock, foreign or institutional flows, recent news/insider trades on a company, or price trends over a period. Builds on sectors-cli.
---

# Market pulse — movers, flows, news

Real-time-ish market monitoring. Most flow/transaction endpoints are **IDX-only**
and cap date ranges to ~90 days (some intraday ≤14 days); inject today's date to
resolve "last N days".

## Movers & activity

```bash
sectors idx ranking top-changes --periods 1d        # top gainers & losers (1d/7d/14d/30d/365d)
sectors idx ranking most-traded                     # most-traded by volume, keyed by date
sectors idx transaction daily BBCA --start <d> --end <d>   # OHLCV trend for one stock
sectors idx transaction idx-total                   # whole-market cap trend
sectors idx transaction index-daily lq45            # index level (lq45, ihsg, idx30)
```

## Smart-money flows (IDX)

Combine net foreign flow + the most active brokers for a stock in one view:

```bash
scripts/flows.sh BBCA --start 2026-05-01 --end 2026-06-01
```

Interpreting the parts:
- `foreign-flow` → `net_foreign_inflow` per day: **positive = foreign net buying**
  (often read as accumulation), negative = net selling.
- `brokers summary-top <symbol>` → `top_buyers` / `top_sellers` ranked by net IDR;
  filter cohorts with `--origin foreign|domestic` and `--cohort institutional|retail|...`.
- Pivot to a broker: `sectors idx brokers activity <broker_code>` shows everything
  that broker is trading; `sectors idx brokers registry` lists valid codes.

## News, filings, suspensions

```bash
sectors idx news list --keyword dividend --max 5     # market/mining news (--extension idx|mining)
sectors idx news filings --symbol BBCA               # insider/major-holder buy/sell filings
sectors idx news suspensions                         # trading suspensions + reasons
```

Filings filters: `--transaction-type buy|sell|others`,
`--holder-type insider|institution|corporate-investor`, `--sector`/`--sub-sector`,
`--tags`. SGX has `sgx news` / `sgx filings` / `sgx buybacks` / `sgx short-sell`.

## Notes

- Broker/foreign-flow analytics exist for **IDX only** (SGX offers only
  `short-sell` and `buybacks`; KLSE has none).
- `most-traded` and `top-changes` return company lists with the relevant figure —
  but for arbitrary fundamentals on the movers, fan out to `company report`.
- Dates `YYYY-MM-DD`; future dates → `invalid_input`; empty range usually means a
  stale/holiday date, not an error.
