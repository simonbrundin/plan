---
title: Boka enskild kommunbokning i MATCHi från /bookings
type: note
tags: [matchi, feature, bookings, projekt:stk]
created: 2026-11-06
updated: 2026-11-06
status: active
---

## Vad

Ny knapp "Boka i MATCHi" på /bookings i stk (repo: stk). Den visas på
kommunbokningar som inte är bokade i MATCHi, har en koppling i
court_mappings och ligger i framtiden. Knappen bokar tiden i MATCHi och
sparar kopplingen i bookings-tabellen (samma tabell som synkroniseringen
använder), så raden blir grön direkt.

## Varför

Tidigare fanns bara en skrivläsande jämförelse (GET /matchi/bookings).
Med en knapp kan klubben själv boka en specifik kommunbokning i MATCHi
utan att vänta på cronjobbet matchi-sync (kör 03:00).

## Viktigt

- Endpointen POST /matchi/bookings/link ÄR skrivande - det är enda
  MATCHi-mutationen som triggas av en användare i gränssnittet. Den
  skrivläsande jämförelsen är oförändrad.
- Serveraten använder kommunens tider (läser bokningen i ett fönster
  runt tiden), inte klientens, så en klient inte kan boka en tid som
  inte finns i kalendern.
- Tiden avrundas till hela/halva timmar, som i synkroniseringen.
- Efter bokningen töms MATCHi-cachen (90 s) via
  occupancy.Service.Invalidate, annars skulle den gröna raden synas
  först efter att cachen gått ut.

## Fynd under implementationen

- nginx (src/nuxt/Dockerfile) behövde proxy_read_timeout 180s.
  MATCHis schemasidor är stora, så både jämförelsen (20-55 s) och
  bokningen tar över en minut - defaulten på 60 s bryter anslutningen.
- Lokal utveckling: environments/dev/db/neon.env kan inte sourcingas
  som vanligt (variablerna blir tomma). Använd i stället clusterhemligheten:
  kubectl get secret repo-template-db-neon -n repo-template
  -o jsonpath='{.data.DATABASE_URL}' | base64 -d  (OBS: ta bort
  &channel_binding=require - lib/pq stöder inte det).

## Test

Bokad, verifierat och sedan avbokat + raderad rad i databasen, så
tillståndet i MATCHi är oförändrat.
