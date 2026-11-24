---
title: Kalenderhändelser kan tas bort och dupliceras
type: note
tags: [kalender, feature, ui, projekt:plan]
created: 2026-11-18
updated: 2026-11-18
status: active
---

Användaren kan nu ta bort och duplicera kalenderhändelser, och kalendrarna
ligger i en lista som öppnas med en knapp i stället för som etiketter i
toppen av tidslinjen.

**Varför:** etiketterna över tidslinjen tog plats och såg ut som klumpar.
Listan innehåller dessutom mer (synlighet, standardkalender, färg) än
etiketterna kunde visa.

**Nyckelinsikter om serier (viktigt att veta framöver):**

- Klienten expanderar serier lokalt, så en synad förekomst bär inte alltid
  seriens `recurrence_rule`. En förekomst med egen rad (override) har bara
  seriens `parent_event_id`. därför avgörs "är detta en förekomst?" av
  `occurrence_id`/`parent_event_id`, inte av regeln.
- Att stänga av en enskild förekomst görs med en override-rad vars status är
  `cancelled`. Finns det redan en rad för förekomsten måste den uppdateras
  (`PATCH`), inte skapas - unique indexet på (parent_event_id,
  original_start_at) gör att `INSERT ... ON CONFLICT DO NOTHING` tyst
  misslyckas.
- Att avsluta en serie görs med `UNTIL` i regeln. Ett befintligt `UNTIL` eller
  `COUNT` måste tas bort först, annars ignorerar rrule den nya gränsen.
- Override-rader för förekomster som faller utanför en avslutad serie måste
  raderas - annars ligger de kvar som egna händelser i vyn.

**Verifiering:** gjort i browsern mot dev-databasen (Neon) via den lokala
Nuxt-devappen och Go-API:t. Testhändelser skapades med titeln "[test] ..." och
städades med `scripts/cleanup-test-events.sh --force`.
