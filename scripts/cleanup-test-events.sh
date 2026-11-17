#!/usr/bin/env bash
#
# Rensar ENDA testhändelser från dev-databasen.
#
# Testhändelser känns igen på att titeln börjar med TEST_TITLE_PREFIX.
# Händelser som INTE har prefixet lämnas orörda, oavsett hur gamla de är.
# Det här är den enda tillåtna vägen att städa bort testdata - radera aldrig
# "alla händelser i ett tidsintervall", det riskerar användarens data.
#
# Användning:
#   scripts/cleanup-test-events.sh            # visa vad som skulle raderas
#   scripts/cleanup-test-events.sh --force    # radera markerad testdata
#
set -euo pipefail

TITLE_PREFIX="${TEST_TITLE_PREFIX-\[test\]}"
FORCE=0

for arg in "$@"; do
  case "$arg" in
    --force) FORCE=1 ;;
    *) echo "Okänd flagga: $arg" >&2; exit 2 ;;
  esac
done

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$REPO_ROOT/environments/dev/db/neon.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Hittar inte $ENV_FILE" >&2
  exit 1
fi

# Sourcing fungerar inte (variablerna blir tomma), därför läses filen rakt av.
DATABASE_URL="$(grep -E '^DATABASE_URL=' "$ENV_FILE" | head -1 | cut -d= -f2- | tr -d '"')"

if [[ -z "$DATABASE_URL" ]]; then
  echo "Kunde inte läsa DATABASE_URL från $ENV_FILE" >&2
  exit 1
fi

# lib/pq (och Neon) stödjer inte channel_binding.
DATABASE_URL="${DATABASE_URL/&channel_binding=require/}"

PSQL=(psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -X -q -A -t)

echo "Markerad testdata (titel börjar med $TITLE_PREFIX):"
MARKED="$("${PSQL[@]}" -c "SELECT count(*) FROM calendar_events WHERE title LIKE '${TITLE_PREFIX}%'")"
echo "  $MARKED rader"

echo
echo "Användardata (lämnas orörd):"
UNMARKED="$("${PSQL[@]}" -c "SELECT count(*) FROM calendar_events WHERE title NOT LIKE '${TITLE_PREFIX}%'")"
echo "  $UNMARKED rader"

if [[ "$MARKED" == "0" ]]; then
  echo
  echo "Inget att rensa."
  exit 0
fi

echo
echo "Raderar:"
"${PSQL[@]}" -c "SELECT id || '  ' || title FROM calendar_events WHERE title LIKE '${TITLE_PREFIX}%'"

if [[ "$FORCE" != "1" ]]; then
  echo
  echo "Inget raderat (--force krävs)."
  exit 0
fi

DELETED="$("${PSQL[@]}" -c "DELETE FROM calendar_events WHERE title LIKE '${TITLE_PREFIX}%' RETURNING 1" | wc -l)"
echo
echo "Raderade $DELETED rader."
