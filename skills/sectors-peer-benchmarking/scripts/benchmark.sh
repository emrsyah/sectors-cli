#!/usr/bin/env bash
#
# Build a peer-comparison table by fanning out `sectors company report` calls
# across a set of companies, extracting chosen metrics, and (optionally)
# normalizing + scoring them. The CLI handles auth, retries, and caching, so this
# script just orchestrates and reshapes — no rate-limit sleeps needed.
#
# Examples:
#   benchmark.sh --symbols BBCA,BMRI,BBRI --metrics overview.market_cap,valuation.forward_pe --sort market_cap
#   benchmark.sh --where "sub_sector='banks'" --top 8 --metrics overview.market_cap,valuation.forward_pe --normalize
#
# Flags:
#   --symbols a,b,c     explicit companies (bare symbols)
#   --where "<expr>"    OR resolve the universe from a screener filter
#   --top N             how many screener results to take (default 15)
#   --market idx|sgx    market group (default idx)
#   --metrics p1,p2     report JSON paths to extract, e.g. overview.market_cap,valuation.forward_pe (required)
#   --sections s        report sections to fetch (default: derived from --metrics)
#   --sort metric       sort by a metric column; prefix '-' for ascending (default: none)
#   --normalize         add min-max normalized columns (_norm_*) and a composite _score
#
# Requires: the `sectors` binary on PATH (or $SECTORS_BIN) and python3.
set -uo pipefail

BIN="${SECTORS_BIN:-sectors}"
PY=python3; echo | "$PY" -c 'pass' >/dev/null 2>&1 || PY=python

SYMBOLS="" WHERE="" TOP=15 METRICS="" SECTIONS="" SORT="" MARKET="idx" NORMALIZE=0
while [ $# -gt 0 ]; do
  case "$1" in
    --symbols) SYMBOLS="$2"; shift 2;;
    --where) WHERE="$2"; shift 2;;
    --top) TOP="$2"; shift 2;;
    --market) MARKET="$2"; shift 2;;
    --metrics) METRICS="$2"; shift 2;;
    --sections) SECTIONS="$2"; shift 2;;
    --sort) SORT="$2"; shift 2;;
    --normalize) NORMALIZE=1; shift;;
    -h|--help) sed -n '2,28p' "$0"; exit 0;;
    *) echo "benchmark: unknown arg '$1'" >&2; exit 2;;
  esac
done

[ -n "$METRICS" ] || { echo "benchmark: --metrics is required" >&2; exit 2; }

# Resolve the universe from a screener if explicit symbols weren't given.
if [ -z "$SYMBOLS" ]; then
  [ -n "$WHERE" ] || { echo "benchmark: provide --symbols or --where" >&2; exit 2; }
  SYMBOLS=$("$BIN" "$MARKET" companies --where "$WHERE" --max "$TOP" --select "results[].symbol" -o json \
    | "$PY" -c 'import json,sys;d=json.load(sys.stdin);print(",".join(r["symbol"].split(".")[0] for r in d.get("results",[]) if r.get("symbol")))')
  [ -n "$SYMBOLS" ] || { echo "benchmark: screener returned no companies for that filter" >&2; exit 1; }
fi

# Default sections = the top-level keys of the requested metric paths.
if [ -z "$SECTIONS" ]; then
  SECTIONS=$("$PY" -c 'import sys;print(",".join(sorted({m.split(".")[0] for m in sys.argv[1].split(",")})))' "$METRICS")
fi

tmp=$(mktemp); echo "[]" > "$tmp"
old_ifs=$IFS; IFS=','
for sym in $SYMBOLS; do
  IFS=$old_ifs
  sym=$(echo "$sym" | tr -d '[:space:]'); [ -n "$sym" ] || continue
  body=$("$BIN" "$MARKET" company report "$sym" --sections "$SECTIONS" -o json 2>/dev/null) \
    || { echo "benchmark: skipped $sym (request failed)" >&2; IFS=','; continue; }
  printf '%s' "$body" | "$PY" -c '
import json,sys
body=json.load(sys.stdin); sym=sys.argv[1]; metrics=sys.argv[2].split(","); f=sys.argv[3]
def dig(d,path):
    for k in path.split("."):
        d = d[k] if isinstance(d,dict) and k in d else None
        if d is None: return None
    return d
row={"symbol": body.get("symbol",sym)}
for m in metrics: row[m.split(".")[-1]] = dig(body,m)
tbl=json.load(open(f)); tbl.append(row); json.dump(tbl,open(f,"w"))
' "$sym" "$METRICS" "$tmp"
  IFS=','
done
IFS=$old_ifs

"$PY" -c '
import json,sys,math
tbl=json.load(open(sys.argv[1])); metrics=[m.split(".")[-1] for m in sys.argv[2].split(",")]
sort=sys.argv[3]; norm=sys.argv[4]=="1"
if norm:
    for m in metrics:
        vals=[r[m] for r in tbl if isinstance(r.get(m),(int,float))]
        if not vals: continue
        lo,hi=min(vals),max(vals)
        for r in tbl:
            v=r.get(m)
            r["_norm_"+m]=(0.0 if hi==lo else (v-lo)/(hi-lo)) if isinstance(v,(int,float)) else None
    for r in tbl:
        ns=[r["_norm_"+m] for m in metrics if isinstance(r.get("_norm_"+m),(int,float))]
        r["_score"]=sum(ns)/len(ns) if ns else None
if sort:
    desc = not sort.startswith("-"); key=sort.lstrip("-")
    def sk(r):
        v=r.get(key)
        return v if isinstance(v,(int,float)) else (-math.inf if desc else math.inf)
    tbl.sort(key=sk, reverse=desc)
json.dump(tbl,sys.stdout,indent=2); print()
' "$tmp" "$METRICS" "$SORT" "$NORMALIZE"
rm -f "$tmp"
