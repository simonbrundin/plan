---
title: Kalendern visar händelser som rader, aldrig som prickar
type: note
tags: [kalender, feature, zoom, ui, session, projekt:plan]
---

Användaren klagade på att kalendern komprimerade event till prickar redan vid mer än sju dagar, så att ingen titel syntes.

**Varför:** De fem detaljnivåerna (full/compact/daily/monthly/yearly) fanns för att bestämma NÄR händelser skulle komprimeras till prickar. Gränsen vid sju dagar gjorde att den vanligaste vyen - en eller två veckor - saknade titlar. Roten är geometri: en tidslinje som är linjär i tid ger en timmes händelse ~1 px hög i ett månadsfönster, så etiketter får inte plats. Lösningen kan därför inte vara 'visa korten som de är', utan måste bestämma vad som står på axeln i stället.

**Vad som gäller framöver:**
- Varje händelse är en rad med ikon, titel och starttid, på alla zoomnivåer.
- Radens höjd följer händelsens längd men är aldrig lägre än 16 px (lägsta lästa höjd).
- Överlappande rader packas i kolumner (max 5 per klump), inga prickar finns kvar.
- Händelser som inte får plats visas som en '+N'-knapp bredvid klumpen, med en panel som listar dem.
- Packningen räknas i pixlar, inte procent - bara pixlarna kan avgöra om en rad hinner under en annan. Koden ligger i app/utils/calendarLayout.ts (ren funktion, testad).
- Detaljnivåerna är borttagna. Tidsaxeln behåller sina fyra skalor i app/utils/calendarScale.ts.

**Öppet:** Vilka händelser som ska filtreras bort helt beroende på hur mycket tid som syns är inte avgjort. Just nu visas allt, så en tätt återkommande serie kan ge tusentals rader.
