---
title: Mål kan schemaläggas och kalenderhändelser flyttas
type: note
tags: [feature, kalender, mål, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Plan har fått drag-and-drop från mållistorna till kalendern. Släpp öppnar tidsalternativ med målets uppskattning först, och vald tid skapar en kalenderhändelse med goal_id, titel och ikon. Befintliga händelser kan flyttas med pekardragning; längden bevaras och återkommande förekomster sparas som overrides. Logiken ligger i src/nuxt/app/utils/calendarScheduling.ts och testas separat. Verifierat med Bun-enhetstester, 89 Cucumber-scenarier och Nuxt-produktionbygge.
