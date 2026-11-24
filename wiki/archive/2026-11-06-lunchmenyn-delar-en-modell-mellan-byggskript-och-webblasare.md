---
title: Lunchmenyn delar en modell mellan byggskript och webblasare
type: note
tags: [lunch, arkitektur, beslut, projekt:stk]
---

Hamtningen av matochmat.se ligger nu i src/nuxt/app/utils/lunchSource.ts och anvands av bade scripts/fetch-lunch.ts (byggtid, skriver app/data/lunch.json) och app/composables/useLiveLunchMenu.ts (startsidans knapp 'Hamtta ny meny', hamtar i webblasaren). Skalet: matochmat.se skickar access-control-allow-origin: *, sa ingen backend behovs och det fungerar even pa den statiska GitHub Pages-sajten. Startsidan kan bla addra veckor med < och > for att se om nastan vecka ar publicerad. Publicera (bygga om sajten) gor GitHub Actions, schemalagd varje morgon kl. 06:00 svensk tid. Lardom: duplicera inte API-tolkningen - den har en lage latt att dela mellan byggtid och runtime.
