---
title: ZITADEL-import: application-ID bevaras men OIDC client-ID genereras om
type: note
tags: [zitadel, oidc, lesson, migrering, projekt:simon]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [concepts/zitadel]
---

Vid ZITADELs officiella Cloud->self-hosted-import bevarades application-ID:n (t.ex. Plan-appen 400928499969314176) men OIDC client_id genererades om (t.ex. 401358699613914656 i stället för det historiska 391431970426454645). Det beror på att addOIDCApp sätter ClientID internt i CreateCommands - API-svaret/importeringen kan inte ange historiskt client ID via normalt AddOIDCAppRequest.

Konsekvens: alla klientreferenser i infra (Envoy SecurityPolicies, Plan-appens env, talos-patchen) pekade på historiska client ID:n som inte existerade efter importen, vilket gav OIDC-fel som 'secret openclaw/openclaw-oidc-client does not exist' och 'The requested redirect_uri is missing in the client configuration'.

Lösning i den här återställningen: client_id skrevs direkt i eventstore (eventstore.events2, eventen project.application.config.oidc.added/changed) och i projectionen (projections.apps7_oidc_configs) så att både historiska och nya ID:n pekar på rätt app. Inte idealiskt - eventstore och projection ska normalt hållas ihop av ZITADEL:s egna kommandon - men det är det enda sättet att bevara exakta historiska ID:n. Bättre alternativ för framtiden: skapa klienterna via management-API:t med önskat client ID där API:t stödjer det, och dokumentera ID:n i 1Password.

## Relaterade

- [[concepts/zitadel]]

