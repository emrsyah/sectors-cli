#!/usr/bin/env bash
#
# Resolve a mining keyword to canonical slugs. Mining detail endpoints are keyed
# by kebab-case slug (company/site) or commodity name — you must look these up
# before calling `mining companies get`, `... financials`, `... sales-destination`, etc.
#
# Usage:
#   discover.sh companies <keyword>     # search mining companies -> slug, name, symbol
#   discover.sh commodities [keyword]   # list commodities (optionally filtered)
# Examples:
#   discover.sh companies adaro
#   discover.sh commodities coal
#
# Requires the `sectors` binary on PATH (or $SECTORS_BIN) and python3.
set -uo pipefail

BIN="${SECTORS_BIN:-sectors}"
PY=python3; echo | "$PY" -c 'pass' >/dev/null 2>&1 || PY=python

KIND="${1:-}"; KW="${2:-}"
case "$KIND" in
  companies)
    [ -n "$KW" ] || { echo "usage: discover.sh companies <keyword>" >&2; exit 2; }
    "$BIN" mining companies list --keyword "$KW" --max 20 \
      --select "results[].slug,results[].name,results[].symbol,results[].commodity_type" -o json \
      || { echo "discover: request failed" >&2; exit 1; }
    ;;
  commodities)
    "$BIN" mining commodities list -o json 2>/dev/null | "$PY" -c '
import json,sys
kw=(sys.argv[1].lower() if len(sys.argv)>1 else "")
d=json.load(sys.stdin)
rows = d.get("results") if isinstance(d,dict) else d
rows = rows if isinstance(rows,list) else []
out=[r for r in rows if not kw or kw in json.dumps(r).lower()]
json.dump(out, sys.stdout, indent=2); print()
' "$KW"
    ;;
  -h|--help|"")
    sed -n '2,15p' "$0";;
  *)
    echo "discover: unknown kind '$KIND' (want: companies | commodities)" >&2; exit 2;;
esac
