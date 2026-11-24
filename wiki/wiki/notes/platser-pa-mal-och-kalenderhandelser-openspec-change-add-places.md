---
title: Platser på mål och kalenderhändelser (OpenSpec-change add-places)
type: note
tags: [feature, platser, goals, calendar, leaflet, bug, session, projekt:plan]
created: 2026-11-18
updated: 2026-11-20
status: active
related: [notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder]
---

Title: Platser på mål och kalenderhändelser (ny OpenSpec-change add-places)
Tags: feature, platser, goals, calendar, leaflet, session, projekt:plan

## Vad som byggdes

En plats är ett återanvändbart geografiskt område (namn, koordinat, radie i
meter) som tillhör en användare. Mål och kalenderhändelser kan kopplas till
flera platser via join-tabellerna `goal_places` och `calendar_event_places`.
Semantiken är OR: målet är görbart om användaren befinner sig på någon av
platserna.

- Sida `/platser` med Leaflet-karta (OSM-tiles, ingen API-nyckel), lista och
  CRUD. En plats skapas genom att klicka i kartan.
- Målsidan: pin-ikon + tangent `p` öppnar platsväljaren.
- Prioriteringslistan: "Jag är här" (dropdown eller GPS) + filtret
  "Görbar här". Filteret borde bara finnas i just den vyn enligt kravet.
- Kalendern: händelsedialogen har en platsväljare, händelseraderna visar
  pin-ikon.

## Varför vissa val gjordes så här

**Egen `places`-tabell i stället för koordinater på målet.** Samma plats
("Affären", "Hem") ska kunna kopplas till många mål utan att koordinaten
skrivs om, och en ändring av radien ska slå igenom överallt. Därför är
platsen en egen entitet med user_id, och målet refererar till den.

**`p` för plats, men bara på målsidan.** `p` var redan globalt bunden till
prioriteringslistan (`usePriorityToggle`). Eftersom platsväljaren bara
finns på målsidan ger tangenten olika betydelse beroende på sidan - det är
medvetet och dokumenterat i `openspec/changes/add-places/tasks.md`.

**"Jag är här" lagras server-side.** Valet ska följa mellan enheter, så det
ligger i `users.current_place_id`. GPS-positionen i sig sparas aldrig - bara
vilken sparad plats som matchar.

## Buggar som hittades under verifieringen (viktiga för framtiden)

1. **`errors.Is` mot wrappade fel.** `scanPlace` wrappar `pgx.ErrNoRows`, så
   `err == pgx.ErrNoRows` var falskt. Följd: "ingen vald plats" returnerade
   HTTP 500 i stället för `{"place": null}`.
2. **Leaflet + dynamisk import.** `await import("leaflet")` ger ett
   namespace utan `default` när Vite chunkar CJS-paketet - kartan renderades
   aldrig. Statisk `import L from "leaflet"` fungerar.
3. **Template-ref före `onMounted`.** Inuti `<ClientOnly>` sätts refen
   *efter* att komponenten mounted, så `onMounted` fick `null`. Kartan
   initieras därför från `watch(mapElement, ...)`.
4. **`step` på sifferfält.** Radiefältet hade `min=1` och `step=10`, vilket
   gjorde standardvärdet 100 ogiltigt. HTML-formuläret blockerade då
   submit utan felmeddelande - platsen kunde inte skapas.
5. **Constraint i fel tabell.** `goals_valid_recurrence_mode` låg i
   `goal_completions` (som refererar `goals.recurrence_mode`). Atlas
   kunde inte applicera schemat, vilket gjorde att `GET /goals` föll med
   "column g.recurrence_rule does not exist" i dev. Constrainten flyttades
   till `goals`-tabellen i både `environments/base/db/schema.yaml` och
   `environments/pr/template/db/schema.sql`.

## Verktygsknep

`bun test` (inte `npx vitest run`) kör de unit-testerna - testerna importerar
från `bun:test`, och vitest är inte installerat i `src/nuxt`. Cucumber körs
med `bun run test`. Webbläsarverifiering gjordes med playwright-core mot
`tilt`-porten 3100 och chromium från `~/.cache/ms-playwright`.

## Relaterade

- [[notes/malgraph-navigering-mellan-malsidor-gar-fran-sekunder-till-millisekunder]]

