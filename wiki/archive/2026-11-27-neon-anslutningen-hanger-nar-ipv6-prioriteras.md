---
title: Neon-anslutningen hänger när IPv6 prioriteras
type: note
tags: [neon, ipv6, postgres, plan, bug, projekt:plan]
---

# Neon-anslutningen hänger sig när IPv6 prioriteras (2026-11-27)

Symtom: Go API:t mot Neon (Neons dev-branch, `environments/dev/db/neon.env`)
startar och skriver "Connected to database" men lyssnar aldrig på sin port.
Processen står i `SYN-SENT` mot Neons **IPv6-adress**.

## Varför

Värden har både A- och AAAA-poster. Resolvern returnerar ibland IPv6 först,
och maskinens IPv6-rutt finns men leder ingenstans - TCP-synkarna försvinner
i tomrum. `connect_timeout` i DSN hjälper inte, eftersom det bara timeoutar
anslutningen, inte gör att IPv4 provas.

Att det ibland fungerar gör felet svårt att hitta: samma URL fungerar från
psql men inte från API-processen.

## Lösningen som fungerar

1. Använd **IPv4-adressen** i DSN:en.
2. Lägg till `options=endpoint%3D<endpoint-id>` - Neon behöver endpoint-id:t
   för att veta vilken databas som avses när man kommer in via IP (SNI går
   inte att använda då).
3. `sslmode=require` räcker - certifikatet verifieras inte mot IP-adressen.

Exempel (Neons endpoint-id är första delen av domännamnet):

```
postgresql://user:pass@52.29.229.153/neondb?sslmode=require&connect_timeout=10&options=endpoint%3Dep-shy-brook-algcxnsz
```

`hostaddr=` fungerar **inte** - pgx v5 tolkar inte den parametern (lib/pq
gör det).

## Var detta används

`~/.config/plan-goal-widget/api.env` - den lokala API-instansen som
skrivbordswidgeten läser ifrån. Samma DSN kan återanvändas av andra
lokalprocesser mot Neon.

## Fällan

Pooler-endpointen (`-pooler.`) var också svår att nå - den kräver
`channel_binding=require`, som bara pgx stöder. Kör en egen API-instans mot
det **direkta** endpointet i stället för poolern.
