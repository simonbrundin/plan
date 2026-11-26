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
)

# Normaliserar tidsstämplar och gör JSON stabilt att jämföra.
normalisera() {
  jq -S 'walk(
    if type == "string" and test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}")
    then (fromdateiso8601 // .)
    else . end
  )' 2>/dev/null
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

if [ "$fel" -ne 0 ]; then
  echo
  echo "Avvikelser hittade - se ovan."
  exit 1
fi

echo
echo "Alla anrop ger samma svar i båda tjänsterna."
