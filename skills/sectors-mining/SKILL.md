---
name: sectors-mining
description: Research the Indonesian mining sector with the `sectors` CLI — mining companies (financials, ownership, performance), commodity prices and exports, production and reserves, and mining licenses, auctions, and contracts (ESDM Minerba). Use when the user asks about Indonesian mining companies, coal/nickel/gold and other commodities, commodity prices or exports, mining production/reserves, or mining licenses/auctions (e.g. "coal price history", "Adaro financials", "nickel exports 2023", "mining auctions in Kalimantan"). Builds on sectors-cli.
---

# Mining sector

A separate data model from equities: keyed on **commodity name**, **province**,
and kebab-case **slugs** (company/site) — not tickers. Resolve slugs first.

## Discover (slugs are required for detail calls)

```bash
scripts/discover.sh companies adaro     # -> slug, name, symbol, commodity_type
scripts/discover.sh commodities coal    # -> commodity names + coverage
```

Or directly: `sectors mining companies list --keyword <kw>`,
`sectors mining commodities list`, `sectors mining sites list`,
`sectors mining reserves index` (provinces with data),
`sectors mining auctions list` (WIUP codes).

## Companies

```bash
sectors mining companies get <slug>          # profile: activities, licenses, contracts, site count
sectors mining companies financials <slug>   # annual assets/revenue/profit (USD millions)
sectors mining companies ownership <slug>    # parent/subsidiary tree with % stakes
sectors mining companies performance <slug>  # production, sales, strip ratio, reserves (by year)
```

## Commodities & trade

```bash
sectors mining commodities price Coal                                # monthly price history (≤3-yr range)
sectors mining commodities exports --commodity-type Coal --year 2023 # top export destinations (both required)
sectors mining commodities global --commodity-type Nickel            # global production/reserves/trade (need commodity OR country)
sectors mining commodities sales-destination <company-slug>          # a company's sales by country
```

## Production, reserves, licenses

```bash
sectors mining production total --commodity-type Coal     # national output by year + YoY (commodity required)
sectors mining reserves index                             # provinces/years/commodities with data
sectors mining reserves get "Kalimantan Timur"            # reserves detail (quote the province)
sectors mining sites list --commodity-type Coal --order-by -production_volume
sectors mining licenses list --license-type IUP           # IUP/IUPK licenses (ESDM Minerba)
sectors mining auctions list                              # license auctions; `auctions get <wiup_code>` for full record
sectors mining contracts list                             # mine-owner ↔ contractor links
```

## Vocabulary & gotchas

- **IUP / IUPK** = mining business licenses; **WIUP** = the licensed-area code
  (the id for `auctions get`); **ESDM Minerba** = the regulator/portal source.
- **Strip ratio** = overburden removed per unit of ore (lower = cheaper to mine).
- Commodity financials are **USD millions** (equities are IDR) — don't mix units.
- Required params: `exports` needs `--commodity-type` + `--year`; `production
  total` needs `--commodity-type`; `global` needs `--commodity-type` or `--country`.
- Province names are exact, capitalized, and space-separated (`"Kalimantan Timur"`),
  unlike equity slugs — quote them.
- A discovered company may lack a sub-resource (e.g. no performance data) → `404`
  / not_found / exit 4; that's data availability, not a bug.
