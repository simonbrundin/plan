---
title: Min kontaktlista - jag som nod i grafen (pappa)
type: note
tags: [pappa, relationer, graf, nätverk, introduktioner, modell, projekt:pappa]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/redigerare-och-agarskap-for-relationer-i-pappa]
---

2026-11-10: Efter ägarskapet kom nästa fråga: 'jag vill kunna lägga till vilka personer som jag känner (alltså michael brundin känner)'. Det är ett annat ting än kandidatrelationer: där är båda parterna kandidater. Här är ena parten JAG - och utan mig i grafen kan appen inte svara på 'vem kan jag nå?'.

MODELL (db/schema.sql, my_contact + my_contact_relation):
- my_contact = en person jag känner. person_id sätts om hon finns i kandidatregistret - DÅ är det samma person, samma nod (annars hade 'Alexander Möllersen' funnits två gånger i grafen: som kandidat och som kontakt). Annars räcker namn/org/titel, så kontakten kan vara någon utanför registret.
- my_contact_relation = kontakten känner kandidaten, med egen styrka per koppling. Det är introduktionen.
- styrkan på kontakten = hur väl JAG känner henne (1-10).

VÄGAR: jag -> kontakt -> kandidat. Vägens styrka = den svagaste länken (min av kontaktens styrka och kopplingens). En kontakt som själv är kandidat ger en längre väg: jag -> kontakt -> kandidat -> kandidat via de vanliga kandidatrelationerna. findMyPaths() räknar ut allt, testat i my-contacts.test.ts.

VARFÖR TVÅ TABELLER och inte en: en kontakt kan känna flera kandidater. Hade kopplingen legat på kontakten hade det blivit en-till-många med en lista - och det är just kandidater som ska kunna kopplas.

FALLGROP (fångad i browsertest): GET /api/my-contacts returnerar en ARRAY (som /api/relations gör) men klienten typade det som { contacts }. Listan var alltid tom, inget fel-loggat, API:t svarade 200 med rätt data. Lärdom: när ett API svarar med en array ska typedResponse/wrappers inte sticka in ett objekt - och hämta svaret i klienten ska testas, inte bara endpointen.

FALLGROP 2: testkontakter skapades i den riktiga databasen under browsertestet ('Michael Brundin känner Alexander Möllersen, styrka 8' - ett påhittat påstående). Raderade efteråt.-principen 'inga gissade relationer' gäller även testdata: det som ligger i databasen är påståenden, även om de kom från ett test.

NULÄGET: kontaktlistan är tom - Michael har inte fyllt i sina kontakter än. Funktionen är klar men tom, precis som vägarna mellan kandidater var.

## Relaterade

- [[notes/redigerare-och-agarskap-for-relationer-i-pappa]]

