---
title: Eviga mål i plan - kan inte slutföras
type: note
tags: [feature, goals, plan, ui, go, nuxt, projekt:simon]
created: 2026-11-16
updated: 2026-11-17
status: active
related: [notes/mal-kan-ha-due-datum-deadline, notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t, notes/manuell-sortering-av-mal-pa-priority, notes/priorityfilter-ska-behandla-saknat-started-som-ej-paborjat]
---

Mål i plan kan nu markeras som EVIGA (eternal): de pågår hela livet (t.ex. 'kroppen') och kan inte slutföras.

VARFÖR: vissa mål ska aldrig kunna markeras som klara - de är livslånga. Tidigare fanns ingen markering för det, så 'd' (klar) gjorde dem till 100%-klara.

HUR: ny boolean 'eternal' i goals-tabellen (kanoniskt schema: environments/base/db/schema.yaml, appliceras av Atlas Operator). Knapp med ∞-ikon i målets header + tangent 'e' på målsidan och i /goals-listan. Eviga mål visas i CYAN (text-cyan-300) med ∞-symbol.

FÄRG och utseende: alla färgkonstanter för eviga mål ligger i src/nuxt/app/utils/eternalStyle.ts - byt färgen där så slår den igenom på målsidan, mållistan och prioriteringslistan. Användaren ville inte ha en förklarande text under titeln ('Evigt mål - kan inte slutföras') - den är borttagen; bara ∞-symbolen och färgen markerar målet.

REGEL (både i API och UI): ett evigt mål kan inte slutföras. PATCH goals/:id med finished på ett evigt mål ger 400 'Ett evigt mål kan inte slutföras'. När ett mål markeras som evigt nollställs finished. UI-regeln ligger i src/nuxt/app/utils/goalRules.ts (canBeFinished).

VERIFIERAT: go test ./internal/models, bun test test/unit, och live i dev-klustret (Neon): PATCH eternal=true -> 400 vid försök att sätta finished; evigt=false -> finished fungerar igen.

FALLFRUKT: sessions för dev-test kan skapas genom att köra API:t lokalt med NUXT_SESSION_PASSWORD ur secret 'auth' (samma secret som klustret använder) - då signerar dev-login en cookie som kluster-API:t accepterar, eftersom sessions-tabellen ligger i samma Neon-databas. API-nycklar (X-API-Key) räcker bara för GET - AuthMiddleware kräver ändå en session för POST/PATCH.

## Relaterade

- [[notes/mal-kan-ha-due-datum-deadline]]
- [[notes/priorityfilter-ska-behandla-saknat-started-som-ej-paborjat]]
- [[notes/manuell-sortering-av-mal-pa-priority]]
- [[notes/mal-kan-ha-en-lang-beskrivning-redigerad-med-tangenten-t]]
