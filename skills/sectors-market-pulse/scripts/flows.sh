#!/usr/bin/env bash
#
# Smart-money snapshot for one IDX stock: combines daily net foreign-broker flow
# with the top accumulating/distributing brokers into a single JSON object.
# (Broker/foreign-flow data is IDX-only.)
#
# Usage:  flows.sh <symbol> [--start YYYY-MM-DD] [--end YYYY-MM-DD]
# Example: flows.sh BBCA --start 2026-05-01 --end 2026-06-01
#
# Requires the `sectors` binary on PATH (or $SECTORS_BIN) and python3.
set -uo pipefail

BIN="${SECTORS_BIN:-sectors}"
PY=python3; echo | "$PY" -c 'pass' >/dev/null 2>&1 || PY=python

SYM="" START="" END=""
while [ $# -gt 0 ]; do
  case "$1" in
    --start) START="$2"; shift 2;;
    --end) END="$2"; shift 2;;
    -h|--help) sed -n '2,10p' "$0"; exit 0;;
    -*) echo "flows: unknown arg '$1'" >&2; exit 2;;
    *) SYM="$1"; shift;;
  esac
done
[ -n "$SYM" ] || { echo "usage: flows.sh <symbol> [--start ...] [--end ...]" >&2; exit 2; }

dargs=""; [ -n "$START" ] && dargs="$dargs --start $START"; [ -n "$END" ] && dargs="$dargs --end $END"

# shellcheck disable=SC2086
foreign=$("$BIN" idx brokers foreign-flow "$SYM" $dargs -o json 2>/dev/null || echo "null")
# shellcheck disable=SC2086
brokers=$("$BIN" idx brokers summary-top "$SYM" $dargs -o json 2>/dev/null || echo "null")

printf '%s\n%s' "$foreign" "$brokers" | "$PY" -c '
import json,sys
f,b=sys.stdin.read().split("\n",1)
foreign=json.loads(f) if f.strip() and f.strip()!="null" else None
brokers=json.loads(b) if b.strip() and b.strip()!="null" else None
out={"symbol": sys.argv[1], "foreign_flow": foreign, "top_brokers": brokers}
json.dump(out, sys.stdout, indent=2); print()
' "$SYM"
