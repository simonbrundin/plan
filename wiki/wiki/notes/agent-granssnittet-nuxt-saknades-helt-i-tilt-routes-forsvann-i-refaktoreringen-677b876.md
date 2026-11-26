---
title: Agent-gränssnittet (Nuxt) saknades helt i Tilt - routes försvann i refaktoreringen 677b876
type: note
tags: [agents, tilt, nuxt, bug, projekt:agents]
created: 2026-11-23
updated: 2026-11-23
status: active
related: [notes/agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev]
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
   agents/src/go/cmd/api/main.go tills commit 677b876 ("refactor: simplify to single
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
- Agents Tiltfile deployar inte längre Plan (beslut av Simon 2026-11-23).
  Plan är ett eget projekt med egen Tilt (plan/environments/dev, fasta portar
  3100/4100/4200). Istället väntar en local_resource "plan-api" (max 60 s,
  blockerar inte) på att go-api-service.plan får endpoints, och agent-API:t
  pratar med Plan via klusterinterna servicenamn.
  PLAN_WEB_URL ändrades till http://localhost:3100 - den används för länkar
  som visas för människor (Slack), så den ska peka på browsern, inte på
  det klusterinterna servicenamnet.

## Lärdom
- Två Tilt-instanser som deployar samma manifests (via kustomize) fightar om
  k8s-objekten och om image-tags. När agents-Tiltfile:en slutar deploya Plan
  raderar Tilt de objekt den tidigare applicerat - Plan's deployment raderades
  alltså ur klustret och Plan's egen Tilt återapplicerade dem inte (Tilt
  applicerar bara vid ändring). Plan's Tilt behövde startas om.
  Lärdom: kör INTE samma manifests från två Tilt-instanser.
- Ett Tilt-objekt som refererar en ConfigMap (envFrom) startas inte om när
  ConfigMap:en ändras - podden behöver raderas manuellt för att få nya env-värden.
- Tiltfile:ens `docker_build`-parametrar (t.ex. `dockerfile=`) är relativa till
  Tiltfile:ens katalog, INTE till buildcontexten. Med contexten "../../" (repo-roten)
  måste dockerfile-sökvägen vara "../../src/agent/Dockerfile", annars letar Tilt i
  environments/dev/ och Tiltfile:en failar att ladda (och Tilt fortsätter med
  den gamla konfigurationen utan att tydligt säga det).

## /readme och /roadmap (fix 2)
- Symptom: agent-API:t svarade 404 på /api/v1/readme och /api/v1/roadmap.
  Två orsaker: (1) README.md fanns inte i agent-imagen (buildcontexten var
  src/agent), (2) getProjectFile f resolves sökvägen via runtime.Caller, som ger
  den KOMPILERADE sökvägen (/home/simon/repos/agents/...) - den finns inte i
  containern.
- Fix: byggcontexten är nu repo-roten (Dockerfile: src/agent/Dockerfile) och
  README.md kopplas in i imagen. Root .dockerignore utesluter allt utom src/agent
  och README.md (node_modules m.m. är 1,5 GB). getProjectFile söker nu i
  ordningen AGENT_DOCS_DIR, PROJECT_ROOT, working directory, repo-rot via
  runtime.Caller - samma mönster som goals-handlern använder.
- ROADMAP.md togs bort i commit 212db36 ("chore: translate goals to Swedish
  and add missing items") - innehållet flyttades till goals.json. /api/v1/roadmap
  genererar därför nu markdown ur goals.json (statusöversikt + en sektion per
  status) i stället för att läsa en fil.

## CI (fix 3)
- .github/workflows/go.yml byggde ./cmd/agent, ./cmd/api och ./cmd/tui - de
  radrades i commit 677b876. Byggsteget är nu `go build -o /tmp/agent .` och
  ett nytt steg bygger imagen (`docker build -f src/agent/Dockerfile .`) så att
  trasiga Dockerfile/.dockerignore fångas i CI. Trigger-sökvägarna täcker nu
  också .dockerignore och README.md.
- Stale bygginstruktioner i README.md, src/nuxt/content/docs/running.md och
  src/nuxt/content/repo-map.md rättade (de pekade på ./cmd/* och src/agent).

## Dokumentationen (fix 4)
- src/nuxt/content/repo-map.md, docs/architecture.md, docs/components.md och
  docs/index.md beskrev fortfarande den gamla arkitekturen: "pi" som
  agent-harness, CLI/TUI som ingångar, ./cmd/agent|api|tui, katalogen
  src/agent/ och målen i "repo-roten". Allt är nu omskrivet mot kodens
  verkliga utseende (src/agent/, en ingång main.go, scheduler + agentworker,
  provider-prioritering, verktygslista, DB-tabeller, Nuxt-sidor och
  server-routs).
- Prod deployas med FLUX (inte ArgoCD) - användaren bekräftade detta.
  ArgoCD-spåret var dessutom dött: environments/prod/argocd-sync-options.yaml
  refererades av ingen kustomization (namespace "software-agent" som inte
  finns) och är borttagen. environments/prod/db/db.yaml har kvar en
  argocd.argoproj.io-annotation - den är inert för Flux men ligger kvar
  (att ta bort den i prod är en riskfråga, inte en ren dokumentationsfix).
- goals.json ligger i src/agent/goals.json (versionsspårad), inte i repo-roten.
  I containern ligger kopian på /app/goals.json (workdir) - det är därför
  /api/v1/goals hittar den via working directory-fallbacken.
- Dockerfile: CMD ["./agent", "-serve"] -> CMD ["./agent"] (main.go parsar
  inga flaggor - "-serve" var vilseledande).
- Viewer-websocketen i internal/core/orchestrator/viewer_websocket.go är inte
  exponerad som endpoint - dokumentationen säger nu det, och pekar på
  /api/v1/ws/chat som det enda exponerade websocketet.

## Kvar
- environments/prod/db/db.yaml har fortfarande en ArgoCD-annotation.
- Prod-ns heter agents-prod, men CNPG-klustret i environments/prod/db heter
  fortfarande software-agent-db (namnet på databasen är software-agent).
- src/nuxt/content/docs/design-tools.md och workflow.md är inte gåtts igenom
  på samma detaljnivå (innehållet i workflow.md stämmer med koden).

## Relaterade

- [[notes/agent-api-crashade-i-tilt-saknad-database-url-mot-cnpg-databasen-i-agents-dev]]

