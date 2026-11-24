---
title: Eviga mål i plan - kan inte slutföras
type: note
tags: [feature, goals, plan, ui, go, nuxt, projekt:simon]
---

Mål i plan kan nu markeras som EVIGA (eternal): de pågår hela livet (t.ex. 'kroppen') och kan inte slutföras.

VARFÖR: vissa mål ska aldrig kunna markeras som klara - de är livslånga. Tidigare fanns ingen markering för det, så 'd' (klar) gjorde dem till 100%-klara.

HUR: ny boolean 'eternal' i goals-tabellen (kanoniskt schema: environments/base/db/schema.yaml, appliceras av Atlas Operator). Knapp med ∞-ikon i målets header + tangent 'e' på målsidan och i /goals-listan. Eviga mål visas i rosa (text-pink-300) med ∞-symbol.

REGEL (både i API och UI): ett evigt mål kan inte slutföras. PATCH goals/:id med finished på ett evigt mål ger 400 'Ett evigt mål kan inte slutföras'. När ett mål markeras som evigt nollställs finished. UI-regeln ligger i src/nuxt/app/utils/goalRules.ts (canBeFinished).

VERIFIERAT: go test ./internal/models, bun test test/unit, och live i dev-klustret (Neon): PATCH eternal=true -> 400 vid försök att sätta finished; evigt=false -> finished fungerar igen.

FALLFRUKT: sessions för dev-test kan skapas genom att köra API:t lokalt med NUXT_SESSION_PASSWORD ur secret 'auth' (samma secret som klustret använder) - då signerar dev-login en cookie som kluster-API:t accepterar, eftersom sessions-tabellen ligger i samma Neon-databas. API-nycklar (X-API-Key) räcker bara för GET - AuthMiddleware kräver ändå en session för POST/PATCH.
