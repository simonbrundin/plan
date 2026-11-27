---
title: Två agenter arbetade parallellt i infrastructure-repot 2026-11-27
type: note
tags: [lesson, projekt:infrastructure, workflow]
created: 2026-11-27
updated: 2026-11-27
status: active
---

Symtom: en parallell agents-session (startad 2026-11-26) var fortfarande aktiv 2026-11-27 och arbetade samtidigt i infrastructure-repot och i Zitadel. Den pushade commits (t.ex. 08c92659 'spara den importerade användarens inloggning i SOPS-secret') och satte om användarens lösenord medan jag felsökte samma konto.

Konsekvenser som behövdes hanteras:
- Två sessioner ändrade samma Zitadel-konto (lösenordsbyte 09:42, username-byte 09:29) - felaktiga slutsatser om vad som orsakade felen.
- Ett *.enc.yaml-secret lades till i repot men missades av .gitignore, så Flux-builden av komponenten zitadel failade ('user-credentials.enc.yaml: no such file or directory') tills filen adderades med 'git add -f'.
- Flux-synken av det nya secreten skedde först efter en manuell 'flux reconcile kustomization flux-system --with-source' (kontroll: flux --context admin@cluster1 ... - den OIDC-kontexten kan gå sönder om exec-pluginen försöker autentisera).

Lärdom: innan man börjar felsöka i ett repo, kontrollera om andra agenter är igång (ls -la ~/.pi/agent/sessions/<repo>/*.jsonl - filer som uppdateras nyligen), och undvik att pusha till samma branch samtidigt. Sätt alltid git add -f på nya *.enc.yaml-filer.
