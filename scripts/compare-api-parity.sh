#!/usr/bin/env bash
#
# Jämför Go-tjänsten och Rust-tjänsten mot samma databas.
#
# Under migreringen kör båda tjänsterna parallellt. Scriptet anropar samma
# vägar i båda och jämför svaren - så att en avvikelse i statuskod, JSON-fält
# eller felmeddelande upptäcks direkt, inte först i gränssnittet.
#
# Tidsstämplar normaliseras till epok innan jämförelsen: Go renderar dem i
# serverns tidszon, Rust i UTC. Det är samma tidpunkt, men olika text.
#
# Användning:
#   scripts/compare-api-parity.sh [go-url] [rust-url]
#
# Standard: http://127.0.0.1:8098 och http://127.0.0.1:8099.

set -uo pipefail

GO_URL="${1:-http://127.0.0.1:8098}"
RUST_URL="${2:-http://127.0.0.1:8099}"
TOKEN="${PARITY_TOKEN:-user_4}"

# Vägar som jämförs. Varje rad är "METOD VÄG".
ANROP=(
  "GET /health"
  "GET /api/v1/statuses"
  "GET /api/v1/goals"
  "GET /api/v1/goals/prioritized"
  "GET /api/v1/goals/graph"
  "GET /api/v1/goals/1"
  "GET /api/v1/goals/1/status"
  "GET /api/v1/goals/1/status/history"
  "GET /api/v1/places"
  "GET /api/v1/tags"
  "GET /api/v1/calendars"
  "GET /api/v1/calendar-events?scope=all"
  "GET /api/v1/calendar-events?from=2026-01-01T00:00:00Z&to=2027-01-01T00:00:00Z"
  "GET /api/v1/calendar-events?scope=xyz"
  "GET /api/v1/agent-runs"
  "GET /api/v1/agent-usage"
  "GET /api/v1/agent-budgets"
  "GET /api/v1/api-keys"
)

# Anrop som skickar en body. Endast felmeddelanden jämförs - de ändrar
# ingen data, så båda tjänsterna kan anropas med samma body.
ANROP_BODY=(
  "POST /api/v1/goals/dependencies {}"
  "POST /api/v1/goals/relations {}"
  "PATCH /api/v1/goals/relations {}"
  "DELETE /api/v1/goals/relations {}"
  "PATCH /api/v1/goals/relations/weights {}"
  "PATCH /api/v1/goals/relations/reorder {}"
  "PATCH /api/v1/goals/reorder {}"
  "PATCH /api/v1/goals/priority/reorder {}"
  "POST /api/v1/places {}"
  "PATCH /api/v1/places/1 {}"
  "PUT /api/v1/goals/1/places {}"
  "PUT /api/v1/calendar-events/1/places {}"
  "POST /api/v1/calendar-events {}"
  "PATCH /api/v1/calendar-events/1 {}"
  "POST /api/v1/calendar-events/1/occurrences {}"
  "POST /api/v1/calendar-events/1/split {}"
  "PUT /api/v1/me/current-place {}"
  "POST /api/v1/goals {}"
  "PATCH /api/v1/goals/1 {}"
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
