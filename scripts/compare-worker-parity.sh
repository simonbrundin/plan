#!/usr/bin/env bash
#
# Jämför Go-workern och Rust-workern mot samma Plan-API.
#
# Workern definierar mål. Scenarierna täcker de felvägar som inte syns
# i det vanliga flödet: trasigt modellsvar, ogiltig ikon och ogiltig
# upprepningsregel.
#
# Modellen ersätts av en mock (scripts/mock-llm.py) som svarar med ett
# bestämt svar. Då kan båda workerna köras mot samma "modell" och deras
# körningar, händelser och godkännanden jämföras.
#
# Användning:
#   scripts/compare-worker-parity.sh
#
# Miljövariabler:
#   PLAN_URL       Plan-API:t (standard http://127.0.0.1:8101)
#   PLAN_API_KEY   Nyckeln mot Plan (standard testkey)
#   GO_WORKER      Go-workerns binär
#   RUST_WORKER    Rust-workerns binär
#   MOCK_PORT      Porten mocken lyssnar på (standard 8299)
#   DB_URL         Databasen för städning

set -uo pipefail

# Bash känner bara igen ascii-tecken i variabelnamn - därför heter de
# här på engelska, medan kommentarerna och utskriften är på svenska.

PLAN_URL="${PLAN_URL:-http://127.0.0.1:8101}"
PLAN_KEY="${PLAN_API_KEY:-testkey}"
FULL_KEY="${FULL_KEY:-fullkey}"
GO_WORKER="${GO_WORKER:-/tmp/agent-parity/agent-worker-go}"
RUST_WORKER="${RUST_WORKER:-src/rust/target/debug/agent-worker}"
MOCK_PORT="${MOCK_PORT:-8299}"
WORKDIR="/tmp/worker-parity"

mkdir -p "$WORKDIR"

fel=0

# --- Scenarier -------------------------------------------------------
#
# Namn och det svar mocken ska ge. Svaren är skrivna som de modellen
# troligen skulle svara - då går samma felväg i båda implementationerna.

SCENARIER=(
  "giltig|{\"title\":\"Paritetsmål\",\"description\":\"Ett mål för paritetstest\",\"icon\":\"🎯\"}"
  "trasig-json|{\"title\": \"Trasig"
  "ogiltig-ikon|{\"title\":\"Paritetsmål\",\"description\":\"Ett mål\",\"icon\":\"inte-en-ikon\"}"
  "ogiltig-regel|{\"title\":\"Paritetsmål\",\"description\":\"Ett mål\",\"recurrence_rule\":\"FREQ=XYLOO\"}"
  "avslagen|{\"title\":\"Paritetsmål\",\"description\":\"Ett mål för paritetstest\",\"icon\":\"🎯\"}"
)

# --- Hjälpfunktioner -------------------------------------------------

# api anropar Plan-API:t med agentens nyckel.
api() {
  curl -s -H "X-API-Key: $PLAN_KEY" "$@"
}

# skapar ett testmål och skriver ut dess id.
#
# Målet skapas utan agentflaggan - den sätts sedan, eftersom
# skapandet inte tar emot flaggan (samma i Go och Rust).
skapa_mal() {
  local svar id
  svar=$(api -X POST "$PLAN_URL/api/v1/goals" \
    -H 'Content-Type: application/json' \
    -d '{"title":"Paritetsmål","status_id":1}')

  id=$(echo "$svar" | jq -r '.id // empty')

  if [ -z "$id" ]; then
    return
  fi

  api -X PATCH "$PLAN_URL/api/v1/goals/$id" \
    -H 'Content-Type: application/json' \
    -d '{"agent":true}' >/dev/null

  echo "$id"
}

# rensar bort allt som scenariot lämnat kvar.
stada() {
  local mal_id="$1"

  PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d plan_rust_test -q -c \
    "delete from agent_run_events where agent_run_id in (select id from agent_runs where goal_id = $mal_id);
     delete from agent_approvals where goal_id = $mal_id;
     delete from agent_run_usage where agent_run_id in (select id from agent_runs where goal_id = $mal_id);
     delete from agent_runs where goal_id = $mal_id;
     delete from user_goals where goal_id = $mal_id;
     delete from goal_tags where goal_id = $mal_id;
     delete from goal_places where goal_id = $mal_id;
     delete from goals where id = $mal_id;" >/dev/null 2>&1
}

# läser en ögonblicksbild av målets körning, händelser och godkännanden.
ogonblicksbild() {
  local mal_id="$1"
  local korningar kor_id handelser godkannanden

  korningar=$(api "$PLAN_URL/api/v1/agent-runs?goal_id=$mal_id")
  kor_id=$(echo "$korningar" | jq -r '.[0].id // empty')

  if [ -z "$kor_id" ]; then
    echo '{"runs":[],"events":[],"approvals":[]}'
    return
  fi

  handelser=$(api "$PLAN_URL/api/v1/agent-runs/$kor_id/events")
  godkannanden=$(api "$PLAN_URL/api/v1/agent-runs/$kor_id/approvals")

  # Mål- och körningsid skiljer sig mellan körningarna - de ersätts
  # med platshållare. Tidsstämplar tas bort av samma skäl.
  jq -n \
    --argjson korningar "$korningar" \
    --argjson handelser "$handelser" \
    --argjson godkannanden "$godkannanden" '
    def rensa:
      walk(if type == "object" then
        with_entries(
          if (.key | test("^(created_at|updated_at|started_at|finished_at|decided_at|requested_at)$"))
          then .value = "<tid>"
          elif .key == "id"
          then .value = "<id>"
          elif (.key | test("^(goal_id|parent_goal_id)$"))
          then .value = "<mål>"
          elif (.key | test("^(run_id|agent_run_id)$"))
          then .value = "<körning>"
          elif .key == "approval_id"
          then .value = "<godkännande>"
          else . end)
      else . end);

    {
      runs: ($korningar | rensa | map(del(.estimated_cost, .actual_cost, .cost, .usage))),
      events: ($handelser | rensa | sort_by(.type, .phase, .created_at)),
      approvals: ($godkannanden | rensa | sort_by(.kind))
    }'
}

# --- Körningarna -----------------------------------------------------

# avslagar definitionens förslag - beslutet tas av en människa, inte av
# agenten. Nyckeln nedan har full behörighet.
avslaga() {
  local mal_id="$1"
  local kor_id godkannande_id

  kor_id=$(curl -s -H "X-API-Key: $FULL_KEY" \
    "$PLAN_URL/api/v1/agent-runs?goal_id=$mal_id" | jq -r '.[0].id // empty')

  if [ -z "$kor_id" ]; then
    return 1
  fi

  godkannande_id=$(curl -s -H "X-API-Key: $FULL_KEY" \
    "$PLAN_URL/api/v1/agent-runs/$kor_id/approvals" | jq -r '.[0].id // empty')

  if [ -z "$godkannande_id" ]; then
    return 1
  fi

  curl -s -X POST "$PLAN_URL/api/v1/agent-approvals/$godkannande_id/decision" \
    -H "X-API-Key: $FULL_KEY" \
    -H 'Content-Type: application/json' \
    -d '{"decision":"reject","feedback":"Gör om så här"}' >/dev/null
}

# startar workern mot ett mål och väntar in definitionen.
#
# Tidsgränsen sätts per scenario - efter ett avslag behöver workern bara
# hinna se beslutet.
kor_workern() {
  local binar="$1" mal_id="$2" logg="$3" tidsgrans="${4:-25}"

  PLAN_API_URL="$PLAN_URL" \
  PLAN_API_KEY="$PLAN_KEY" \
  MINIMAX_API_KEY="" \
  MINIMAX_ENDPOINT="" \
  ANTHROPIC_API_KEY="" \
  OPENAI_API_KEY="mock" \
  OPENAI_BASE_URL="http://127.0.0.1:$MOCK_PORT/v1" \
  OPENAI_MODEL="mockmodell" \
  AGENT_GOAL_ID="$mal_id" \
  AGENT_POLL_INTERVAL="2s" \
  AGENT_STOP_WHEN_FINISHED="true" \
  AGENT_DATA_DIR="$WORKDIR" \
  timeout "$tidsgrans" "$binar" > "$logg" 2>&1
}

# --- Huvud -----------------------------------------------------------

if [ ! -x "$GO_WORKER" ]; then
  echo "Go-workern saknas: $GO_WORKER"
  exit 1
fi

if [ ! -x "$RUST_WORKER" ]; then
  echo "Bygger Rust-workern..."
  (cd src/rust && cargo build --offline -p plan-agent --bin agent-worker) || exit 1
fi

# Mocken läser svaret ur filen vid varje anrop - därför räcker det att
# skriva ett nytt svar mellan scenarierna. Miljövariabeln måste sättas
# innan mocken startar, så att den ärverkar den.
export MOCK_RESPONSE_FILE="$WORKDIR/svar.json"

python3 scripts/mock-llm.py "$MOCK_PORT" > "$WORKDIR/mock.log" 2>&1 &
MOCK_PID=$!
trap 'kill $MOCK_PID 2>/dev/null' EXIT

sleep 1

# En gammal mock kan ligga kvar och svara i stället för den nya - då
# jämförs fel svar. Skriptet ska inte köras vidare i det fallet.
if grep -q "Traceback" "$WORKDIR/mock.log"; then
  echo "Mocken kunde inte starta - porten är upptagen:"
  tail -3 "$WORKDIR/mock.log"
  exit 1
fi

# Kör ett scenario mot en workern och skriver ut ögonblicksbilden.
kör_scenario() {
  local namn="$1" binar="$2" logg="$3"
  local mal_id bild

  mal_id=$(skapa_mal)

  if [ -z "$mal_id" ]; then
    return 1
  fi

  kor_workern "$binar" "$mal_id" "$logg"

  # Ett avslaget förslag ska inte köras om - workern får köra igen
  # efter beslutet och då ska inget nytt hända.
  if [ "$namn" = "avslagen" ]; then
    avslaga "$mal_id"
    kor_workern "$binar" "$mal_id" "$logg.2" 8
  fi

  bild=$(ogonblicksbild "$mal_id")
  stada "$mal_id"

  echo "$bild"
}

for scenarie in "${SCENARIER[@]}"; do
  namn="${scenarie%%|*}"
  svar="${scenarie#*|}"

  printf '%s' "$svar" > "$WORKDIR/svar.json"

  echo "== $namn"

  # Go-workern.
  go_bild=$(kör_scenario "$namn" "$GO_WORKER" "$WORKDIR/go.log")

  if [ -z "$go_bild" ]; then
    echo "Kunde inte köra scenariot med Go-workern - hoppar över $namn"
    fel=1
    continue
  fi

  # Rust-workern.
  rust_bild=$(kör_scenario "$namn" "$RUST_WORKER" "$WORKDIR/rust.log")

  if [ "$go_bild" = "$rust_bild" ]; then
    echo "OK: $namn"

    if [ -n "${WORKER_PARITY_VERBOSE:-}" ]; then
      echo "$go_bild" | jq -S .
    fi

    continue
  fi

  fel=1
  echo "AVVIKELSE: $namn"
  diff <(echo "$go_bild" | jq -S .) <(echo "$rust_bild" | jq -S .) | head -40
done

if [ "$fel" -ne 0 ]; then
  echo
  echo "Avvikelser hittade - se ovan."
  exit 1
fi

echo
echo "Workarna beter sig likadant i alla scenarier."
