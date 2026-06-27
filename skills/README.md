# Sectors agent skills

Agent Skills that teach an AI agent (Claude Code, etc.) to drive the `sectors`
CLI for financial research. The agent shells out to `sectors` and parses JSON;
because the CLI already handles auth, retries, caching, pagination, error
categories, and token-shaping, these skills stay focused on *workflow and
analytical know-how* rather than HTTP plumbing.

## The suite

| Skill | Job | Helper script |
|---|---|---|
| **sectors-cli** | foundation: setup, discovery, output/errors, token discipline, universal gotchas | — |
| **sectors-screening** | find & rank companies (screener `--q` / `--where` / `--order-by`) | `find-slug.sh` |
| **sectors-company-analysis** | single-company deep dive (report sections, financials, dividends, ownership) | `snapshot.sh` |
| **sectors-peer-benchmarking** | compare/score N companies (TTM, normalization, composite) | `benchmark.sh` |
| **sectors-market-pulse** | movers, broker/foreign flows, news, filings | `flows.sh` |
| **sectors-mining** | commodity/mining domain (prices, exports, reserves, licenses) | `discover.sh` |

`sectors-cli` is the base; the others go deep on one analysis archetype and load
when the user's intent matches their description.

## Install

Copy the skill directories where your agent looks for skills, e.g. for Claude Code:

```bash
cp -r skills/sectors-* ~/.claude/skills/      # user-level
# or per-project:  cp -r skills/sectors-* .claude/skills/
```

The agent then loads the relevant skill automatically based on its `description`.

## Requirements

- The `sectors` binary on `PATH` (or set `SECTORS_BIN`), authenticated via
  `SECTORS_API_KEY` or `sectors auth login`.
- Helper scripts need `bash` and `python3` (they orchestrate `sectors` calls and
  reshape JSON). If a checkout loses the executable bit, restore it with
  `git update-index --chmod=+x skills/*/scripts/*.sh` (or run `bash <script>`).
