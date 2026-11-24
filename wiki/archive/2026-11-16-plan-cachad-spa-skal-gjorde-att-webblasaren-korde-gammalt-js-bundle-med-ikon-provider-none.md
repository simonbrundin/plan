---
title: Plan: cachad SPA-skal gjorde att webbläsaren körde gammalt JS-bundle med ikon-provider none
type: note
tags: [bug, nuxt, cache, plan, projekt:plan]
---

Symtom: efter att ikonfixen deployats till dev-klustret söktes ikoner korrekt men rutorna i ikonväljaren var tomma (bara grå kvadrater).

Förklaring: webbläsaren körde fortfarande det GAMLA JS-bundlet (icon.provider = 'none'), medan /api/icons-routen (ny, server-side) svarade korrekt. Alltså: sökningen fungerade men klienten hade ingen ikon-provider alls. Bland annat därför att index.html (SPA-skalet) serverades UTAN cache-headers - browsern återanvände en gammal HTML som pekade på gamla, hashade bundles. Ett redan öppet tab med gamla JS gav samma symptom.

Verifiering (headless chromium mot dev-klustret, inloggad via /auth/dev-login): 38 träffar på 'muscle', varje ikon-span hade mask-image + --svg satt => nuvarande build renderar ikonerna korrekt. Det serverade bundlet innehöll icon:{provider:'iconify', fallbackToApi:true}.

Fix: ny middleware src/nuxt/server/middleware/spa-cache.ts sätter 'cache-control: no-store, must-revalidate' på HTML-svar (men inte /api/ eller /_nuxt/, som behåller sina egna cache-regler).

Lärdom: SPA-appar utan cache-headers på index.html ger spretiga 'halvfungerande' appar efter deploy - server-side nya endpoints fungerar medan gamla JS-funktioner tyst faller bort. Testa alltid den DEPLOYADE versionen i en browser, inte bara lokalt.
