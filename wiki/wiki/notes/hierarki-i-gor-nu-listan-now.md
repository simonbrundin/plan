---
title: Hierarki i Gör nu-listan (/now)
type: note
tags: [feature, nuxt, prioritetssida, ui, projekt:simon]
created: 2026-11-23
updated: 2026-11-23
status: active
---

Mål i /now som har en förälder som också syns i /now visas nu indenterade under föräldern. Sorteringen (ArrowUp/ArrowDown) sker inom syskonskivan: ett barnmål kan bara byta plats med ett annat barn under samma förälder, medan föräldern sorteras bland toppnivåmålen.

Varför: den vanliga listan blandade föräldrar och barn i en platt ordning, så barn kunde hamna långt från sin förälder och flyttas ur sin grupp.

Hur det är löst: serverlistan (goals.priority_order) är fortfarande den enda källan för ordningen - inga nya kolumner, inga API-ändringar. Klienten bygger ett träd av föräldrekrelationerna (app/utils/nowHierarchy.ts) och grupperar barnen under föräldern när listan ritas. Sorteringen byter bara plats på två mål i den sparade listan (swapIds), så barnen ligger kvar under föräldern även efter en omsortering.

Viktiga detaljer:
- Ett mål läggs under den första föräldern som också finns i listan (föräldrarna kommer sorterade på vikt).
- Om föräldern filtreras bort (t.ex. filtret 'Nu') lyfts barnet upp till toppnivån - det syns då som ett vanligt mål och sorteras bland toppnivåmålen.
- j/k-navigeringen går i den visade ordningen (djup-först), inte i den platta listan.
- Toppnivåmålen numreras 1..n, barnmål har inget nummer men en ↳-markering.

Verifierat headless (lokal Postgres + Go API + byggd Nuxt-app): barn indenteras, ArrowUp på ett barn byter plats med syskonet men inte med föräldern, ArrowDown på föräldern flyttar föräldern bland toppnivåmålen (barnen följer med), och ArrowUp på det första barnet gör ingenting.
