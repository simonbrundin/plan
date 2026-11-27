---
title: "Dashboarden: ikonerna hämtades från Iconifys CDN och gjorde SSR-sidladdningen till en 504"
type: note
tags: [dashboard, nuxt, bug, lesson, projekt:simon]
created: 2026-11-27
updated: 2026-11-27
status: active
related: []
---

Dashboarden på `dashboard.simonbrundin.com` svarade `504 Gateway time-out`
efter en lyckad OIDC-inloggning. Envoy-loggen visade:

```
GET / -> 504 response_timeout flags= UT dur= 30013 up= None
```

`upstream_service_time = None` - begäran kom aldrig till podden, utan den
stannade i Envoy. Samtidigt svarade `GET /favicon.ico` med 200 efter 16,5
sekunder. Appens logg var full av:

```
[Icon] loading icon `lucide:house` timed out after 1500ms
```

## Varför det hände

`@nuxt/icon` (v2) har en server-bundle: `serverBundle: { collections:
["lucide", "simple-icons"] }`. Men Nitro baktar inte in de importerade
JSON-samlingarna - de ligger kvar som externa moduler
(`import("@iconify-json/lucide/icons.json")`) och måste finnas i
**runner-steget** av imagen. I `environments/dev/Dockerfile` kopierades bara
`.output` och `package.json`, så ikon-API:et svarade 404 i produktion.
Utan lokal bundle föll SSR tillbaka mot Iconifys CDN, varje ikon kostade
upp till 1,5 s timeout, och vid nätverksproblem översteg
sidladdningen Envoy-route-timeoutet på 30 s.

## Lösningen

- `COPY --from=builder /app/node_modules/@iconify-json
  ./node_modules/@iconify-json` i runner-steget (båda Dockerfilerna).
- `timeouts.request: 30s -> 120s` på dashboardens HTTPRoute, så en tillfälligt
  långsam SSR ger långsammare svar i stället för 504.

## Viktigare fynd: dashboardens CI var trasig sedan 2026-10-21

Utan fungerande byggflöde hade ingen dashboard-image byggts på flera månader.
Två fel gjorde att varje körning föll:

1. Enhetstesterna importerar app-kod via Nuxt-aliasen `~/...` och relativa
   importer utan filändelse. Node kan inte resolvinga dem utan Nuxt, så
   `bun run test` föll med `Cannot find package '~'`. Lösning: en
   resolver-hook (`tests/nuxt-alias-loader.mjs` + `nuxt-alias-hooks.mjs`)
   som laddas via `node --import` i test-scriptet.
2. Byggsteget kopierade bara `app/`, `server/` och `public/` till
   byggkontexten. Nuxt 4 har även en `shared/`-katalog som `server/` och
   `app/` importerar, så bygget föll med `Could not resolve
   "../../shared/utils/modelEffort"`.

CI triggar nu även när workflow-filen ändras - annars hade fixen inte
kunnat testas.

## Lärdomar

- En `serverBundle` i `@nuxt/icon` kräver att ikonpaketen finns i
  runtime-imagen; kontrollera med `curl /_nuxt_icon/<collection>/<icon>.svg`
  mot podden innan man drar slutsatser om att ikonerna är inlagda.
- Ett trasigt byggflöde är ett säkerhetsproblem i sig: en gammal image
  ligger kvar i produktion utan att det syns.
- När SSR laddar externa resurser (ikoner, API:er) ska route-timeouten
  rymma deras worst case, annars blir nätverksglapp en 504.

## Relaterade

- [[notes/zitadel-importerade-anvandarens-losenordshash-stammde-inte-med-nagot-kant-losenord]]
