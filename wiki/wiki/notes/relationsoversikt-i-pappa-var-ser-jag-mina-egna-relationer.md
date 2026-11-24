---
title: Relationsöversikt i pappa - var ser jag mina egna relationer
type: note
tags: [pappa, relationer, ui, navigering, fråga, projekt:pappa]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/redigerare-och-agarskap-for-relationer-i-pappa, notes/relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage]
---

2026-11-10: Efter att ägarskapet infördes kom frågan 'var ser jag alla mina (Michael Brundins) relationer?'. Svaret var: ingenstans - relationerna syndes bara per kandidat i dialogen och i detaljvyn. Ingen översikt.

LÖSNING: ny sida /relationer i menyn (Rekrytering, 💞). Filtren Mina / Gemensamma / Alla, sökning på person/organisation/typ, ägarlista. 'Redigerar som' i headern är samma useState-val som i dialogen, så identiteten är gemensam mellan sidorna.

BESLUT OM STARTFILTRET: sidan startar på 'Alla', inte 'Mina'. En tom lista när man kommer in är ett svagare första intryck, och 'Alla' visar att det finns något. Väljer man sig själv i listan hoppar vyn automatiskt till 'Mina' - det är det man nästan alltid är ute efter. Kopplad till frågan: en sida ska svara på den fråga användaren ställer, inte på den man själv tycker är intressant.

DJUPLÄNK: /kandidater?person=<id> öppnar kandidatens detaljvy direkt (via route.query i kandidater.vue). Behövdes för att relationer ska kunna länka till en kandidat. Detta mönster (query-param som öppnar en vy) är generellt och kan återanvändas av andra sidor.

FALLGROP: relationerna är symmetriska, så 'mina' är relationer ÄGDA av mig (paret ägs, inte riktningen) - det går inte att äga en relation åt ena hållet. Konsekvent med symmetrin, men värt att komma ihåg om man vill ha 'min relation till B' och 'B:s relation till mig' som olika saker.

DATASKEDE: de 39 befintliga relationerna är importerade ur YAML och saknar ägare - de är gemensamma. Michael Brundins egna lista är alltså tom tills några relationer läggs in i appen. Det är korrekt, inte en bugg.

## Relaterade

- [[notes/relationer-kan-nu-laggas-till-fran-oversikten-dialogen-har-par-lage]]
- [[notes/redigerare-och-agarskap-for-relationer-i-pappa]]
