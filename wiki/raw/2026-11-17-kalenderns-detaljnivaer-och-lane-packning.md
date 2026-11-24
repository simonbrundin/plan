---
title: Kalenderns detaljnivåer och lane-packning
type: source
added: 2026-11-17
origin: https://github.com/simonbrundin/plan
immutable: true
---

Kalenderns detaljnivåer: hur mycket information en händelse får visa
beroende på hur mycket tid som syns i den vertikala tidslinjen.

## Varför

Tidslinjen kan zooma från timmar till flera år. En och samma händelse kan
inte visas likadant i alla de fallen: ett kort som visar "14:00 - 15:00"
är meningslöst i en årsvy och en prick per månad är värdlös i en timvy.

## Lösning: fem detaljnivåer

Nivån väljs utifrån tidsfönstrets storlek, inte utifrån hur "viktig"
händelsen är:

| Nivå       | Fönster         | Visning                          |
| ---------- | --------------- | -------------------------------- |
| `full`     | < 48 h          | kort med ikon, titel och klockslag |
| `compact`  | 48 h - 7 d      | kort med ikon och titel           |
| `daily`    | 7 d - 45 d      | prick per dag med antal           |
| `monthly`  | 45 d - 3 år     | prick per månad                   |
| `yearly`   | > 3 år          | prick per år                      |

Klick på en prick hoppar in en nivå: dag -> dygn, månad -> månad,
år -> år.

## Varför gränserna ser ut så de är

Gränserna är satta efter **hur många markörer som får plats i vyhöjden**,
inte efter kalendiska gränser. En prick per dag i ett årfönster skulle kräva
365 prickar, vilket varken går att läsa eller klicka på.

Det ledde till en viktig insikt: `scaleLevel` (hur tidslinjen skalas: timmar,
dagar, månader, år) och `detailLevel` (hur mycket en händelse visas) är två
**oberoende** koncept. Skalan har fortfarande kalendiska gränser, medan
detaljnivån har gränser satta efter presentation.

## Överlapp: lane-packning

På kompakt nivå kan flera händelser ligga nära varandra i tid, och på låg
nivå ligger prickarna tätt. Då överlappar de och går inte att klicka på.

Lösningen är **lane-packning**: händelser som överlappar i tid flyttas åt
sidan i kolumner, ungefär som en kalendervy gör. Kolumnerna delas bara
mellan de händelser som faktiskt överlappar (`assignClusteredLanes`), så att
en ensam händelse får hela bredden.

Viktigt: avståndet i packningen ska mäta i **pixlar**, inte i procent av
tidsfönstret. Ellers blir resultatet olika på mobil och desktop. Därför
följs `viewport.clientHeight` med en `ResizeObserver`.

En subtilitet som kostade tid: intervallen i packningen ska utökas med
kortets **höjd**, inte med avståndet mellan korten. Annars tror algoritmen
att kort som ligger en timma ifrån varandra överlappar, och de delas då
onödigt bredd.

## Gotchas

- Markörerna på låg nivå grupperas i **lokal tid**, inte UTC. Användaren
  ska känna igen "måndag 12 november", inte ett UTC-datum.
- `intensity` på prickarna är `0.45 + 0.55 * min(1, count / 3)`. Antalet
  visas bara i text när det är fler än ett - ett "1" säger ingenting.
- Ett enda gemensamt `icon`-fält används för både emoji och Iconify-namn.

## Verktyg

- `src/nuxt/app/utils/calendarDetail.ts` - ren logik för nivåer, gruppering
  och packning.
- `src/nuxt/features/calendar-detail.feature` - Cucumber-tester.
