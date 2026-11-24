---
title: Parallella agent-sessioner i samma repo och databas
type: note
tags: [incidents, workflow, projekt:plan, dev-miljo]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis, notes/vertikal-kalender-med-kontinuerlig-zoom]
---

## Varför det här spelar roll

Det har visat sig finnas flera parallella agent-sessioner som arbetar i
`~/repos/plan` samtidigt, mot samma git-arbetsyta och samma Neon-databas. Det
har orsakat två konkreta problem som är svåra att felsöka om man inte vet om
det.

## Symptom 1: `.output` raderas mitt under en pågående build

Den parallella sessionen kör `rm -rf .output && bun run build` i
`src/nuxt`. En server som startats från `.output` börjar då leverera
 trasiga assets:

```
Refused to apply style ... because its MIME type ('application/json') is
not a supported stylesheet MIME type
Failed to load resource: the server responded with a status of 500
```

Symptomet är vilseledande - det ser ut som ett byggfel, men bygget är
fint. HTML-sidan refererar en CSS-fil med en hash som inte längre finns,
eftersom katalogen byggdes om under den körande processen.

**Lösning:** bygg till en egen katalog och starta därifrån.

```bash
bun run build
rm -rf /tmp/nuxt-mine && cp -r .output /tmp/nuxt-mine
NUXT_GO_API_INTERNAL_URL=... PORT=3114 node /tmp/nuxt-mine/server/index.mjs
```

En dev-server (`bun run dev`) fungerar också men rör inte `.output`. Observera
att dev-servern behöver `NUXT_GO_API_INTERNAL_URL` satt - annars pekar
proxyn mot fel port.

## Symptom 2: dev-sessioner ogiltigförklaras

Sessionen i `plan_session` slutade fungera flera gånger under en och samma
session, trots att cookien var i sin tid. Orsaken var att Go API:t startades
om av den parallella sessionen. Inloggningen måste göras om, och
`/api/proxy/goals/prioritized` svarade 401.

**Symtom att misstolka:** en 401 från kalendern eller prioriteringslistan
ser ut som ett auth-fel i den egna koden, men är bara en omstartad server.

## Konsekvens för databasen

`calendar_events` i dev-databasen tömdes vid någon tidpunkt mellan 03:18Z och
04:46Z den 17 november 2026. Händelserna med id 1-19 fanns då inte kvar,
och de syntes inte i någon Time Travel-punkt. Den enda kvarvarande
förklaringen är att en parallell session raderade dem - det fanns inga
snapshots, inga lokala dumpar och ingen backup.

**Regel:** rader aldrig data i dev-databasen med ett brett tidsintervall.
Använd `scripts/cleanup-test-events.sh --force`, som bara raderar rader med
`[test]`-prefix, och kontrollera Neon Time Travel före varje radering.

## Hur man upptäcker en parallell session

```bash
ps -eo pid,lstart,etime,args | grep -E "nuxt build|index.mjs" | grep -v grep
```

Flera `nuxt build`-processer samtidigt, eller en `.output` som byts om
under ett pågående test, är tydliga tecken.

## Relaterade

- [[notes/kalenderns-nivavaxling-animeras-crossfade-pop-in-och-hysteresis]]
- [[notes/vertikal-kalender-med-kontinuerlig-zoom]]
