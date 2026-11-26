#!/usr/bin/env bash
#
# Jämför Go-agenten och Rust-agenten mot samma databas.
#
# Samma princip som compare-api-parity.sh, men för agent-API:t: samma
# vägar, samma statuskoder och samma felmeddelanden i båda tjänsterna.
#
# Tidsstämplar normaliseras till epok - Go renderar dem i serverns
# tidszon, Rust i UTC. Det är samma tidpunkt, men olika text.
#
# Användning:
#   scripts/compare-agent-parity.sh [go-url] [rust-url]
#
# Standard: http://127.0.0.1:8198 och http://127.0.0.1:8199.

set -uo pipefail

GO_URL="${1:-http://127.0.0.1:8198}"
RUST_URL="${2:-http://127.0.0.1:8199}"
TOKEN="${PARITY_TOKEN:-testkey}"

# Vägar som jämförs. Varje rad är "METOD VÄG".
ANROP=(
  "GET /health"
  "GET /api/v1/health"
  "GET /api/v1/status"
  "GET /api/v1/conversations"
  "GET /api/v1/conversations/fel"
  "GET /api/v1/conversations/1/messages"
  "GET /api/v1/api-keys"
  "GET /api/v1/approvals"
  "GET /api/v1/approvals/540"
  "GET /api/v1/approvals/fel"
)

# Anrop som skickar en body. Endast felmeddelanden jämförs - de ändrar
# ingen data, så båda tjänsterna kan anropas med samma body.
ANROP_BODY=(
  "PATCH /api/v1/conversations/fel {\"title\":\"x\"}"
  "POST /api/v1/approvals/fel/decision {\"decision\":\"approve\"}"
  "POST /api/v1/approvals/540/decision {\"decision\":\"nonsense\"}"
  "POST /api/v1/approvals/540/decision {\"decision\":\"approve\"}"
  "POST /api/v1/api-keys {\"provider\":\"anthropic\",\"api_key\":\"hemligt\",\"model\":\"claude\"}"
  "DELETE /api/v1/api-keys/anthropic {}"
  "POST /api/v1/tools/cancel {\"toolCallId\":\"saknas\"}"
  "POST /api/v1/tools/cancel {}"
  "POST /api/v1/tools/cancel-all {}"
  "POST /api/v1/workflow {\"description\":\"x\"}"
  "POST /api/v1/workflow/saknas/approve {}"
  "POST /api/v1/workflow/saknas/reject {}"
  "POST /api/v1/workflow/decision {}"
)

# Normaliserar tidsstämplar och gör JSON stabilt att jämföra.
normalisera() {
  # Tidsstämplar skrivs olika i de två tjänsterna - Go använder
  # databasens tidszon, Rust UTC. Tiden görs därför om till epok, och
  # tidszonen räknas bort.
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

walk(normalisera_tid)
JQ
) 2>/dev/null
}

fel=0

for anrop in "${ANROP[@]}"; do
  metod="${anrop%% *}"
  vag="${anrop#* }"

  go_svar=$(curl -s -o /tmp/parity_go.json -w '%{http_code}' \
    -X "$metod" -H "Authorization: Bearer $TOKEN" "$GO_URL$vag")
  rust_svar=$(curl -s -o /tmp/parity_rust.json -w '%{http_code}' \
    -X "$metod" -H "Authorization: Bearer $TOKEN" "$RUST_URL$vag")

  if [ "$go_svar" != "$rust_svar" ]; then
    echo "STATUSKOD skiljer sig: $metod $vag (go=$go_svar, rust=$rust_svar)"
    fel=1
    continue
  fi

  go_json=$(normalisera < /tmp/parity_go.json)
  rust_json=$(normalisera < /tmp/parity_rust.json)

  if [ "$go_json" != "$rust_json" ]; then
    echo "BODY skiljer sig: $metod $vag (status $go_svar)"
    diff <(echo "$go_json") <(echo "$rust_json") | head -20
    fel=1
    continue
  fi

  echo "OK: $metod $vag ($go_svar)"
done

for anrop in "${ANROP_BODY[@]}"; do
  metod="${anrop%% *}"
  resten="${anrop#* }"
  vag="${resten%% *}"
  body="${resten#* }"

  go_svar=$(curl -s -o /tmp/parity_go.json -w '%{http_code}' \
    -X "$metod" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "$body" "$GO_URL$vag")
  rust_svar=$(curl -s -o /tmp/parity_rust.json -w '%{http_code}' \
    -X "$metod" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "$body" "$RUST_URL$vag")

  if [ "$go_svar" != "$rust_svar" ]; then
    echo "STATUSKOD skiljer sig: $metod $vag (go=$go_svar, rust=$rust_svar)"
    fel=1
    continue
  fi

  go_json=$(normalisera < /tmp/parity_go.json)
  rust_json=$(normalisera < /tmp/parity_rust.json)

  if [ "$go_json" != "$rust_json" ]; then
    echo "BODY skiljer sig: $metod $vag (status $go_svar)"
    diff <(echo "$go_json") <(echo "$rust_json") | head -20
    fel=1
    continue
  fi

  echo "OK: $metod $vag ($go_svar)"
done

# Dokumentationsvägarna jämförs på status, innehållstyp och Location -
# Scalar-sidan innehåller versionsbunden HTML som inte ska jämföras.
for vag in /docs /docs/ /docs/abc; do
  go_status=$(curl -s -o /dev/null -D /tmp/parity_go_head.txt -w '%{http_code}' "$GO_URL$vag")
  rust_status=$(curl -s -o /dev/null -D /tmp/parity_rust_head.txt -w '%{http_code}' "$RUST_URL$vag")

  # HTTP-headernamn är skiftlägesokänsliga - jämför dem i gemensam
  # storlek.
  go_plats=$(grep -i '^location' /tmp/parity_go_head.txt | tr -d '\r' | tr 'A-Z' 'a-z')
  rust_plats=$(grep -i '^location' /tmp/parity_rust_head.txt | tr -d '\r' | tr 'A-Z' 'a-z')

  if [ "$go_status" != "$rust_status" ] || [ "$go_plats" != "$rust_plats" ]; then
    echo "Dokumentationsvägen skiljer sig: $vag (go=$go_status $go_plats, rust=$rust_status $rust_plats)"
    fel=1
    continue
  fi

  echo "OK: GET $vag ($go_status)"
done

if [ "$fel" -ne 0 ]; then
  echo
  echo "Avvikelser hittade - se ovan."
  exit 1
fi

# WebSocket-chatten jämförs med ett eget verktyg - det kan inte göras
# med curl. Verktyget byggs om det saknas.
WS_PARITY="src/rust/target/debug/ws_parity"

if [ ! -x "$WS_PARITY" ]; then
  echo "Bygger ws_parity-verktyget..."
  (cd src/rust && cargo build --offline -p plan-parity --bin ws_parity) || exit 1
fi

if ! "$WS_PARITY" "$GO_URL" "$RUST_URL" "$TOKEN"; then
  fel=1
fi

if [ "$fel" -ne 0 ]; then
  echo
  echo "Avvikelser hittade - se ovan."
  exit 1
fi

echo
echo "Alla anrop ger samma svar i båda tjänsterna."
