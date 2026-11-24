---
title: Barnmärke på mål - antal undermål visas överallt
type: note
tags: [feature, goals, nuxt, ui, projekt:plan]
---

Önskemål: överallt där ett mål visas ska det framgå hur många barn (undermål) det har.

Varför: målen är en trädstruktur, men i listor (/goals, /now, inbox, avslutade mål, kommandopaletten, beroendeväljaren) syntes bara titeln - man kunde inte se om ett mål var stort eller litet utan att öppna det.

Hur det löstes:
- Ny komponent src/nuxt/app/components/ChildCountBadge.vue. Visar en git-fork-ikon + antal undermål, döljs helt om målet saknar barn. Tooltip: 'N undermål (M klara)'.
- Ny getter childStats i src/nuxt/app/stores/goals.ts: Map<goalId, {total, finished}> räknat ur relationerna i målgraphen. Ingen extra serverfråga - grafen laddas redan i layouten, och addRelation/removeRelation i store-actions uppdaterar räknaren direkt.
- Inbyggt i: /goals, /now (Gör nu), målsidan (rubrik + barnrader + sökresultat + beroenden), /completed, inbox, kommandopaletten (suffix 'N undermål'), beroendeväljaren, QuickAdd-förälderväljaren samt de äldre komponenterna GoalHeader/GoalChildrenList/GoalParentSection/GoalChildSearch.

Ej gjort: VerticalCalendarns select för 'Kopplat mål' (native option kan inte ha badge) och GoalWeightVerticalEditor (lodräta segment, ingen plats).

Verifiering: bun run build + cucumber (89 scenarios) + headless chromium mot den byggda appen med dev-login-session (8 märken på /goals, 51 på /now med filtret 'Alla', badge i rubriken och på barnrader på målsidan, suffix i kommandopaletten).
