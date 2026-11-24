---
title: Learn: REST-API och Nuxt-SPA mot samma kunskapsbas
type: note
tags: [learn, go, api, nuxt, feature, session, projekt:learn]
created: 2026-11-03
updated: 2026-11-17
status: active
related: [notes/learn-designsystem-a3-folio-i-nuxt, notes/learn-kunskapssystemet-portat-fran-python-till-go]
---

Kunskapssystemet i 'learn' exponerades nu även över HTTP. Nya paketet src/go/internal/api (Gin-handlare) plus routes i cmd/api/main.go; OpenAPI-specen regenererades med swag och Scalar UI visar 14 endpoints.

Arkitekturbeslut: handlarna arbetar mot gränssnittet store.Repository, inte mot *store.Store konkret. Det gjorde att learn.BuildGraph/ImportBook kunde skrivas om till interface och att ett PostgreSQL-lager kan sättas in senare utan att röra HTTP-lagret. SQLite (modernc.org/sqlite) är fortfarande den enda implementationen.

Viktig fälla: API:t kan inte krascha för att kunskapsdatabasen inte kan öppnas - i distroless-image saknas ofta en skrivbar hemkatalog. Därför faller store.Open tillbaka på ':memory:' och /health rapporterar knowledge_base: ready|empty. Utan det hade poden kraschat och readiness probe failat.

Nuxt-SPA:t (src/nuxt/app/pages/index.vue) visar böcker, koncept, grafen som SVG, frågor och quiz mot /api/. Fällan här: i Nuxt 4 fungerar INTE devServer.proxy - det ska vara nitro.devProxy, annars svarar dev-servern med SPA-html för /api-anrop. I produktion går /api/ via nginx (samma relativa sökväg som health.vue använder).

Verifiering: go test ./... (alla paket gröna), docker build + körning av API-image (health, docs och books svarar), samt Playwright-skript mot dev-servern som klickade koncept, ställde en fråga och svarade på en quizfråga - inga JS-fel. Playwright-paketet hade ingen egen browser; använde systemets chromium via executablePath: '/usr/bin/chromium'.

## Relaterade

- [[notes/learn-designsystem-a3-folio-i-nuxt]]
- [[notes/learn-kunskapssystemet-portat-fran-python-till-go]]
