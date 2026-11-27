#!/usr/bin/env bash
#
# Jämför Go-wikictl och Rust-wikictl.
#
# Samma princip som compare-api-parity.sh: samma kommandon, samma
# utdata och samma felmeddelanden i båda implementationerna.
#
# Läsande kommandon körs mot den riktiga wikin. Skrivande kommandon
# körs mot kopior av en testwiki, så att den riktiga wikin inte ändras.
#
# Användning:
#   scripts/compare-wiki-parity.sh [go-bin] [rust-bin] [wiki-dir]

set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GO_BIN="${1:-$REPO/src/agent/bin/wikictl}"
RUST_BIN="${2:-$REPO/src/rust/target/debug/wikictl}"
WIKI_DIR="${3:-$HOME/wiki}"

if [ ! -x "$GO_BIN" ]; then
  echo "Saknar Go-binären: $GO_BIN"
  echo "Bygg den med: cd $REPO/src && go build -o agent/bin/wikictl ./agent/cmd/wikictl"
  exit 1
fi

if [ ! -x "$RUST_BIN" ]; then
  echo "Saknar Rust-binären: $RUST_BIN"
  echo "Bygg den med: cd $REPO/src/rust && cargo build -p plan-wiki --bin wikictl"
  exit 1
fi

fel=0

# Jämför ett anrop i båda implementationerna.
jmför() {
  local beskrivning="$1"
  shift

  local go_utdata rust_utdata go_status rust_status

  go_utdata=$("$GO_BIN" --wiki "$WIKI_DIR" "$@" 2>&1)
  go_status=$?

  rust_utdata=$("$RUST_BIN" --wiki "$WIKI_DIR" "$@" 2>&1)
  rust_status=$?

  # Go skriver felmeddelanden med "fel: " i prefix, Rust gör det inte.
  go_utdata="${go_utdata#fel: }"
  rust_utdata="${rust_utdata#fel: }"

  if [ "$go_status" != "$rust_status" ]; then
    echo "EXITKOD skiljer sig: $beskrivning (go=$go_status, rust=$rust_status)"
    fel=1
    return
  fi

  if [ "$go_utdata" != "$rust_utdata" ]; then
    echo "UTDATA skiljer sig: $beskrivning"
    diff <(echo "$go_utdata") <(echo "$rust_utdata") | head -20
    fel=1
    return
  fi

  echo "OK: $beskrivning"
}

# Läsande kommandon mot den riktiga wikin.
jmför "list" list
jmför "list --project plan" list --project plan
jmför "list --project saknas" list --project saknas
jmför "search arkitektur" search arkitektur
jmför "search med tagg" search agent --tag agent
jmför "search med projekt" search rust --project plan
jmför "search med limit" search agent --limit 3
jmför "search saknar fråga" search
jmför "search okänt flagg" search agent --saknas
jmför "show concepts/authentik" show concepts/authentik
jmför "show saknas" show saknas
jmför "graph" graph
jmför "graph --mermaid" graph --mermaid
jmför "inbox" inbox
jmför "lint" lint
jmför "lint --stale-days 30" lint --stale-days 30
jmför "source list" source list
jmför "source add saknar fil" source add
jmför "okänt kommando" knas
jmför "help" help

# Skrivande kommandon körs mot kopior av en testwiki.
FIXTUR="$(mktemp -d)"
TESTWIKI="$FIXTUR/wiki"
trap 'rm -rf "$FIXTUR"' EXIT

mkdir -p "$TESTWIKI/wiki/concepts" "$TESTWIKI/wiki/notes" "$TESTWIKI/inbox" "$TESTWIKI/raw"

cat > "$TESTWIKI/wiki/concepts/Flux.md" <<'EOF'
---
type: concept
tags: [flux, gitops]
summary: En operator för GitOps.
---

# Flux

Flux synkroniserar Kubernetes.
EOF

cat > "$TESTWIKI/wiki/notes/arkitektur.md" <<'EOF'
---
title: Arkitektur
type: note
tags: [arkitektur, go]
---

Paketen är delade i lager.
EOF

cat > "$TESTWIKI/inbox/ny-not.md" <<'EOF'
---
title: Ny not
tags: [nytt]
---

Innehåll i noten.
EOF

echo "Råtext." > "$TESTWIKI/kalla.md"

# Kör ett skrivande kommando i båda kopierna och jämför resultaten.
jmför_skrivning() {
  local beskrivning="$1"
  shift

  local go_dir="$FIXTUR/go"
  local rust_dir="$FIXTUR/rust"

  rm -rf "$go_dir" "$rust_dir"
  cp -r "$TESTWIKI" "$go_dir"
  cp -r "$TESTWIKI" "$rust_dir"

  local go_utdata rust_utdata

  go_utdata=$("$GO_BIN" --wiki "$go_dir" "$@" 2>&1)
  rust_utdata=$("$RUST_BIN" --wiki "$rust_dir" "$@" 2>&1)

  go_utdata="${go_utdata//$go_dir/WIKI}"
  rust_utdata="${rust_utdata//$rust_dir/WIKI}"

  if [ "$go_utdata" != "$rust_utdata" ]; then
    echo "UTDATA skiljer sig: $beskrivning"
    diff <(echo "$go_utdata") <(echo "$rust_utdata") | head -20
    fel=1
    return
  fi

  # Jämför filträdet - samma filer med samma innehåll.
  local go_filer rust_filer

  go_filer=$(cd "$go_dir" && find . -type f | sort)
  rust_filer=$(cd "$rust_dir" && find . -type f | sort)

  if [ "$go_filer" != "$rust_filer" ]; then
    echo "FILTRÄD skiljer sig: $beskrivning"
    diff <(echo "$go_filer") <(echo "$rust_filer") | head -20
    fel=1
    return
  fi

  local fil

  while IFS= read -r fil; do
    if ! diff -q "$go_dir/$fil" "$rust_dir/$fil" > /dev/null 2>&1; then
      echo "FIL skiljer sig: $beskrivning -> $fil"
      diff "$go_dir/$fil" "$rust_dir/$fil" | head -20
      fel=1
      return
    fi
  done <<< "$go_filer"

  echo "OK: $beskrivning"
}

jmför_skrivning "note" note --title "En not" --tags "bug,session" --body "Vad som hände och varför." --project testwiki
jmför_skrivning "note utan titel" note --body "Text"
jmför_skrivning "page" page --type concept --title "Kyverno CEL" --body "CEL är standarden." --summary "Policies." --tags kyverno --project testwiki
jmför_skrivning "page utan body" page --type concept --title "Tom"
jmför_skrivning "page med ogiltig typ" page --type nonsens --title "X" --body "Y"
jmför_skrivning "link" link --from concepts/flux --to notes/arkitektur
jmför_skrivning "link till saknad sida" link --from concepts/flux --to notes/saknas
jmför_skrivning "unlink" link --from concepts/flux --to notes/arkitektur --remove
jmför_skrivning "source add" source add --file kalla.md --title "Min källa" --origin "https://example.com"
jmför_skrivning "ingest" ingest
jmför_skrivning "ingest --dry-run" ingest --dry-run
jmför_skrivning "index" index
jmför_skrivning "init" init .

if [ "$fel" -ne 0 ]; then
  echo
  echo "Avvikelser hittade - se ovan."
  exit 1
fi

echo
echo "Alla anrop ger samma svar i båda implementationerna."
