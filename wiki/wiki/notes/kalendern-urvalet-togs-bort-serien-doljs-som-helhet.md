---
title: Kalendern: urvalet togs bort - serien döljs som helhet
type: note
tags: [kalender, session, beslut, korrigering, projekt:plan]
created: 2026-11-19
updated: 2026-11-19
status: active
---

Korrigering av noten 'Kalendern: urval istället för dold serie'. Användaren tyckte inte om urvalet: 'nej det där blev jättekonstigt nu står det typ x3 till höger istället det vill jag inte. jag vill bara att eventen ska visas som vanligt tills det börjar visas 8 då vill jag alla liknande återkommande event döljs'.

**Den rätta modellen:** eventen visas som vanligt (inga ×N-märkningar, inget urval i tidslinjen). När det blir den åttonde instansen av en återkommande serie i vyn döljs HELA serien. Den ligger då kvar i sammanfattningen ovanför tidslinjen, märkt 'dold', med knappen 'Visa ensam'.

**Lärdom:** när användaren säger 'max 7 instanser' kan det betyda två helt olika saker: (a) visa 7 av 30 med etiketter om vad varje rad står för, eller (b) visa alla tills det blir för många, och dölj då hela serien. (b) är betydligt enklare och matchar hur kalendrar normalt fungerar. Min första tolkning (a) gav mer information men mer komplexitet och en synlig etikett '×42' i varje rad - det ville användaren inte ha.

**Kvarvarande bra förbättringar från samma session:** layoutens 'overflow' (rad trängd ut av stappling) skiljs från 'clipped' (titelbandet ryms inte) - annars gömdes hela serier för att en förekomst låg vid fönstrets kant. Rader vid kanten flyttas nu upp så att titelbandet syns.
