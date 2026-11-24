---
title: Återkommande serier döljs inte längre beroende på zoomnivå
type: note
tags: [kalender, feature, beslut, bugg, session, projekt:plan]
created: 2026-11-17
updated: 2026-11-18
status: active
related: [notes/en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret]
---

Tidigare begränsades varje återkommande serie till 8 förekomster innan händelserna grupperades per månad/år (limitRecurringOccurrences). Följden blev att en tät serie (t.ex. daglig) tog alla 8 platsen i den första månaden och sedan försvann helt från de senare månaderna och åren - en serie syntes alltså beroende på hur mycket tid som rymdes i vyn, och antalet i tooltipen var felaktigt.

Nu gäller: en serie döljs aldrig för att zoomnivån är låg. Alla förekomster finns kvar och grupperas, så markörerna kommer tillbaka i varje månad/år och antalet i etiketten är sant (en serie skrivs ut en gång med sitt antal, t.ex. 'Löpning ×12'). Istället för en statusrad om dolda förekomster visas en sammanfattning av serierna ovanför tidslinjen, med seriens regel ('varje vecka'), antalet förekomster i vyn och en knapp som hoppar till nästa förekomst. Korten för återkommande förekomster märks med ↻.

Anledning: att tappa en serie ur tidslinjen beroende av zoomnivå är vilseledande - användaren kan inte veta om serien upphörde eller om den bara doldes. Tätheten hanteras med gruppering och etiketter, inte med bortval.

Övrigt: en serie med trasig upprepningsregel försvann tidigare helt ur kalendern (regeln kunde inte tolkas, så inga förekomster skapades). Den visas nu som en enstaka händelse istället.

## Relaterade

- [[notes/en-uttunnad-forekomst-maste-vara-synlig-aven-om-den-ligger-utanfor-fonstret]]

