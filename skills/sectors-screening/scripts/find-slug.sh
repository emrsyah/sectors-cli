#!/usr/bin/env bash
#
# Resolve a fuzzy term (e.g. "bank", "coal", "tech") to the exact kebab-case
# slugs the screener's `where`/`--sector` filters require. Wrong slugs return an
# empty list silently, so always resolve first.
#
# Usage:
#   find-slug.sh <term> [--type subsectors|industries|subindustries|tags] [--market idx|sgx|klse]
# Examples:
#   find-slug.sh bank                      # -> banks
#   find-slug.sh coal --type subsectors    # -> oil-gas-coal, ...
#   find-slug.sh tech --market sgx --type sectors
#
# Requires the `sectors` binary on PATH (or $SECTORS_BIN) and python3.
set -uo pipefail

BIN="${SECTORS_BIN:-sectors}"
PY=python3; echo | "$PY" -c 'pass' >/dev/null 2>&1 || PY=python

TERM="" MARKET="idx" TYPE="subsectors"
while [ $# -gt 0 ]; do
  case "$1" in
    --type) TYPE="$2"; shift 2;;
    --market) MARKET="$2"; shift 2;;
    -h|--help) sed -n '2,15p' "$0"; exit 0;;
    -*) echo "find-slug: unknown arg '$1'" >&2; exit 2;;
    *) TERM="$1"; shift;;
  esac
done
[ -n "$TERM" ] || { echo "usage: find-slug.sh <term> [--type ...] [--market ...]" >&2; exit 2; }

out=$(
  case "$MARKET" in
    idx)  "$BIN" idx list "$TYPE" -o json 2>/dev/null;;
    sgx)  "$BIN" sgx "$TYPE" -o json 2>/dev/null;;   # sgx sectors | tags
    klse) "$BIN" klse sectors -o json 2>/dev/null;;
    *) echo "find-slug: unknown market '$MARKET'" >&2; exit 2;;
  esac
)
[ -n "$out" ] || { echo "find-slug: could not fetch $MARKET $TYPE" >&2; exit 1; }

printf '%s' "$out" | "$PY" -c '
import json,sys
term=sys.argv[1].lower()
def walk(x, out):
    if isinstance(x,str): out.add(x)
    elif isinstance(x,dict):
        for v in x.values(): walk(v,out)
    elif isinstance(x,list):
        for v in x: walk(v,out)
vals=set(); walk(json.load(sys.stdin), vals)
hits=sorted(v for v in vals if term in v.lower())
print("\n".join(hits) if hits else "(no %s slug matches \"%s\")"%(sys.argv[2], sys.argv[1]))
' "$TERM" "$TYPE"
