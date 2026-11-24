---
title: En uttunnad förekomst måste vara synlig även om den ligger utanför fönstret
type: note
tags: [kalender, bugg, gotcha, projekt:plan]
created: 2026-11-18
updated: 2026-11-18
status: active
related: [notes/uttunnade-serier-vags-sa-att-antalen-blir-exakta]
---

När en serie tunnas ut behålls en förekomst per period, och den förekomsten ligger i periodens början. Om perioden går över fönstrets kant kan förekomsten ligga utanför fönstret.

Buggen: synligheten filtrerades med overlaps(), som bara kollar förekomstens egen start och slut. En förekomst som låg strax före fönstret kastades bort - och med den hela periodens vikt. Antalet i vyn blev då felaktigt, utan att något syntes som var fel.

Fixen: isVisible() använder förekomstens period (occurrence_period) istället för dess egen sträcka när vikten är större än 1. Då visas förekomsten så länge perioden överlappar fönstret, vilket är det som ska synas.

Varför detta är viktigt: hela poängen med viktade förekomster är att antalen ska vara exakta. Om vikten kan tappa bort på fönstrets kant så är antalen inte exakta, och då är vikterna värda mindre än de kostar.

Obs: rrule mellan-anrop är dyra (ett anrop över 240 månader tog över 5 minuter i en mätning). Räkna därför aldrig förekomster per period med upprepade between()-anrop - gruppera förekomsterna som redan expanderats i stället.

## Relaterade

- [[notes/uttunnade-serier-vags-sa-att-antalen-blir-exakta]]

