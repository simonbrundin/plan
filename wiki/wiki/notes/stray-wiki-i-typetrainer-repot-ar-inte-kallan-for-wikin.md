---
title: Stray wiki/ i typetrainer-repot är inte källan för wikin
type: note
tags: [wiki, gotcha, typetrainer, verktyg, projekt:agents]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/kunskaps-wiki]
---

I typetrainer-repot finns en oversionerad katalog `wiki/` med en genererad `index.md` från 2026-10-22. Den är INTE wikin.

Varför: `~/wiki` är en symlänk till `~/repos/agents/wiki`, och det är där wikin versioneras. Filen i typetrainers `wiki/` är en gammal genererad kopia som hamnade där när `wikictl index` kördes från fel arbetskatalog. Den innehåller bara typetrainer-anteckningar och är föråldrad (den riktiga wikin har 115 noter).

Följd: committa inte `wiki/` i typetrainer. Den riktiga indexen genereras med `wikictl index` från `~/wiki`.

## Relaterade

- [[notes/kunskaps-wiki]]

