---
title: Plan-repots CI- och prod-pipeline var trasig - fixade sökvägar och bygg
type: note
tags: [ci, deploy, pipeline, docker, kustomize, projekt:plan]
created: 2026-11-10
updated: 2026-11-27
status: active
related: [notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt, notes/delivery-pipeline-for-plan-fran-dev-till-prod-ar-nu-dokumenterad]
---

Pushar till main i plan-repot räckte inte till prod - tre oberoende
fel gjorde att prod aldrig uppdaterades (2026-11-10).

1) Nuxt-bilden gick inte att bygga i CI. node:lts-alpine har flyttat
   sig till Node 26, och better-sqlite3 (ett peer-dependency som bun
   auto-installerar) kompilerar inte mot Node 26:s V8-API
   ("class v8::Object has no member named GetPrototype"). Samma Dockerfile
   byggde lokalt (Node 24) men inte i CI - ett klassiskt rörligt-tag-problem.
   Fix: pinna node:24-alpine + "bun install --frozen-lockfile --omit=peer".
   Peer-beroendena används inte av appen.

2) Prod-overlay: environments/prod pekade på ../common, men katalogen
   heter base (naming-skillnaden när common slogs ihop med base).
   "kustomize build environments/prod" misslyckades alltså helt - Flux
   kunde inte ens applicera något. Samma fel i environments/pr/template/db.

3) commit.nuxt.yaml skrev till environments/production/patches/deployment.yaml
   (katalogen heter environments/prod) -> image-uppdateringen kunde aldrig
   köra. commit.api.yaml byggde dessutom från src/api (som då bara innehåller
   en kvarvarande Dockerfile.dev - Go-koden låg i src/go), triggade på
   src/api/** och publicerade till fel bildnamn (plan-api istället för
   go-api).

   (Sedan dess har strukturen ändrats: src/api/Dockerfile.dev är borttagen
   och Plan-API:t ligger nu i src/api, medan agenten ligger i src/agent.)

Lärdom: en rörlig base-image-tagg (node:lts-alpine) gör byggen
icke-deterministiska - pinna versionen. Och verifiera att kustomize
overlayerna faktiskt byggs ("kustomize build environments/<env>") - de
kan ligga trasiga länge utan att någon märker, eftersom inget steg
failade tillsammans med pushen.

Efter fixen: CI bygger och publicerar båda bilderna, och bot-PR:arna
(111 = Nuxt, 112 = Go API) skapades korrekt och mergades - båda
uppdaterar prod-manifesten. Prod-klustret (hemlabet) var dock
otillgängligt från laptopen, så själva rollouten kunde inte verifieras.

## Relaterade

- [[notes/delivery-pipeline-for-plan-fran-dev-till-prod-ar-nu-dokumenterad]]
- [[notes/authflytten-i-plan-sessionen-ligger-nu-i-go-api-t-inte-i-nuxt]]
