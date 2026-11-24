---
title: Sortering av grundmål på /goals med piltangenterna
type: note
tags: [feature, ui, goals, go, nuxt, projekt:plan]
created: 2026-11-16
updated: 2026-11-23
status: active
related: [notes/prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko, notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan]
---

Grundmål på /goals kan nu sorteras med ArrowUp/ArrowDown, precis som undermålen
på en målsida. Ordningen sparas i databasen och följer med mellan sidladdningar.

Varför inte goal_relations: grundmål är mål UTAN förälder, så de har ingen
rad i goal_relations där orderfältet bor. Därför fick tabellen goals en egen
kolumn "order" (INTEGER NOT NULL DEFAULT 0), som bara används för grundmål.

Ny endpoint: PATCH /api/v1/goals/reorder {goalIds[]} - skriver om ordningen
för användarens grundmål i en transaktion. Endast mål utan förälder och utan
annan användare påverkas (nyckeln goalIds, inte childIds som på
relations/reorder).

Sorteringsregler i gränssnittet (app/utils/goalOrder.ts):
- Målen med lägst order först.
- Målen som delar order hamnar efter varandra med nyast först - det är
  samma ordning som API:t gav innan orderfältet fanns (created DESC), så
  befintlig data ser likadan ut tills användaren sorterar.
- Dolda mål (slutförda, eller dolda av sök/filter) behåller sin plats, bara
  det markerade målet och dess synliga granne byter plats - samma logik som
  moveChildId för undermål.

Övrigt:
- Nya grundmål får order = max(order)+1, alltså hamnar de sist.
- createGoal i repot sätter nu order vid INSERT (subselect), så att den
  nya raden inte får order 0.
- Kopplingen till /goals: listan filtreras med isRootGoal + isUnfinishedGoal
  och sorteras med sortGoalsByOrder.

## Relaterade

- [[notes/prioriteringsordningen-pa-now-sparades-inte-bulk-update-och-spar-ko]]
- [[notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan]]
