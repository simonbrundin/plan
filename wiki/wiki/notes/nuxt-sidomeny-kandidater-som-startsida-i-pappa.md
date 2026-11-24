---
title: Nuxt-sidomeny + /kandidater som startsida i pappa
type: note
tags: [nuxt, ui, feature, projekt:pappa]
created: 2026-11-16
updated: 2026-11-17
status: active
related: [notes/kandidatlistan-heter-nu-personer]
---

Alla sidor i pappa:s Nuxt-app ligger nu under en default-layout med sidebar (src/nuxt/app/layouts/default.vue). Menystruktur: Rekrytering (Kandidater /kandidater, Uppdrag /uppdrag), Mail (Inkorg /mail, Regler /mail/rules) och Kameo utbetalningar som egen toppnivå utan rubrik (sökvägen ligger under /mail men är inte en del av Mail). Varför Kameo utanför: uppgiften säger uttryckligen att den inte ska ligga under Mail, så menyn får en egen sektion även om URL-strukturen är nästlad. Aktiv markering väljer längsta sökvägsmatchning, så /mail och /mail/kameo-utbetalningar inte lyser samtidigt. Kandidatlistan flyttades från / till /kandidater och / redirectar dit via routeRules + definePageMeta, eftersom sidomenyn nu bär navigeringen och sidorna slipper egna tillbaka-knappar.

## Relaterade

- [[notes/kandidatlistan-heter-nu-personer]]

