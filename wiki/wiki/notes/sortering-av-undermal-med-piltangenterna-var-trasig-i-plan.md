---
title: Sortering av undermål med piltangenterna var trasig i plan
type: note
tags: [bug, plan, nuxt, go, ui, projekt:plan]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Symtom: piltangenterna (ArrowUp/ArrowDown) som flyttar undermål fungerade ibland, ibland inte.

Rotorsaker (tre sammanfallande):
1. filteredChildren i src/nuxt/app/pages/goal/[id].vue sorterade om listan topologiskt utifrån beroenden. Den visade ordningen var alltså INTE order-fältet, så en flytt som sparade order inte syntes - bara när undermålen saknade beroenden verkade det fungera. Topologisk sortering är nu borttagen; listan följer order-fältet och beroenden visas bara som indikatorer (blockerad/väntar på).
2. moveChildUp/moveChildDown skickade två parallella PATCH:er och väntade på en hel GET-refresh. Snabba tangenttryck raceade och skrev över varandra (gammala order-värden). Nu uppdateras listan lokalt (optimistiskt) direkt och sparas via en ny endpoint som körs i kedjan och slås ihop.
3. Ordningen skapades som 'antal barn' vid nytt undermål, vilket gav dubbletter i order-fältet. Ett byte av två lika order-värden är en no-op. Nu används max(order)+1.

Ny endpoint: PATCH /api/v1/goals/relations/reorder {parentId, childIds[]} - skriver om alla ordningsnummer i en transaktion. Okända barn ignoreras, barn som saknas i requesten läggs sist (mergeChildOrder).

Tangentbevakningen ignorerar nu tangenter i formulärfält (viktslidern), annars flyttades listan när man justerade vikt med piltangenterna.
