---
title: Kalendern: urval istället för dold serie
type: note
tags: [kalender, session, beslut, feature, projekt:plan]
created: 2026-11-19
updated: 2026-11-19
status: active
---

Användaren ville 'max kunna se 7 instanser av samma återkommande event'. Min tidigare implementation gömde HELA serien när den översteg gränsen - det var fel tolkning. Nu tunnas serien ut istället: högst N förekomster visas, spridda över fönstret, och varje rad märks med hur många förekomster den står för (×42).

**Varför inte dölja:** att gömma en hel serie tar bort information som användaren faktiskt vill ha (rutinens mönster). Urvalet visar mönstret OCH bevarar antalet.

**Viktig felsökning - planens 'fits'-kontroll:** jag använde layoutens 'clipped' (titelbandet ryms inte i vyn) för att avgöra om kolumnen är full. Det gav fel resultat: en förekomst vid fönstrets kant är 'clipped' utan att kolumnen är full - och då döljdes hela serien. Nu finns två begrepp: 'clipped' (radens band ryms inte - visas i '+N till') och 'overflow' (raden har TRÄNTS UT av staplingen - det är det som betyder att kolumnen är full). Planen kollar overflow.

**Följdfix:** rader vid fönstrets kant flyttas nu upp så att titelbandet syns (max ned till cursor, så att de inte överlappar).

**Övriga följder:**
- Inställningen heter nu 'maxInstances' (tidigare 'dailyVisibleLimit') och gäller alla serier, inte bara dagliga. Gamla sparade värden migreras vid läsning.
- Sammanfattningen märker en tunnad serie med 'urval' och får knappen 'Visa hela serien' (visar alla förekomster, de andra serierna döljs). Dolda serier har fortfarande 'Visa ensam'.
- Hysteresisen ändrades: en serie döljs inte längre 'tills antalet faller under gränsen' (den regeln finns inte längre). Istället: vid lika många förekomster döljs den som var dold sist, så att samma serie försvinner och kommer tillbaka.
