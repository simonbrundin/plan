---
title: Go-integrationstest tömde kalenderhändelserna i Neon igen
type: note
tags: [bug, incident, kalender, neon, test, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
related: [notes/kalenderhandelser-aterstallda-fran-neon-wal-historik, notes/parallella-agent-sessioner-i-samma-repo-och-databas]
---

2026-11-20 (16:11 UTC): Neon dev-branchen tömdes igen - den här gången försvann 17 av 18 kalenderhändelser (bara "Esters kalas" var kvar, skapad 16:11:18Z). Samma mönster som incidenten 13:42-13:43Z.

**Rotorsak hittad:** `src/go/internal/repository/calendar_event_repository_integration_test.go` (TestIconFlowIntegration) raderade ALLA händelser för första användaren i databasen i sin defer: `DELETE FROM calendar_events WHERE user_id=$1`. Testet väljer `SELECT id FROM users LIMIT 1` - alltså den riktiga användaren (id 1) när det körs mot dev-databasen i Neon. `calendar_event_places` tömdes via ON DELETE CASCADE. Goals, calendars och users var oskadade. Go API-loggen visar inga DELETE-anrop eftersom testet går direkt mot databasen, utanför API:t.

**Återställning:** 17 rader + 1 rad i calendar_event_places lästa read-only ur Neon time-travel (snapshot 2026-11-20T16:10:30Z, före tömningen) via `neon connection-string dev@<timestamp>` och COPY tillbaka med ursprungliga id:n, parent_event_id, calendar_id, tidsstämplar m.m. Verifierat: 18 rader i calendar_events, 1 rad i calendar_event_places.

**Fix:** testet raderar nu bara sina egna rader (`WHERE id = $1 OR parent_event_id = $1` per skapat id) och testhändelserna heter "[test] Löpning"/"[test] Fest" så att `scripts/cleanup-test-events.sh` kan städa dem om testet avbryts. Testet kördes efter fixen mot dev-databasen: PASS, 18 rader kvar, inga [test]-rader kvar.

**Regel:** kör aldrig Go-integrationstester mot dev-databasen i Neon med den riktiga sessionen - eller se till att testerna bara raderar rader de själva skapat. Kontrollera alltid `SELECT count(*) FROM calendar_events` före och efter.

Se [[notes/kalenderhandelser-aterstallda-fran-neon-wal-historik]] och [[notes/parallella-agent-sessioner-i-samma-repo-och-databas]].

## Relaterade

- [[notes/parallella-agent-sessioner-i-samma-repo-och-databas]]
- [[notes/kalenderhandelser-aterstallda-fran-neon-wal-historik]]
