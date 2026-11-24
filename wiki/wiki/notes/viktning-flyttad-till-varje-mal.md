---
title: Viktning flyttad till varje mål
type: note
tags: [feature, goals, frontend, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Efter feedback togs den separata viktpanelen bort. Viktningen visas nu som en liten balanserad slider direkt på varje undermålsrad i src/nuxt/app/components/GoalWeightSlider.vue. Mittlinjen betyder lika vikt; dragning uppdaterar lokalt alla relationer och sparar hela fördelningen atomiskt.
