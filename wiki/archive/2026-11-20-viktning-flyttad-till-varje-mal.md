---
title: Viktning flyttad till varje mål
type: note
tags: [feature, goals, frontend, projekt:plan]
---

Efter feedback togs den separata viktpanelen bort. Viktningen visas nu som en liten balanserad slider direkt på varje undermålsrad i src/nuxt/app/components/GoalWeightSlider.vue. Mittlinjen betyder lika vikt; dragning uppdaterar lokalt alla relationer och sparar hela fördelningen atomiskt.
