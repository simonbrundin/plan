---
title: Kalendern begränsar återkommande förekomster i långtidvyer
type: note
tags: [kalender, feature, beslut, projekt:plan]
created: 2026-11-17
updated: 2026-11-17
status: superseded
related: [notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva, notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt]
---

**Superseded:** max-antal-iden är borttagen. Serier döljs inte längre beroende på zoomnivå - se [[notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva]].

Planens vertikala kalender visade tidigare högst 8 förekomster från samma återkommande serie i månads- och årsvyer. Enstaka händelser päverkades inte. Dolda förekomster indikerades med en statusrad och blev synliga igen när användaren zoomade in till dag-/veckonivå. Detta var en första prototyp av max-antal-idén; nästa steg blev serie-sammanfattning i stället för att bara dölja resten.

Varför prototypen togs bort: begränsningen kördes före grupperingen, så en tät serie tog alla platsen i de första månaderna och sedan saknades markörer helt i de senare. Serien syntes alltså beroende på hur mycket tid som rymdes i vyn, och antalet i etiketten var felaktigt.

## Relaterade

- [[notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva]]
- [[notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt]]
