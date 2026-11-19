# Händelserader i den vertikala kalendern

## Problem

Den vertikala kalendern komprimerade sina händelser när tidsfönstret blev
större. Från sju dagar och uppåt grupperades händelserna per dag, månad eller
år och visades som prickar. Användaren såg då bara antalet händelser i en
period - aldrig någon titel, ikon eller klockslag.

Det gjorde kalendern nästan oanvändbar för den som vill se vad som ligger
under ett par veckor eller en månad. Användaren tvingades zooma in till
timmar för att ens se vad händelserna heter, och då försvann just den
översikt som gjorde det möjligt att planera.

Roten till problemet är geometri: en tidslinje som är linjär i tid gör en
timmes händelse ungefär en pixel hög i ett månadsfönster. En etikett får
inte plats. Lösningen kan därför inte vara att visa korten som de är - den
måste vara att bestämma vad som ska stå på axeln i stället.

## Lösning

Varje händelse visas alltid som en rad, på alla zoomnivåer:

- Radens höjd följer händelsens längd, men aldrig mindre än 16 pixlar. Det är
  den lägsta höjden som går att läsa ikon och titel på. En lång händelse
  begränsas till en fjärdedel av vyn, så att den inte tar över kolumnen.
- Radens innehåll är ikon, titel och starttid. Titeln klipps med ellips när
  den inte ryms.
- Alla rader ligger i **en enda kolumn**, så att varje rad får hela bredden
  och titeln går att läsa. Rader som överlappar i tid kan inte ligga ovanpå
  varandra, så en rad skjuts ned under den som redan står. Tidsaxeln blir
  då ungefärlig, men raden visar sin exakta tid.
- En händelse som inte hinner med i kolumnen får inte försvinna tyst. Den
  visas som en "+N till"-knapp, och knappen öppnar en panel där de går att se
  och klicka på.

Det finns alltså inga prickar och inga tomma prickar: en händelse är antingen
en läsbar rad eller en siffra i en knapp som leder till en lista.

## Döljning av återkommande serier

En kolumn ryms inte alltid. En daglig rutin som syns trettio gånger i en
månadsvy tar över kolumnen, och då blir de enskilda händelserna omöjliga att
läsa.

Då döljs återkommande serier, i en bestämd ordning:

1. **Dagliga serier** döljs när de har fler förekomster i vyn än gränsen.
   Gränsen är inställbar och ligger som standard på sju.
2. Om kolumnen ändå inte ryms döljs den serie som har flest förekomster i vyn.
   En serie med en enda förekomst döljs inte - den är ju inte brusan, bara
   olycklig.
3. **Enstaka händelser döljs aldrig.** De kan i stället hamna i "+N till".

Döljningen är aldrig tyst:

- En dold serie står kvar i sammanfattningen ovanför tidslinjen, märkt som
  "dold", med sitt antal förekomster.
- Klick på "Visa ensam" visar bara den serien. Då sätts döljningsreglerna åt
  sidan, så att även en viktig daglig rutin går att se.
- Döljningen har **hysteresis**: en serie som varit dold förblir dold tills
  antalet förekomster faller *under* gränsen. Annars fladdrar serier fram och
  tillbaka när man scrollar över gränsen.

## Implementation

1. `app/utils/calendarLayout.ts` packar händelserna till rader i en kolumn.
   Allt räknas i pixlar, eftersom bara pixlarna kan avgöra om en rad hinner
   rymmas under en annan. Funktionen är ren och testad separat.
2. `app/utils/calendarVisibility.ts` planerar vilka serier som ska döljas.
   Planen är ren och tar den tidigare döljningen som input, så att
   hysteresen kan räknas utan att planen blir stateful.
3. `VerticalCalendar.vue` renderar raderna, märker dolda serier i
   sammanfattningen och visar de händelser som inte ryms i en panel.
4. `app/composables/useCalendarSettings.ts` läser och sparar inställningarna i
   webbläsaren.
5. Detaljnivåerna (full, compact, daily, monthly, yearly) försvinner, eftersom
   de fanns just för att bestämma när händelser skulle komprimeras till
   prickar. Tidsaxeln behåller sina fyra skalor (timme, dag, månad, år) i
   `app/utils/calendarScale.ts`.

## Avgränsningar

- Tidsaxeln är ungefärlig i de fall då raderna skjuts ned. Radens etikett
  visar den exakta tiden.
- Kolumnen rymmer så många rader som får plats. När den blir för full döljs
  serier, och enstaka händelser hamnar i "+N till".
- Inställningarna sparas i webbläsaren, inte i kontot. De följer med på den
  enhet de gjordes på.

## Acceptance Criteria

- [x] Varje händelse visas som en rad med ikon, titel och starttid.
- [x] Ingen rad är lägre än 16 pixlar, oavsett zoomnivå.
- [x] Alla rader ligger i en enda kolumn, och ingen rad döljer en annan.
- [x] Det finns inga prickar: inga händelser visas som en punkt.
- [x] Dagliga serier döljs när de har fler förekomster i vyn än gränsen, och
      gränsen går att ändra i inställningarna.
- [x] En dold serie listas i sammanfattningen, märkt som "dold", och kan
      visas ensam.
- [x] En dold serie förblir dold tills antalet förekomster faller under
      gränsen.
- [x] Enstaka händelser döljs aldrig utan att det syns i "+N till".
- [x] Klick på en rad öppnar händelsen för redigering.

## Status

- [x] Define
- [x] Plan
- [x] Execute
- [x] Verify
