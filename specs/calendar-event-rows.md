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
  varandra, så en rad skjuts ned under den som redan står.
- Bara **etikettbandet** - de översta pixlarna där titeln står - reserverar
  plats. En lång händelse ska inte skjuta undan alla andra händelser som
  ligger inne i den; de behåller sina egna tider och ritas ovanpå den långa
  radens tomma yta. Rader som ligger längre ned ritas ovanpå, så att ingen
  titel döljs.
- Tidsaxeln är därför så exakt som den kan vara. Den blir *ungefärlig* bara
  för händelser som startar inom några minuter av varandra - de staplas då
  under varandra. Radens etikett visar den exakta tiden.
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

1. **Antal.** En serie visas med högst ett bestämt antal förekomster i vyn.
   Gränsen är inställbar och ligger som standard på sju. Så snart den åttonde
   förekomsten syns, döljs hela serien - den visas som vanligt fram till dess.
2. Om kolumnen ändå inte ryms döljs den serie som har flest förekomster i vyn.
   En serie med en enda förekomst döljs inte - den är ju inte brusan, bara
   olycklig. Vid lika många förekomster döljs den som var dold sist, så att
   det är samma serie som försvinner och kommer tillbaka.
3. **Enstaka händelser döljs aldrig.** De kan i stället hamna i "+N till".

Döljningen är tyst: en dold serie syns inte i tidslinjen. Det är ett medvetet
val - sammanfattningen av serierna ovanför tidslinjen togs bort, eftersom den
skräpade ner vyn. Vill man se serien får man zooma in på den, så att den syns
som vanligt igen.

## Implementation

1. `app/utils/calendarLayout.ts` packar händelserna till rader i en kolumn.
   Allt räknas i pixlar, eftersom bara pixlarna kan avgöra om en rad hinner
   rymmas under en annan. Funktionen är ren och testad separat.
2. `app/utils/calendarVisibility.ts` planerar vilka serier som ska döljas.
   Planen är ren och testad separat.
3. `VerticalCalendar.vue` renderar raderna och visar de händelser som inte ryms
   i en panel. Serierna räknas bara - de visas inte i en egen sammanfattning.
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
- [x] En lång händelse skjuter inte undan händelser som ligger inne i den -
      de behåller sina egna tider.
- [x] Etikettbanden överlappar aldrig, så ingen titel döljs.
- [x] Det finns inga prickar: inga händelser visas som en punkt.
- [x] Varje återkommande serie visas med högst sju förekomster i vyn. Så snart
      den åttonde syns, döljs hela serien, och gränsen går att ändra i
      inställningarna.
- [x] En dold serie visas inte i tidslinjen, och syns inte i någon
      sammanfattning - vyn ska vara ren.
- [x] Enstaka händelser döljs aldrig utan att det syns i "+N till".
- [x] Klick på en rad öppnar händelsen för redigering.

## Status

- [x] Define
- [x] Plan
- [x] Execute
- [x] Verify
