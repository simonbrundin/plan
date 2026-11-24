---
title: Händelsen har ingen egen färg
type: note
tags: [kalender, färg, förenkling, projekt:plan]
---

Användaren bad om att ta bort möjligheten att sätta en egen färg på ett
visst event. Nu har varje händelse alltid kalenderns färg, och händelser
utan kalender har standardfärgen #6366f1.

## Vad som togs bort

- Kolumnen `calendar_events.color` (ur schemat och ur dev-databasen).
- `color` ur API-payloaden för händelser: modellerna
  CreateCalendarEventRequest, UpdateCalendarEventRequest,
  CreateCalendarOccurrenceRequest och SplitCalendarEventRequest har inte
  längre något color-fält, och `CalendarEvent` svarar inte längre med
  `color`.
- Färgvalet i händelseformuläret och knappen "Följ kalendern" i gränssnittet.
- `resolveEventColor` i src/nuxt/app/utils/calendarColors.ts tar bara kalendern
  och standardfärgen i sitt fall.

## Varför

Färgen är ett kalenderegenskap. När den även kunde sättas per händelse blev
det två ställen att hålla färgen och en knapp ("Följ kalendern") som bara
fanns för att avbryta en override. Med en enda regel - händelsens färg är
kalenderns färg - följer färgen med när kalendern byts, och det finns inget
som kan glida isär.

## Gotchas

- **Tre-state-semantik försvann med färgen.** Tom sträng ("töm färgen") och
  nil ("behåll") gav mening när färgen kunde nollställas. Nu finns bara
  kalendern, så inget att tripla-stat:a.
- **SQL-parameternumreringen ändrades** i Update och Create i
  calendar_event_repository.go när color togs bort ($13/$14 flyttades).
  Fel numrering ger "bind message supplies 13 parameters, but prepared
  statement requires 14".
- **Dev-databasens kolumn försvann redan** innan ALTER TABLE kördes - någon
  annan session hade redan kört en schemamigrering. Kontrollera alltid
  \d calendar_events innan du migrerar manuellt.
