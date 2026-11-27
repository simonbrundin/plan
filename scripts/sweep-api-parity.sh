#!/usr/bin/env bash
#
# Går igenom alla GET-vägar i API:t och jämför Go och Rust.
#
# `compare-api-parity.sh` går igenom en handplockad lista av anrop.
# Det här scriptet tar listan från route-tabellen i Rust (`ROUTES` i
# `src/rust/crates/plan-api/src/routes/table.rs`) och jämför varje väg,
# så att en ny endpoint inte glöms bort i listan.
#
# Tidsstämplar och kalenderhändelsernas ordning normaliseras på samma sätt
# som i `compare-api-parity.sh`.
#
# Användning:
#   scripts/sweep-api-parity.sh [go-url] [rust-url]
#
# Standard: http://127.0.0.1:8098 och http://127.0.0.1:8099.

set -uo pipefail

GO_URL="${1:-http://127.0.0.1:8098}"
RUST_URL="${2:-http://127.0.0.1:8099}"
TOKEN="${PARITY_TOKEN:-user_4}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Normaliserar tidsstämplar och sorterar kalenderhändelser.
normalisera() {
  jq -S -f <(cat <<'JQ'
def tidszon:
  if (.tz // "Z") == "Z" then 0
  else (.tz
        | capture("(?<sgn>[+-])(?<h>[0-9]{2}):(?<m>[0-9]{2})") as $tz
        | (if $tz.sgn == "-" then -1 else 1 end)
          * (($tz.h | tonumber) * 3600 + ($tz.m | tonumber) * 60))
  end;

def normalisera_tid:
  if type == "string" and test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}")
  then (capture("^(?<stamp>[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2})(\\.(?<frac>[0-9]+))?(?<tz>Z|[+-][0-9]{2}:[0-9]{2})?$")
        | (.stamp + "Z") as $stamp
        | ($stamp | fromdateiso8601) as $bas
        | tidszon as $offset
        | $bas - $offset)
  else . end;

def sortera_händelser:
  if type == "array" and length > 0 and (.[0] | type == "object" and has("occurrence_id"))
  then sort_by(.id, .occurrence_id)
  else . end;

walk(normalisera_tid) | walk(sortera_händelser)
JQ
) 2>/dev/null
}

# Plockar alla GET-vägar ur route-tabellen.
vagar=$(grep -o '("GET", *"[^"]*")' \
  "$REPO_ROOT/src/rust/crates/plan-api/src/routes/table.rs" \
  | sed 's/("GET", *"//; s/")//' \
  | sed 's/{id}/1/g' \
  | sort -u)

fel=0
antal=0
for vag in $vagar; do
  antal=$((antal + 1))
  go_svar=$(curl -s -m 20 -o /tmp/sweep_go.json -w '%{http_code}' \
    -H "Authorization: Bearer $TOKEN" "$GO_URL$vag")
  rust_svar=$(curl -s -m 20 -o /tmp/sweep_rust.json -w '%{http_code}' \
    -H "Authorization: Bearer $TOKEN" "$RUST_URL$vag")

  if [ "$go_svar" != "$rust_svar" ]; then
    echo "STATUSKOD skiljer sig: GET $vag (go=$go_svar, rust=$rust_svar)"
    fel=1
    continue
  fi

  go_json=$(normalisera < /tmp/sweep_go.json)
  rust_json=$(normalisera < /tmp/sweep_rust.json)

  if [ "$go_json" != "$rust_json" ]; then
    echo "BODY skiljer sig: GET $vag (status $go_svar)"
    diff <(echo "$go_json") <(echo "$rust_json") | head -20
    fel=1
    continue
  fi

  echo "OK: GET $vag ($go_svar)"
done

echo
echo "$antal GET-vägar jämförda."

if [ "$fel" -ne 0 ]; then
  echo "Avvikelser hittade - se ovan."
  exit 1
fi

echo "Alla GET-vägar ger samma svar i båda tjänsterna."
