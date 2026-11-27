---
title: Authflytten i plan: sessionen ligger nu i Go API:t, inte i Nuxt
type: note
tags: [auth, session, nuxt, go, zitadel, cookie, dev, projekt:plan]
created: 2026-11-10
updated: 2026-11-27
status: active
related: [concepts/zitadel, notes/dev-inloggning-i-plan-session-token-db-rattigheter-och-oauth-url-er, notes/started-finished-kunde-inte-rensas-i-plan-null-ignorerades-i-go-api-t, notes/plan-api-webblasaren-anropar-api-t-direkt-nuxt-proxyn-ar-borttagen]
---

Flyttade authen från Nuxt till Go API:t i plan-repot (2026-11-10).

Varför: sessionstoken cirkulerade i URL:en (Go redirectade till
Nuxt /api/auth/callback?session=...), Nuxt satte själv cookies och
översatte cookie -> Bearer i proxyn. Två ägare av sessionen = två
sätt att gå sönder.

Nu:
- Go skapar sessionen och sätter plan_session (httpOnly) + plan_user
  (lasbar hint) direkt i svaret. Sessionen valideras från cookie ELLER
  Authorization: Bearer (SessionTokenFromRequest).
- Go har GET /api/v1/auth/session och POST /api/v1/auth/logout.
- Nuxt var en tunn proxy (server/utils/goApi.ts): vidarebefordrade Cookie,
  kopierade tillbaka Set-Cookie och Location, och foljde inte redirects
  (annars kor OAuth-floden server-side). Proxyn togs bort 2026-11-27 -
  webbläsaren anropar nu API:t direkt, se
  [[notes/plan-api-webblasaren-anropar-api-t-direkt-nuxt-proxyn-ar-borttagen]].
- Nuxt-routen /api/auth/callback ar borttagen - Go redirectar direkt till
  appens rot.
- Cookien sattas for host "localhost" (porten spelar ingen roll for
  cookies), sa dev fungerar trots att callbacken hamnar pa port 4100 och
  appen pa 3100.

Viktig lardom: Atlas Operator applicerar schemat DEKLARATIVT - bara DDL.
INSERT-raderna i schema.yaml (statuses + Root-malet) kors aldrig att
koras. Applikationen seedar darfor sjalv (internal/database/seed.go) vid
start. Detta avhjalpar ocfar heller en sekvensbugg: Root-malet seedas med
explicit id = 1, vilket lat identity-sekvensen efter - utan fix krockar
den forsta goal-inserten med "duplicate key value violates unique
constraint goals_pkey".

Varning: en Atlas-apply ar destruktiv. Ett schema som saknar en tabell
RADERAR den. Testa aldrig genom att tomma schema.yaml - det tog bort
devdatabasens tabeller under arbete med detta.

Deploy-ordning: Go och Nuxt måste deployas tillsammans - den gamla
Nuxt-bilden forutsatte /api/auth/callback och cookie->Bearer i proxyn.

## Relaterade

- [[notes/started-finished-kunde-inte-rensas-i-plan-null-ignorerades-i-go-api-t]]
- [[concepts/zitadel]]
- [[notes/dev-inloggning-i-plan-session-token-db-rattigheter-och-oauth-url-er]]
