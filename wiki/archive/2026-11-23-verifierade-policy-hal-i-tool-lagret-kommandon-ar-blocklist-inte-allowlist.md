---
title: Verifierade policy-hål i tool-lagret - kommandon är blocklist, inte allowlist
type: note
tags: [säkerhet, governance, tools, bug, projekt:agents]
---

Relaterat: notes/artikeln-om-agentic-software-factory-vad-som-ar-relevant-for-agentsystemet (punkt C - governance som deklarativ policy).

## Vad jag verifierade (2026-11-23)
Test (temp-test i internal/core/tools, sedan borttagen) anropade ValidateCommandStrict och ValidatePathStrict:

cmd="cat /etc/shadow"                    allowed=true
cmd="cat /home/simon/.ssh/id_rsa"         allowed=true
cmd="kubectl apply -f prod.yaml"          allowed=true
cmd="flux apply"                          allowed=true
cmd="helm upgrade --install foo ./chart"  allowed=true
cmd="git push --force origin main"        allowed=true
path="/home/simon/.ssh/id_rsa"            allowed=true
path="/home/simon/repos/agents/.git/config" allowed=true

## Tre hål
1. Bash kringgar path-valideringen. ValidatePathStrict används bara av read/write; bash.go anropar bara ValidateCommandStrict, och `bash -c` kör med användarens egna rättigheter (ingen sandbox).
2. Kommandon är **blocklist** (sanitize.go `CommandBlocklist`), inte allowlist. `sudo`/`kubectl`/`flux`/`helm`/`git push --force` går igenom.
3. Delar av `PathDenylist` är död kod: `/home/*/.ssh` matchar aldrig (glob expanderas inte), `/.git/objects` och `/.git/config` matchar aldrig för att `filepath.Clean` ger absolut sökväg (eller relativ utan leading slash). Bara `/etc/shadow` och `/etc/passwd` fungerar.

## Motproblem: blocklisten är för tight
Den blockerar `2>&1`, `$(...)` och `$VAR` – kommandon agenten faktiskt behöver. En agent som stoppas i onodan hittar workarounds (skript via write -> bash script.sh), och da gar den forbi valideringen av själva kommandot. Friktion skapar risk.

## Rekommenderad form av C (avgransad, inte en generell policy-motor)
- `policy.yaml` med allowlist på binärer (prefixmatch), `require_human`-mönster, miljöscope, paths.
- Enforcement i tool-lagret (`bash.go`/`tools.go`) – deterministiskt, inte prompt.
- **Två konsumenter av samma fil**: (1) enforcement i kod, (2) injiceras i system-prompten så agenten vet vad som är tillåtet. Annars bränner agenten turer på förbjudna kommandon.
- Byt ut hård block mot eskalering: `require_human`-listan returnerar "needs approval" → Slack/approvals → agenten fortsätter. Då kan man slappa på blocklisten och ändå ha kontroll.
- Önskar inte: eget DSL, arv, budget-motor i v1 – blir konfigdjungel utan skydd.
- Prioritering: A (auktoritetskedja i wikin) först, sedan C i denna avgränsade form (~150 rader + en YAML).

## Varför detta mäter
Det här är inte artikelns argument i abstrakt form – det är tre konkreta saker som gick att visa i kod, plus en förklaring av varför nuvarande blocklist faktiskt motverkar sitt eget mål.
