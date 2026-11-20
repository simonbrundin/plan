# Dra mål till kalendern och flytta händelser

## Problem

Ett mål i Plan är en intention - "springa i parken", "skriva rapporten". Ett
kalenderevent är något annat: en händelse som har en starttid, en sluttid och
en plats i tiden. Det finns ingen väg från det ena till det andra.

Användaren som vill schemalägga ett mål måste därför göra det manuellt:
öppna kalendern, trycka "+ Händelse", skriva målets titel, hitta rätt
mål i rullmenyn, och själv räkna ut hur lång tid målet tar. Varje mål som
ska schemaläggas kostar alltså en handfull klick och en egen uträkning.

Det andra problemet är att flytta en händelse som ligger fel. En händelse som
ska flyttas från tisdag till onsdag måste öppnas i formuläret, och de två
datetime-local-fälten skrivas om manuellt. Att "hålla i" en händelse och
draga den vore den naturliga rörelsen - samma som i en vanlig kalender.

## Lösning

### Dra ett mål in i kalendern

Målrader i mållistan (/goals) och i prioriteringslistan (/priority) går att
dra. När ett mål släpps i kalendern skapas en händelse av det:

- Händelsen får målets **titel**, **ikon** och **koppling** (`goal_id`).
- Händelsen hamnar vid **markören** - den tid som pekaren pekade på när
  målet släpptes.
- Tiden **snappar** till en rimlig upplösning, som beror på hur utzoomat
  vyn är:

| Tidsfönster | Upplösning |
| ----------- | ---------- |
| upp till 2 dygn | 15 minuter |
| upp till 10 dygn | 30 minuter |
| upp till 90 dygn | 1 timme |
| mer än 90 dygn | 1 dygn |

### Listan med tidsalternativ

Att veta *när* ett mål ska ske är bara halva frågan - den andra halvan är
hur lång tid det tar. Därför öppnas en meny när målet släppts, med de
alternativ som är relevanta:

- **Målets egen uppskattning** ligger först, om målet har en
  uppskattad tidsåtgång (`estimated_minutes`). Den är märkt som
  "uppskattning", så att det går att se varför den finns med.
- **Vanliga längder** följer: 15 min, 30 min, 1 h, 2 h, 4 h.
- **Ett eget värde** kan skrivas in, för längder som inte finns i listan.

Ett val skapar händelsen direkt. Menyn kan avbrytas med Escape eller med
Avbryt - då skapas ingen händelse.

### Flytta en händelse

En händelse i tidslinjen kan dras med pekaren:

- När pekaren rör sig följer raden pekaren, och tiden räknas om efter
  samma snapping som ovan.
- När knappen släpps sparas den nya tiden. **Händelsens längd behålls** -
  en timmes händelse som flyttas en timme blir en timme lång, inte två.
- En **förekomst i en återkommande serie** flyttas med en override-rad, så
  att bara den förekomsten flyttas och serien i övrigt ligger kvar.
- Ett klick utan drag öppnar händelsen i formuläret, som tidigare. Ett
  klick som följer direkt efter ett drag öppnar inte formuläret - raden
  har ju redan flyttats.

## Interaktion

### Dragdata

Målet packas i dragdata under MIME-typen `application/x-plan-goal`, som
innehåller målets id, titel, ikon och uppskattning. Kalendern reagerar bara
på den typen - annat som dras i siden påverkar den inte.

### Var logiken ligger

All räkning ligger i `app/utils/calendarScheduling.ts`, utan Vue:

- `serializeGoalDrag` / `parseGoalDrag` - packa och packa upp ett mål.
- `startGoalDrag` / `carriesGoal` - starta en dragning, och avgöra om
  något som dras är ett mål.
- `durationOptions` - listan med tidsalternativ.
- `timeAtOffset` - översätta en pekarposition till en tid.
- `snapStepMs` / `snapTimestamp` - upplösningen, och avrundningen.
- `shiftEventTimes` - flytta en händelse och behålla dess längd.

Det gör att logiken går att testa utan att starta en browser, och att
kalendern bara behöver hantera själva interaktionen.

## Avgränsningar

- Dragningen fungerar med mus och pekare. På touch-skärmar finns ingen
  HTML5-dragning, så en händelse flytts där genom formuläret.
- Målet kopplas till händelsen, men kopplingen är inte synlig i vyn - den
  finns i händelsens formulär under "Kopplat mål".
- Menyn för tidsalternativ visas bara för den händelse som skapas just nu.
  Det finns ingen lista över "planerade mål" i kalendern.

## Acceptance Criteria

- [x] Ett mål kan dras från mållistan till kalendern.
- [x] Släppet visar en meny med tidsalternativ.
- [x] Målets uppskattade tidsåtgång ligger först i menyn.
- [x] Ett eget tidsvärde kan skrivas in.
- [x] Vald tid skapar en händelse med målets titel, ikon och koppling.
- [x] Escape avbryter menyn utan att skapa något.
- [x] En händelse kan flyttas genom att dra den.
- [x] Händelsens längd behålls vid flytt.
- [x] Bara den valda förekomsten flyttas i en återkommande serie.
- [x] Ett klick utan drag öppnar fortfarande händelsen i formuläret.
