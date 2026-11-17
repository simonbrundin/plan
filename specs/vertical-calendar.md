# Vertikal kalender med kontinuerlig zoom

## Problem

Plan saknar en kalender som gör det möjligt att läsa och navigera tid på ett
naturligt sätt utan att byta mellan separata dag-, vecka- och månadsvyer.
Användaren ska kunna fokusera på ett tidsintervall och samtidigt förflytta sig
framåt eller bakåt i tiden.

## Lösning

Skapa en kalender som visar tiden som en vertikal, kontinuerlig tidslinje:

- Tiden löper uppifrån och ned.
- Vid första visningen ligger `nu` längst upp.
- Standardintervallet är 24 timmar från starttiden.
- Vertikal skrollning ändrar tidsskalan kontinuerligt.
- Horisontell skrollning flyttar starttiden framåt eller bakåt utan att ändra
  hur mycket tid som visas.
- Vertikal skrollning kan fortsätta visa mer tid utan en produktdefinierad
  maxgräns.
- När användaren skrollar långsamt ska zoomningen ske långsamt. Ett snabbt
  scrollsteg ska komprimera tiden snabbare.
- Tidsfönstrets topp flyttas framåt eller bakåt när användaren skrollar
  horisontellt.
- Vertikal skrollning zoomar in eller ut och ändrar tidsintervallet.
- Det ska inte finnas fasta zoomnivåer; övergången mellan tidsintervallen ska
  vara kontinuerlig.

Exempel:

```text
Startläge: 24 timmar
08:00 idag       <- nu
09:00
...
08:00 imorgon
```

Efter vertikal zoomning:

```text
Startläge: cirka 3 dagar
08:00 måndag      <- nu
Tisdag
Onsdag
08:00 torsdag
```

## Interaktion

### Tidsfönster

Kalendern har ett tidsfönster med:

- `startTime`: tiden som visas överst
- `endTime`: tiden som visas längst ned
- `visibleDuration`: `endTime - startTime`

Vid öppning sätts `startTime` till aktuell tid och `visibleDuration` till 24
timmar. Därefter beräknas `endTime` från dessa värden.

### Skrollning och zoom

Vertikal skrollning ska påverka `visibleDuration` kontinuerligt i stället
för att välja bland fasta förinställningar. Zoomningen ska använda en
multiplikativ/logaritmisk skala. Horisontell skrollning ändrar i stället
`startTime` proportionellt mot tidsfönstrets storlek:

- liten/långsam vertikal förflyttning ger en liten förändring av
  tidsintervallet
- stor/snabb vertikal förflyttning ger en större förändring
- horisontell skrollning åt höger flyttar starttiden framåt
- horisontell skrollning åt vänster flyttar starttiden bakåt

Tangentbordsbeteendet följer samma konvention: `j`/`↓` och `k`/`↑` zoomar
(som vertikal skrollning), medan `h` och `l` flyttar tidsfönstret (som
horisontell skrollning).

### Rendering

- Tidsmarkeringar ska anpassa sin täthet efter aktuell zoomnivå.
- Datumgränser ska vara tydliga när tidsfönstret passerar midnatt.
- `nu` ska markeras visuellt när den aktuella tiden ligger inom fönstret.
- Innehåll som kalenderhändelser läggs ovanpå tidslinjen och följer samma
  tidskoordinat som markeringarna.

## Implementation

1. Skapa en återanvändbar kalenderkomponent i Nuxt med en viewport och en
   virtuell tidsyta.
2. Inför en tidsmodell som konverterar mellan tid och vertikal pixelposition:
   - `timeToY(time)`
   - `yToTime(y)`
3. Rendera bara det tidsintervall som behövs för viewporten plus en buffert,
   så att långa tidsintervall inte kräver en oproportionerligt stor DOM.
4. Implementera horisontell skrollning som panorerar `startTime` utan att ändra
   `visibleDuration`.
5. Använd en multiplikativ zoomfunktion via vertikal skrollning utan en
   produktdefinierad maxgräns.
6. Låt scrollhastighet och scrollavstånd styra hur snabbt zoomningen sker.
7. Lägg till responsiv rendering för tidsmarkeringar, datumgränser och
   aktuell-tid-markering.
8. Lägg till tester för tidskonvertering, hastighetsberoende zoom och
   gränsfall vid midnatt/datumbyte.

## Avgränsningar

- Flera kalendrar stöds sedan [calendrarna med egna färger](calendars.md).
  Tidszoner stöds fortfarande inte utöver att tidsstämpeln lagras med
  IANA-zon.
- Minsta zoomnivå fastställs vid implementation efter att viewportens storlek
  och händelsernas informationsdensitet har verifierats. Största intervallet
  begränsas bara av JavaScript `Date`-typens tekniska maxintervall.
- Touch- och musinteraktion ska använda samma tidsmodell, men kan ha olika
  input-adaptrar.
- Zoomningen saknar en produktdefinierad maxgräns. Den stoppas först av
  JavaScript `Date`-typens tekniska maxintervall, vilket i praktiken är långt
  bortom normal kalenderanvändning.

## Acceptance Criteria

- [ ] När kalendern öppnas visas 24 timmar med aktuell tid överst.
- [ ] Tiden ökar visuellt från toppen mot botten.
- [x] Vertikal skrollning zoomar kontinuerligt in och ut utan fasta steg.
- [x] Horisontell skrollning flyttar starttiden framåt och bakåt.
- [x] Långsam vertikal skrollning ger långsam zoomning och snabb
      skrollning ger snabb komprimering.
- [x] Horisontell skrollning flyttar tidslinjens starttid utan att ändra
      tidsfönstrets storlek.
- [ ] Tidsmarkeringar och datumrubriker anpassas efter aktuell zoomnivå.
- [ ] Aktuell tid markeras när den ligger inom det synliga tidsfönstret.
- [ ] Tidskonvertering och interaktioner har automatiserade tester.

## Status

- [x] Define
- [x] Plan
- [x] Execute
- [x] Verify
