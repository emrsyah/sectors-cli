#!/usr/bin/env bash
#
# Live smoke test for sectors-cli. Exercises representative commands across every
# market against the real API and fails if any returns a non-2xx (the CLI exits
# non-zero with an error category on failure). Dynamic IDs (slugs, broker codes,
# WIUP codes) are discovered from list endpoints; dependents are skipped (not
# failed) if discovery yields nothing, so a transient empty list won't red the build.
#
# Usage:  SECTORS_API_KEY=... SECTORS_BIN=./sectors bash scripts/smoke.sh

set -uo pipefail

BIN="${SECTORS_BIN:-./sectors}"
: "${SECTORS_API_KEY:?set SECTORS_API_KEY}"

# Pick a Python that actually runs (Windows ships a non-functional `python3` stub).
PY=python3
echo | "$PY" -c 'pass' >/dev/null 2>&1 || PY=python

pass=0 fail=0 skip=0
run() {
  if "$BIN" "$@" -o json >/dev/null 2>/tmp/smoke_err; then
    pass=$((pass + 1)); printf 'ok    %s\n' "$*"
  else
    local code=$?
    fail=$((fail + 1)); printf 'FAIL  %s (exit %s) %s\n' "$*" "$code" "$(head -c 160 /tmp/smoke_err)"
  fi
}
# run_id is for lookups keyed by a discovered id. A 404 (exit 4) means the
# endpoint works but this particular id has no data for it — benign, so we count
# it as a pass. Any other non-2xx still fails.
run_id() {
  if "$BIN" "$@" -o json >/dev/null 2>/tmp/smoke_err; then
    pass=$((pass + 1)); printf 'ok    %s\n' "$*"
  else
    local code=$?
    if [ "$code" -eq 4 ]; then
      pass=$((pass + 1)); printf 'ok    %s (404 — no data for this id)\n' "$*"
    else
      fail=$((fail + 1)); printf 'FAIL  %s (exit %s) %s\n' "$*" "$code" "$(head -c 160 /tmp/smoke_err)"
    fi
  fi
}
skipped() { skip=$((skip + 1)); printf 'skip  %s (no discovered id)\n' "$*"; }

# pick(): read JSON on stdin, print the first of the given keys found in the
# first record of a list (handles both bare arrays and {results:[...]}).
pick() {
  "$PY" -c '
import json,sys
try: d=json.load(sys.stdin)
except Exception: print(""); sys.exit()
rows = d.get("results") if isinstance(d,dict) else d
rows = rows if isinstance(rows,list) else []
if rows and isinstance(rows[0],dict):
    for k in sys.argv[1:]:
        if k in rows[0] and rows[0][k]: print(rows[0][k]); sys.exit()
print("")' "$@" 2>/dev/null
}
firstkey() { "$PY" -c 'import json,sys
try: d=json.load(sys.stdin)
except Exception: print(""); sys.exit()
print(next(iter(d)) if isinstance(d,dict) and d else "")' 2>/dev/null; }

echo "== discovery =="
BROKER=$("$BIN" idx brokers registry -o json 2>/dev/null | pick broker_code code)
MINECO=$("$BIN" mining companies list --max 1 -o json 2>/dev/null | pick slug company_slug)
SITE=$("$BIN" mining sites list --max 1 -o json 2>/dev/null | pick slug site_slug)
WIUP=$("$BIN" mining auctions list --max 1 -o json 2>/dev/null | pick wiup_code code)
COMMO=$("$BIN" mining commodities list -o json 2>/dev/null | pick name commodity)
PROV=$("$BIN" mining reserves index -o json 2>/dev/null | firstkey)
: "${COMMO:=Coal}"
echo "broker=$BROKER mineco=$MINECO site=$SITE wiup=$WIUP commodity=$COMMO province=$PROV"

echo "== IDX =="
run idx companies --where "sub_sector='banks'" --max 2
run idx free-float --sub-sector banks
run idx company report BBCA --sections overview
run idx company segments TLKM
run idx company financials BBCA --n-quarters 1
run idx company corporate-actions BBCA
run idx company shareholders BBCA
run idx company ipo-performance BREN
run idx company quarterly-dates BBCA
run idx subsector report banks --sections statistics
run idx brokers registry
run idx brokers top
run idx brokers summary-top BBCA
run idx brokers foreign-flow BBCA
[ -n "$BROKER" ] && run idx brokers activity "$BROKER" || skipped idx brokers activity
[ -n "$BROKER" ] && run idx brokers activity-top "$BROKER" || skipped idx brokers activity-top
run idx brokers summary BBCA
run idx transaction idx-total
run idx transaction index-daily lq45
run idx ranking most-traded
run idx ranking top-changes --periods 1d
run idx news list --max 2
run idx news filings --max 2
run idx news suspensions --max 2
run idx list industries
run idx list subindustries
run idx list subsectors
run idx list tags
run idx list segments-companies

echo "== SGX =="
run sgx companies --where "sub_sector='banks'" --max 2
run sgx top --n-stock 2
run sgx report D05 --sections overview
run sgx sectors
run sgx tags
run sgx news --max 2
run sgx filings --max 2
run sgx buybacks --max 2
run sgx short-sell --max 2
run sgx daily D05

echo "== KLSE =="
run klse sectors
run klse companies --sector financials
run klse top --n-stock 2
run klse report 1155 --sections overview

echo "== MINING =="
run mining companies list --max 2
run mining commodities list
run mining commodities price "$COMMO"
run mining commodities exports --commodity-type "$COMMO" --year 2023
run mining commodities global --commodity-type Nickel
run mining sites list --max 2
run mining production total --commodity-type "$COMMO"
run mining reserves index
run mining licenses list --max 2
run mining auctions list --max 2
run mining contracts list
[ -n "$MINECO" ] && run_id mining companies get "$MINECO" || skipped mining companies get
[ -n "$MINECO" ] && run_id mining companies financials "$MINECO" || skipped mining companies financials
[ -n "$MINECO" ] && run_id mining companies ownership "$MINECO" || skipped mining companies ownership
[ -n "$MINECO" ] && run_id mining companies performance "$MINECO" || skipped mining companies performance
[ -n "$MINECO" ] && run_id mining commodities sales-destination "$MINECO" || skipped mining commodities sales-destination
[ -n "$SITE" ] && run_id mining sites get "$SITE" || skipped mining sites get
[ -n "$WIUP" ] && run_id mining auctions get "$WIUP" || skipped mining auctions get
[ -n "$PROV" ] && run_id mining reserves get "$PROV" || skipped mining reserves get

echo
echo "== summary: $pass passed, $fail failed, $skip skipped =="
[ "$fail" -eq 0 ]
