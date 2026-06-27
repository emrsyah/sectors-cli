#!/usr/bin/env bash
#
# Compact one-company snapshot: profile + valuation in a single small JSON object,
# optionally with recent quarterly financials. Curated `--select` paths keep the
# payload tiny (a full report is ~56 KB; this is a few hundred bytes).
#
# Usage:  snapshot.sh <symbol> [--market idx|sgx|klse] [--quarters N]
# Examples:
#   snapshot.sh BBCA
#   snapshot.sh BBCA --quarters 4
#   snapshot.sh D05 --market sgx
#
# Requires the `sectors` binary on PATH (or $SECTORS_BIN) and python3.
set -uo pipefail

BIN="${SECTORS_BIN:-sectors}"
PY=python3; echo | "$PY" -c 'pass' >/dev/null 2>&1 || PY=python

SYM="" MARKET="idx" QUARTERS=0
while [ $# -gt 0 ]; do
  case "$1" in
    --market) MARKET="$2"; shift 2;;
    --quarters) QUARTERS="$2"; shift 2;;
    -h|--help) sed -n '2,13p' "$0"; exit 0;;
    -*) echo "snapshot: unknown arg '$1'" >&2; exit 2;;
    *) SYM="$1"; shift;;
  esac
done
[ -n "$SYM" ] || { echo "usage: snapshot.sh <symbol> [--market idx|sgx|klse] [--quarters N]" >&2; exit 2; }

sel="symbol,company_name,overview.sector,overview.sub_sector,overview.market_cap,overview.last_close_price,overview.listing_date,valuation.forward_pe,valuation.intrinsic_value"
report=$("$BIN" "$MARKET" company report "$SYM" --sections overview,valuation --select "$sel" -o json) \
  || { echo "snapshot: report request failed for $SYM" >&2; exit 1; }

fin="null"
if [ "$QUARTERS" != "0" ] && [ "$MARKET" = "idx" ]; then
  fin=$("$BIN" idx company financials "$SYM" --n-quarters "$QUARTERS" -o json 2>/dev/null || echo "null")
fi

printf '%s\n%s' "$report" "$fin" | "$PY" -c '
import json,sys
parts=sys.stdin.read().split("\n",1)
report=json.loads(parts[0])
fin=json.loads(parts[1]) if len(parts)>1 and parts[1].strip() else None
out={"symbol": report.get("symbol"), "company_name": report.get("company_name")}
out.update(report.get("overview",{}))
out["valuation"]=report.get("valuation",{})
if fin is not None: out["recent_financials"]=fin
json.dump(out, sys.stdout, indent=2); print()
'
