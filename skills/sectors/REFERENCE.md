# sectors CLI — reference

Authoritative surface is always `sectors <group> --help` and `sectors manifest`.
This file is a curated map + recipes. All commands output JSON unless `-o table`.

## Global flags (any command)

| flag | purpose |
|---|---|
| `--api-key` | override `$SECTORS_API_KEY` |
| `-o, --output` | `auto` (default) / `json` / `pretty` / `table` |
| `--select <paths>` | keep only these JSON paths (`a.b`, `results[].field`) |
| `--max N` | truncate the result list |
| `--count` | output only `{"count": n}` |
| `--retries`, `--retry-max-wait` | tune auto-retry (default 3) |
| `--no-cache`, `--cache-ttl` | control the on-disk response cache |
| `-v, --verbose` | log requests to stderr |
| `--dry-run` | print the request, don't send it |

## Command map

### `idx` — Indonesia Stock Exchange
- `companies` — screener: `--q`, `--where`, `--order-by`, `--desc`, `--limit`, `--offset`
- `free-float` — `--sector|--sub-sector|--industry|--sub-industry`
- `company report <symbol>` — `--sections overview,valuation,future,peers,financials,dividend,management,ownership`
- `company segments <symbol>` — `--financial-year`
- `company financials <symbol>` — `--report-date`, `--n-quarters`, `--approx`
- `company corporate-actions <symbol>`
- `company shareholders <symbol>` — `--year`
- `company ipo-performance <symbol>`
- `company quarterly-dates <symbol>`
- `subsector report <sub_sector>` — `--sections statistics,market_cap,stability,valuation,growth,companies`
- `brokers activity <broker_code>` | `activity-top <code>` | `summary <symbol>` | `summary-top <symbol>` | `registry` | `top` | `foreign-flow <symbol>`
- `transaction daily <symbol>` | `idx-total` | `index-daily <index_code>`  (e.g. `lq45`, `ihsg`)
- `ranking most-traded` | `top-changes` (`--periods 1d,7d,14d,30d,365d`)
- `news list` | `filings` | `suspensions`
- `list industries | subindustries | subsectors | tags | segments-companies`

### `sgx` — Singapore
`companies` (screener) · `top` · `report <symbol>` (e.g. `D05`) · `sectors` · `tags` · `news` · `filings` · `buybacks` · `short-sell` · `daily <symbol>`

### `klse` — Malaysia
`sectors` · `companies --sector <slug>` (required) · `top` · `report <symbol>` (4-digit, e.g. `1155`)

### `mining` — Indonesian mining extension
- `companies list|get <slug>|financials <slug>|ownership <slug>|performance <slug>`
- `commodities list|price <name>|exports|global|sales-destination <slug>`
- `sites list|get <slug>` · `production total` · `reserves index|get <province>`
- `licenses list` · `auctions list|get <wiup_code>` · `contracts list`

## Screener query syntax (`idx companies`, `sgx companies`)

- `--q "<plain English>"` — natural-language; when set it **overrides** `--where`,
  `--order-by`, etc. Easiest path for fuzzy requests.
- `--where "<expr>"` — operators `= != > >= < <= like in`, combine with `and`/`or`.
  Time-series fields use brackets: `revenue[2024] > 1e12`. List membership:
  `indices in ['lq45','idxbumn20']`.
- `--order-by "<field>"` — prefix `-` for descending, e.g. `-market_cap`,
  `-revenue[2024]`, `-yoy_quarter_revenue_growth`.
- `--limit` (≤200) / `--offset` for pagination.
- Result rows are minimal (symbol, company_name). For metrics, fetch a
  `company report` per symbol, or sort by the metric and read the ranking.

## Common parameters

- Dates: `--start` / `--end` are `YYYY-MM-DD`. Most ranges are capped to the most
  recent ~90 days; future dates return `invalid_input`.
- Slugs: sector/industry/tag values are kebab-case; get valid ones from the
  `list` / `sectors` / `tags` commands. Symbols are case-insensitive.

## Exit codes

`0` ok · `1` client/config/transport · `2` invalid_input · `3` auth ·
`4` not_found · `5` rate_limited · `6` server. Errors print JSON to stderr.

## Recipes

```bash
# Compare the largest Indonesian banks by market cap
sectors idx companies --where "sub_sector='banks'" --order-by "-market_cap" --max 10 \
  --select "results[].symbol,results[].company_name"

# Natural-language screen
sectors idx companies --q "profitable coal miners with dividend yield above 5%"

# One company's valuation snapshot (small payload)
sectors idx company report BBCA --sections overview,valuation

# Latest 4 quarters of financials
sectors idx company financials BBCA --n-quarters 4

# Today's top gainers and losers
sectors idx ranking top-changes --periods 1d --select "top_gainers,top_losers"

# Foreign-broker net flow on a stock
sectors idx brokers foreign-flow BBCA --start 2026-05-01 --end 2026-06-01

# SGX bank screen + a report
sectors sgx companies --q "Singapore banks by dividend yield" --max 5
sectors sgx report D05 --sections overview,dividend

# Mining: find a company, then its financials
sectors mining companies list --keyword adaro --select "results[].slug,results[].name"
sectors mining companies financials <slug>

# Coal commodity price history and top export destinations
sectors mining commodities price Coal
sectors mining commodities exports --commodity-type Coal --year 2023
```

## Tips

- Chain discovery → action: list slugs/symbols first, then query specifics.
- Pipe to `jq` for further extraction when needed.
- If a list comes back empty, it's usually a too-narrow filter or a stale date —
  widen the `--where` or adjust `--start`/`--end`, not a tool error.
