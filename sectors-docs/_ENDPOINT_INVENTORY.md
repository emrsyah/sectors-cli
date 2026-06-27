# Sectors Financial API v2 — Endpoint Inventory (for CLI command tree)

Base URL: `https://api.sectors.app` · Auth: header `Authorization: <api-key>` · All GET.
Source: extracted from local `sectors-docs/` v2 pages (4 sub-agents). ~57 live endpoints.

## Indonesia — Screener / Companies
| URL path | Path params | Key query params | CLI command |
|---|---|---|---|
| `/v2/companies/` | — | `where`, `q` (NL, overrides all), `order_by`, `desc`, `limit`(≤200), `offset`, `include_query_values` | `id companies screener` |
| `/v2/free-float/` | — | `sector`, `sub_sector`, `industry`, `sub_industry` | `id companies free-float` |

## Indonesia — Report
| URL path | Path params | Key query params | CLI command |
|---|---|---|---|
| `/v2/company/report/{symbol}/` | symbol | `sections`(overview,valuation,future,peers,financials,dividend,management,ownership) | `id company report <symbol>` |
| `/v2/company/get-segments/{symbol}/` | symbol | `financial_year` | `id company segments <symbol>` |
| `/v2/financials/quarterly/{symbol}/` | symbol | `report_date`, `approx`, `n_quarters` | `id company financials <symbol>` |
| `/v2/subsector/report/{sub_sector}/` | sub_sector | `sections`(statistics,market_cap,stability,valuation,growth,companies) | `id subsector report <slug>` |

## Indonesia — Company
| URL path | Path params | Query | CLI |
|---|---|---|---|
| `/v2/company/corporate-actions/{symbol}/` | symbol | — | `id company corporate-actions <symbol>` |
| `/v2/company/shareholders-composition/{symbol}/` | symbol | `year` | `id company shareholders <symbol>` |
| `/v2/listing-performance/{symbol}/` | symbol | — | `id company ipo-performance <symbol>` |

## Indonesia — Helper lists
| URL path | CLI |
|---|---|
| `/v2/companies/list_companies_with_segments/` | `id list segments-companies` |
| `/v2/company/get_quarterly_financial_dates/{symbol}/` | `id list quarterly-dates <symbol>` |
| `/v2/industries/` | `id list industries` |
| `/v2/subindustries/` | `id list subindustries` |
| `/v2/subsectors/` | `id list subsectors` |
| `/v2/tags/` | `id list tags` |

## Indonesia — Brokers
| URL path | Path params | Key query params | CLI |
|---|---|---|---|
| `/v2/broker-activity/{broker_code}/` | broker_code | `symbol`,`start`,`end` | `id brokers activity <code>` |
| `/v2/broker-activity/{broker_code}/top/` | broker_code | `start`,`end`,`n_brokers` | `id brokers activity-top <code>` |
| `/v2/broker-summary/{symbol}/` | symbol | `broker_code`,`start`,`end` | `id brokers summary <symbol>` |
| `/v2/broker-summary/{symbol}/top/` | symbol | `start`,`end`,`cohort`,`n_brokers`,`origin` | `id brokers summary-top <symbol>` |
| `/v2/brokers/` | — | `cohort`,`origin` | `id brokers registry` |
| `/v2/brokers/top/` | — | `cohort`,`date`,`metric`,`n_brokers`,`origin` | `id brokers top` |
| `/v2/foreign-flow/{symbol}/` | symbol | `start`,`end` | `id brokers foreign-flow <symbol>` |

## Indonesia — Transaction
| URL path | Path params | Query | CLI |
|---|---|---|---|
| `/v2/daily/{symbol}/` | symbol | `start`,`end` | `id transaction daily <symbol>` |
| `/v2/idx-total/` | — | `start`,`end` | `id transaction idx-total` |
| `/v2/index-daily/{index_code}/` | index_code | `start`,`end` | `id transaction index-daily <code>` |

## Indonesia — Ranking
| URL path | Key query params | CLI |
|---|---|---|
| `/v2/most-traded/` | `sub_sector`,`start`,`end`,`adjusted`,`n_stock` | `id ranking most-traded` |
| `/v2/companies/top-changes/` | `sub_sector`,`n_stock`,`classifications`,`periods`,`min_mcap_billion` | `id ranking top-changes` |

## Indonesia — News
| URL path | Key query params | CLI |
|---|---|---|
| `/v2/filings/` | `symbol`,`sector`,`sub_sector`,`start`,`end`,`limit`,`offset`,`transaction_type`,`tags`,`holder_type` | `id news filings` |
| `/v2/news/` | `extension`(idx/mining),`sector`,`sub_sector`,`commodity_type`,`keyword`,`symbols`,`tags`,`start`,`end`,`limit`,`offset` | `id news list` |
| `/v2/suspensions/` | `symbol`,`start`,`end`,`limit`,`offset` | `id news suspensions` |

## Malaysia / KLSE
| URL path | Path params | Key query params | CLI |
|---|---|---|---|
| `/v2/klse/companies/` | — | `sector`(req) | `my companies --sector <slug>` |
| `/v2/klse/companies/top/` | — | `sector`,`n_stock`,`classifications`,`min_mcap_million` | `my top-companies` |
| `/v2/klse/company/report/{symbol}/` | symbol | `sections`(overview,valuation,financials,dividend) | `my report <symbol>` |
| `/v2/klse/sectors/` | — | — | `my sectors` |

## Singapore / SGX
| URL path | Path params | Key query params | CLI |
|---|---|---|---|
| `/v2/sgx/companies/` | — | `where`,`q`,`order_by`,`desc`,`limit`,`offset`,`include_query_values` | `sg companies screener` |
| `/v2/sgx/companies/top/` | — | `sector`,`n_stock`,`classifications`,`min_mcap_million` | `sg top-companies` |
| `/v2/sgx/company/report/{symbol}/` | symbol | `sections` | `sg report <symbol>` |
| `/v2/sgx/sectors/` | — | — | `sg sectors` |
| `/v2/sgx/tags/` | — | — | `sg tags` |
| `/v2/sgx/news/` | — | `sector`,`sub_sector`,`start`,`end`,`limit`,`offset`,`tags`,`symbols` | `sg news` |
| `/v2/sgx/filings/` | — | `symbol`,`start`,`end`,`limit`,`offset`,`transaction_type`,`holder_type` | `sg filings` |
| `/v2/sgx/buybacks/` | — | `symbol`,`start`,`end`,`limit`,`offset` | `sg buybacks` |
| `/v2/sgx/short-sell/` | — | `symbol`,`start`,`end`,`limit`,`offset` | `sg short-sell` |
| `/v2/sgx/daily/{symbol}/` | symbol | `start`,`end` | `sg daily <symbol>` |

## Mining — Commodities-Trade
| URL path | Path params | Key query params | CLI |
|---|---|---|---|
| `/v2/mining/commodities/` | — | — | `mining commodities list` |
| `/v2/mining/commodities/{commodity_name}/price/` | commodity_name | `start_year`,`end_year` (≤3y) | `mining commodities price <name>` |
| `/v2/mining/exports/` | — | `commodity_type`(req),`year`(req),`limit` | `mining commodities exports` |
| `/v2/mining/global-commodity/` | — | `commodity_type`||`country`(≥1 req),`limit` | `mining commodities global` |
| `/v2/mining/sales-destination/{slug}/` | slug | `year` | `mining commodities sales-destination <slug>` |

## Mining — Companies
| URL path | Path params | Key query params | CLI |
|---|---|---|---|
| `/v2/mining/companies/` | — | `commodity_type`,`limit`,`offset`,`keyword`,`company_type`,`has_financials` | `mining companies list` |
| `/v2/mining/companies/{slug}/` | slug | — | `mining companies get <slug>` |
| `/v2/mining/companies/financials/{slug}/` | slug | `year` | `mining companies financials <slug>` |
| `/v2/mining/companies/ownership/{slug}/` | slug | — | `mining companies ownership <slug>` |
| `/v2/mining/companies/performance/{slug}/` | slug | `commodity_type`,`year` | `mining companies performance <slug>` |

## Mining — Licenses / Auctions
| URL path | Path params | Key query params | CLI |
|---|---|---|---|
| `/v2/mining/contracts/` | — | `contractor`,`mine_owner` | `mining contracts list` |
| `/v2/mining/license-auctions/` | — | `province`,`commodity_type`,`order_by`,`limit`,`offset`,`area_type`,`status`,`participant`,`qualified`,`min_participants` | `mining auctions list` |
| `/v2/mining/license-auctions/{wiup_code}/` | wiup_code | — | `mining auctions get <code>` |
| `/v2/mining/licenses/` | — | `province`,`commodity_type`,`company`,`order_by`,`limit`,`offset`,`expiring_soon`,`license_type`,`activity`,`cnc` | `mining licenses list` |

## Mining — Sites / Production
| URL path | Path params | Key query params | CLI |
|---|---|---|---|
| `/v2/mining/total-production/` | — | `commodity_type`(req) | `mining production total` |
| `/v2/mining/resources-reserves/` | — | — | `mining reserves index` |
| `/v2/mining/resources-reserves/{province}/` | province | `commodity_type`,`year` | `mining reserves get <province>` |
| `/v2/mining/sites/` | — | `province`,`commodity_type`,`company`,`year`,`order_by`,`min_production`,`limit`,`offset` | `mining sites list` |
| `/v2/mining/sites/{slug}/` | slug | — | `mining sites get <slug>` |

## Notes
- Screener `/v2/companies/` & `/v2/sgx/companies/`: `q` (natural language) overrides ALL other params. Without `q`: `where` (SQL-like) + `order_by`/`desc`/`limit`/`offset`. This is the killer endpoint for agents.
- Dates: `start`/`end` are YYYY-MM-DD; most ranges clamped to recent 90 days; future dates → 400.
- Common errors: 400 (bad request), 429 (rate limit). Need retry/backoff on 429.
- Path symbols given without `.JK`/`.SI`/`.KL` suffix (e.g. `BBCA`), responses return with suffix.
