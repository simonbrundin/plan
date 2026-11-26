---
title: Agent-gränssnittet (Nuxt) saknades helt i Tilt - routes försvann i refaktoreringen 677b876
type: note
tags: [agents, tilt, nuxt, bug, projekt:agents]
---

## Symptom
Tilt (environments/dev i agents-repot) visade bara Plan-appen. Agent-gränssnittet
fanns aldrig som Tilt-resurs: `docker_build("agents-nuxt", ...)` saknades helt i
Tiltfile:en, trots att environments/README.md och environments/base redan
beskrev en `agents-nuxt`-deployment.

## Varför
Två separata glapp:

1. Tiltfile:en deployade Plan (från plan-repots environments/dev), agent-API:t
   och agent-databasen - men aldrig agents-repots egen Nuxt. resources/base
   innehöll agents-nuxt, men dev-overlay applicerades aldrig av Tilt.
2. Agent-API:t saknade routes som frontend:an anropar. De fanns i
   src/agent/cmd/api/main.go tills commit 677b876 ("refactor: simplify to single
   entry point - API server only") raderade cmd/ och flyttade routerna - men
   flyttningen till internal/api/server.go glömde bort:
   PATCH /goals/:id, POST /goals/start-agent, GET /readme, GET /roadmap,
   GET /ws/chat, POST /tools/cancel, POST /tools/cancel-all, GET /tools/status.
   Chatten (WebSocket) var alltså död kod i API:t.

## Fix
- Tiltfile: bygger och deployar agents-nuxt i default-namespace, port-forward
  + länk "Agent Interface". Namespace agents-dev ägs inte av Tilt (där ligger
  CNPG-databasen, appliceras via kubectl så att den inte raderas av tilt down).
- server.go: routes återkopplade (samma wiring som i gamla cmd/api/main.go).
- Ny composable useApiBase(): browser och SSR kan inte använda samma URL.
  Browsern når agent-API:t via port-forward (NUXT_PUBLIC_API_BASE), servern via
  klusterinternt DNS (NUXT_API_BASE_INTERNAL = http://agent.default.svc.cluster.local:8080).
  Mönstret är kopierat från plan-repot (NUXT_PUBLIC_GO_API_URL +
  NUXT_GO_API_INTERNAL_URL).

## Kvar
- Två Tilt-instanser körs parallellt (agents + plan) och BÅDA deployar Plan:s
  manifests via kustomize -> de fightar om samma k8s-objekt och olika image-tags.
  Bör lösas: antingen deployar agents-Tiltfile inte Plan, eller så körs bara
  ett Tilt.
- /readme och /roadmap svarar 404 i containern: README.md/ROADMAP.md finns inte
  i agent-imagen (build-contexten är src/agent) och getProjectFile f resolves via
  runtime.Caller (kompilerad sökväg).
- .github/workflows/go.yml bygger fortfarande ./cmd/agent, ./cmd/api, ./cmd/tui
  som inte finns sedan refaktoreringen.
