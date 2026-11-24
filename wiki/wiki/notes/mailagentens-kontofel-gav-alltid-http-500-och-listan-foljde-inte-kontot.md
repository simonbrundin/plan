---
title: Mailagentens kontofel gav alltid HTTP 500 och listan följde inte kontot
type: note
tags: [bug, http, mail, nuxt, projekt:pappa, projekt:simon]
created: 2026-11-16
updated: 2026-11-16
status: active
related: [notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter, notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema]
---

2026-11-16: Efter att kontovalet lades till i mailagenten i pappa kom tre smarare saker:

1. HTTP 500 for allt. Ett ej inloggat konto och ett okänt konto gav bada HTTP 500. Nu mappas felet till status: "unknown account" -> 404, "not logged in"/"login failed" -> 401, allt annat -> 500. Klienten kan alltsa saga "logga in" (401) istallet for "serverfel" (500).

2. Gammal lista under nytt kontonamn. När kontot byttes i granssnittet lam den gamla kontots mejl kvar i listan tills man tryckte "Hamtta mejl" - de visades da under det nya kontots namn. Nu tommes listan via watch(currentAccount) i useMailEmailProcessing.

3. Nuxt-komponenterna lag i components/mail/ och registrerades med prefix (MailAccountSelector), medan sidan refererade dem utan prefix (AccountSelector). Vue hittade dem aldrig - tyst fel, ingen byggvarning. Fix: pathPrefix: false i nuxt.config.ts.

## Varfor det tog sa lang tid att hitta

Symtomet var "kontodropdownen syns inte". Koden var korrekt, API:t svarade med bada kontona, men komponenterna renderades inte alls. Ingen byggvarning pekade riktningen - man fick gissa pa Nuxt-komponentregistreringen.

Satt att verifiera utan browser: greppa SSR-HTML:t (eller det byggda .output/) efter komponentens CSS-klass. Finns klassen i HTML:t sa finns komponenten. Finns den inte ar den inte importerad alls.

## Testmetod

Playwright med systemets chromium (/usr/bin/chromium) - den installerade playwright-paketet har ingen egen browser. Anvand chromium.launch({ executablePath: '/usr/bin/chromium' }).

## Relaterade

- [[notes/mailagentens-mcp-verktygen-saknade-konto-och-hade-fel-schema]]
- [[notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter]]
