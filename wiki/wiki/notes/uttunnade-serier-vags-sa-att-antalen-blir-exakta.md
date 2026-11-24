---
title: Uttunnade serier vägs så att antalen blir exakta
type: note
tags: [kalender, feature, beslut, prestanda, projekt:plan]
created: 2026-11-18
updated: 2026-11-18
status: active
related: [notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva]
---

En serie kan ha fler förekomster än expanderaren klarar (MAX_OCCURRENCES_PER_SERIES = 5000). Tidigare försvann då förekomster, vilket gav felaktiga antal i månads- och årsvyer.

Nu tunnas serien ut istället: en förekomst behålls per period (dag om antalet dagar ryms under capen, annars månad), och den behållna förekomsten får en vikt (occurrence_weight) som motsvarar hur många verkliga förekomster den står för. Alla antal - i tooltip, i markörens count och i seriens sammanfattning - summerar vikterna, så de är exakta även när förekomsterna inte visas var för sig.

Varför vikt och inte ett godtyckligt antal: antalet ska gå att lita på. Ett antal som är approximativt måste dessutom markeras som det, annars är det vilseledande.

Fält:
- occurrence_weight - hur många verkliga förekomster den visade förekomsten står för
- occurrence_total - det sanna antalet förekomster i fönstret
- occurrence_period - om serien tunnats ut per 'day' eller 'month', så urvalet kan beskrivas (visar

## Relaterade

- [[notes/aterkommande-serier-doljs-inte-langre-beroende-pa-zoomniva]]

