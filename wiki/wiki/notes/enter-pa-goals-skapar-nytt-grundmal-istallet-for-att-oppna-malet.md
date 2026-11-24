---
title: Enter på /goals skapar nytt grundmål istället för att öppna målet
type: note
tags: [ui, goals, feature, projekt:plan]
created: 2026-11-16
updated: 2026-11-17
status: active
related: [notes/eviga-mal-i-plan-kan-inte-slutforas, notes/goals-listar-bara-grundmal-scroll-vid-j-k, notes/sortering-av-grundmal-pa-goals-med-piltangenterna]
---

Ändring i src/nuxt/app/pages/goals/index.vue: tangenten Enter (och en + -knapp i sidhuvudet) öppnar nu en inline-rad där man skriver ett nytt grundmål (mål utan förälder). Enter sparar, Escape avbryter. Navigering till det markerade målet flyttades till tangenten 'l' (klick på raden navigerar fortfarande).

Varför: /goals-listan visar alla mål och saknade helt ett sätt att skapa ett grundmål därifrån; att trycka EnterNavigerade istället för att skapa. Ett grundmål är per appens modell ett mål utan parent-relationer (store-goals.ts getRootGoals).

FALLER: Go-API:t kräver icke-tom titel (CreateGoalRequest har binding:"required"), så man kan inte skapa målet med tom titel först och döpa det senare som man gör på /goal/[id]. Därför skapas målet först när titeln skrivs. Ett mål som skapas via API:t får inbox=1 (DB-default) - API:t kan inte sätta inbox, samma som befintligt beteende på målsidan.

## Relaterade

- [[notes/sortering-av-grundmal-pa-goals-med-piltangenterna]]
- [[notes/goals-listar-bara-grundmal-scroll-vid-j-k]]
- [[notes/eviga-mal-i-plan-kan-inte-slutforas]]
