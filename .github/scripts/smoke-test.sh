#!/usr/bin/env bash
# Rökprov mot en deployad miljö.
#
# Körs efter en rollout - i previewen direkt efter att overlayen
# publicerats, i produktion efter att bot-PR:en med den nya imagen
# mergats.
#
# Argument:
#   --url URL        Grund-URL, t.ex. https://plan.simonbrundin.com
#   --mode preview   Preview: dev-login + databasberoende anrop
#   --mode prod      Produktion: kontrollerar att auth skyddar API:t
#   --expect SHA     Förväntad build-sha (från /api/health)
#   --wait SEKUNDER Hur länge vi väntar på att den nya versionen ska
#                    svara (default 300)
#
# Scriptet skriver ut en rad per kontroll och avslutar med 0 om allt
# gick bra, 1 annars. Det är avsiktligt tolerant: en kontroll som
# fallerar ska inte stoppa de övriga - en rapport som visar vad som
# är trasigt är mer värdefull än den som avbryts vid första felet.

set -uo pipefail

URL=""
MODE="prod"
EXPECT=""
WAIT=300

while [[ $# -gt 0 ]]; do
  case "$1" in
    --url) URL="$2"; shift 2 ;;
    --mode) MODE="$2"; shift 2 ;;
    --expect) EXPECT="$2"; shift 2 ;;
    --wait) WAIT="$2"; shift 2 ;;
    *) echo "Okänt argument: $1" >&2; exit 2 ;;
  esac
done

if [[ -z "$URL" ]]; then
  echo "--url saknas" >&2
  exit 2
fi

URL="${URL%/}"
FAILURES=0
SESSION_COOKIE=""

# Kollar en URL och förväntar sig en statuskod.
#   check "namn" "förväntad status" "curl-argument..."
check() {
  local name="$1"; shift
  local expected="$1"; shift

  local status
  status="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 20 "$@" 2>/dev/null || echo "000")"

  if [[ "$status" == "$expected" ]]; then
    printf '  ✓ %-42s %s\n' "$name" "$status"
  else
    printf '  ✗ %-42s %s (väntade %s)\n' "$name" "$status" "$expected"
    FAILURES=$((FAILURES + 1))
  fi
}

echo "Rökprov mot $URL (läge: $MODE)"

# Väntar tills /api/health svarar med den förväntade build-sha:n.
# Utan det går det inte att veta om den nya eller den gamla versionen
# svarar - och testet skulle bara bekräfta att något svarar.
wait_for_build() {
  if [[ -z "$EXPECT" ]]; then
    return 0
  fi

  echo "  … väntar på build $EXPECT (max ${WAIT}s)"
  local deadline=$((SECONDS + WAIT))

  while (( SECONDS < deadline )); do
    local body sha
    # /health är API:t:s egen kontroll - den rapporterar build-sha.
    # /api/health tillhör Nuxt och säger ingenting om builden.
    body="$(curl -sS --max-time 10 "${URL}/health" 2>/dev/null || echo "")"
    sha="$(printf '%s' "$body" | grep -o '"build_sha":"[^"]*"' | cut -d'"' -f4)"
    # Image-tagen använder kort sha, API:t rapporterar full - jämför
    # så att kort-sha är ett prefix av den rapporterade.
    if [[ -n "$sha" && "$sha" == "$EXPECT"* ]]; then
      printf '  ✓ %-42s %s\n' "bygg-versionen är ute" "${sha:0:12}"
      return 0
    fi
    sleep 10
  done

  printf '  ✗ %-42s %s (väntade %ss)\n' "bygg-versionen kom aldrig ut" "$EXPECT" "$WAIT"
  FAILURES=$((FAILURES + 1))
  return 1
}

wait_for_build

# Grundläggande: appen svarar, API:t svarar.
check "appen (root)" "200" "${URL}/"
check "nuxt /api/health" "200" "${URL}/api/health"
check "api /health" "200" "${URL}/health"

if [[ "$MODE" == "preview" ]]; then
  # Dev-login: en session skapas utan Zitadel. Sedan används sessionen
  # för ett anrop som går mot databasen - en 401 där betyder att
  # authen fungerar men att databasen inte svarar.
  login_status="$(curl -sS -o /dev/null -w '%{http_code}' -c /tmp/smoke-cookies.txt \
    -X POST --max-time 20 "${URL}/api/v1/auth/dev-login" 2>/dev/null || echo "000")"

  if [[ "$login_status" == "307" || "$login_status" == "200" || "$login_status" == "302" ]]; then
    printf '  ✓ %-42s %s\n' "dev-login" "$login_status"
    SESSION_COOKIE="$(awk '/plan_session/ {print $7}' /tmp/smoke-cookies.txt | head -1)"
  else
    printf '  ✗ %-42s %s\n' "dev-login" "$login_status"
    FAILURES=$((FAILURES + 1))
  fi

  if [[ -n "$SESSION_COOKIE" ]]; then
    check "databasberoende: /api/v1/goals" "200" \
      -H "Cookie: plan_session=${SESSION_COOKIE}" "${URL}/api/v1/goals"
    check "agenten: /api/agent/workflow/plans" "200" \
      -H "Cookie: plan_session=${SESSION_COOKIE}" "${URL}/api/agent/workflow/plans"
  fi
else
  # I produktion finns ingen dev-login. Authen ska i stället blockera
  # anrop utan session - en 401 visar att API:t lever och att authen
  # är på.
  check "auth: /api/v1/goals utan session" "401" "${URL}/api/v1/goals"
  check "auth: /api/v1/auth/session utan session" "401" "${URL}/api/v1/auth/session"
fi

rm -f /tmp/smoke-cookies.txt

if (( FAILURES > 0 )); then
  echo "MISSLYCKADES: $FAILURES kontroll(er) gick inte igenom"
  exit 1
fi

echo "OK: alla kontroller gick igenom"
exit 0
