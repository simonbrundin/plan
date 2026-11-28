---
title: Arkitektur
description: Plan består av tre tjänster - en Nuxt-SPA, ett Go-API och en agenttjänst - som delar en PostgreSQL-databas och deployas till Kubernetes via Flux.
order: 1
icon: i-lucide-layers
updated: 2026-11-25
group: Översikt
---

# Arkitektur

Plan är ett personligt målsystem. Det består av tre tjänster som byggs
från samma repository men körs som separata deployables, plus en
PostgreSQL-databas.

```text
┌──────────────────────────────────────────────────────────────┐
│                        Användaren                            │
│                  (webbläsare, SPA)                            │
└────────────────────────────┬─────────────────────────────────┘
                             │ HTTPS
                             ▼
┌──────────────────────────────────────────────────────────────┐
│  Nuxt 4 (SPA)          plan.simonbrundin.com                │
│  - inga SSR (routeRules /**  ssr:false)                      │
│  - Pinia för tillstånd, Nuxt UI + Tailwind för utseende      │
│  - pratar direkt med API:t (ingen proxy)                    │
└───────────┬──────────────────────────────┬───────────────────┘
            │ sessioncookie / Bearer       │ intern nyckel
            ▼                              ▼
┌──────────────────────────┐   ┌──────────────────────────────┐
│  Go API (Gin)            │   │  Agent-API (Gin)             │
│  plan-api / port 8080    │   │  agent-api / port 8080        │
│  - äger mål och körningar│◄──┤  - definierar, planerar,     │
│  - godkännanden, usage   │   │    utför och verifierar      │
│  - OpenAPI (Scalar)      │   │  - chatt, verktyg, gates     │
└───────────┬──────────────┘   └───────────┬──────────────────┘
            │                              │
            │                    ┌─────────▼──────────┐
            │                    │  Worker (Job)      │
            │                    │  agent-worker-run  │
            │                    │  - isolerad kopia  │
            │                    │    av repot        │
            │                    └─────────┬──────────┘
            ▼                              ▼
┌──────────────────────────────────────────────────────────────┐
│  PostgreSQL                                                 │
│  - plan (mål, kalender, körningar)   - plan-agent (chatt)   │
│  - Neon i dev, CNPG i prod                                   │
└──────────────────────────────────────────────────────────────┘
```

## Tjänsterna

| Tjänst | Kod | Image | Port | Ansvar |
| --- | --- | --- | --- | --- |
| Gränssnitt | `src/nuxt` | `node:lts-alpine` | 3000 | SPA, anropar API:t direkt |
| Plan API | `src/api` | `gcr.io/distroless` | 8080 | Mål, kalender, agentkörningar, godkännanden, kostnad |
| Agent-API | `src/agent` | `plan-agent` | 8080 | Verktyg, chatt, gates, jobbhantering |
| Agent-worker | `src/agent/cmd/agent-worker` | `plan-agent` | - | Utför en körning i ett Kubernetes Job |

Båda Go-träden ligger i **en** Go-modul med rot i `src/` (modulnamn
`plan`): `src/api` ger `plan/api/...` och `src/agent` ger
`plan/agent/...`. Två binärer byggs ur modulen (`plan-api` och
`plan-agent`), men de är fortfarande två tjänster med egna images,
egna databaser och egna deploymenter - de pratar via det
versionerade kontraktet i `src/contracts/plan-agent.yaml`.

Gränssnittet och Plan-API:t är byggda med Bun, men körs i produktion på
Node.js - Bun gav problem med HTTP-svar i produktion, så runtime byttes
i PR #72. `bun.lock` måste hållas synkad med `package.json` eftersom
Dockerfilen kör `bun install --frozen-lockfile`.

## Repository-layout

```text
src/
├── nuxt/            Gränssnitt (Nuxt 4, SPA)
│   ├── app/
│   │   ├── pages/       Sidor: inbox, mål, kalender, agent ...
│   │   ├── components/  UI-komponenter
│   │   ├── composables/ API-klienter och tangentbevakare
│   │   ├── content/docs/ Dokumentationens markdown-filer
│   │   └── stores/      Pinia (goals)
│   └── server/      Proxy mot Go- och agent-API:t
├── api/             Plan API (Gin)
│   ├── cmd/api/     Startpunkt
│   ├── internal/
│   │   ├── handlers/    HTTP-handlers per domän
│   │   ├── middleware/  Auth, scopes
│   │   ├── database/    Anslutning och seed
│   │   └── pricing/     Prislista för agentkostnad
│   └── docs/        Genererad OpenAPI-spec
├── agent/           Agentmotorn
│   ├── cmd/agent-api/    Agent-API:t (långlivad)
│   ├── cmd/agent-worker/ Workerns ingång (kör en uppgift och avslutar)
│   ├── cmd/wikictl/      Verktyget för kunskapswikin
│   └── internal/
│       ├── agentworker/    Internt paket (namnet är oförändrat)
│       ├── definer/        Skriver definitionen för ett mål
│       ├── goaldefinition/ Bygger förslaget - fyller luckor, skriver aldrig över
│       ├── executor/       Utför en definierad uppgift
│       ├── proposer/       Föreslår kodändringar
│       ├── coder/          Klonar, ändrar, testar, skapar PR
│       ├── jobcontroller/  Skapar Kubernetes Jobb per körning
│       ├── profile/        coding- och life-profiler
│       ├── core/           Verktyg, gates, orchestrator, minne
│       ├── planclient/     Agentens klient mot Plan
│       └── wiki/           Kunskapswikin
├── contracts/       Kontraktet mellan Plan och agenten
└── go.mod           Gemensam Go-modul för API och agent
```

## Flöde för en vanlig uppgift

```text
1. Användaren skriver i snabb lägg till (n)   → webbläsaren → Go API
2. Go API skriver målet i goals               → PostgreSQL
3. Pinia uppdateras med det nya målet         → UI ritas om
```

Ingen komponent pratar direkt med databasen utöver Go API:t. Undantaget
är agenten, som går genom Plan API:t med sin egen nyckel - det gör att
Plan fortsätter att äga sanningen om mål och körningar även om
agenttjänsten är nere.

## Dataflöde för agentkörningar

```text
Användaren markerar mål för agent
   │
   ▼
Go API skapar AgentRun (status: queued)
   │
   ▼  jobbcontroller i agent-API:t
Kubernetes Job (agent-worker-run-<id>)
   │
   ├── define    - skriver målets definition
   ├── plan      - bryter ner i checkpoints
   ├── execute   - gör arbetet i en isolerad kopia av repot
   └── verify    - kör tester, rapporterar resultat
   │
   ▼
Plan API: status, events, usage, cost
   │
   ▼
Gränssnittet visar körningen och eventuella godkännanden
```

## Deployment

| Miljö | Kluster | Deploy | Database |
| --- | --- | --- | --- |
| dev | kind-dev | Tilt (`environments/dev/Tiltfile`) | Neon, branch `dev` |
| pr | Kubernetes | GitHub Actions + Flux | egen Neon-branch |
| prod | cluster1 (Talos) | Flux (GitOps) | CNPG i klustret |

Flux synkar `environments/*` till klustret. Hemligheter krypteras med
**SOPS + age** och dekrypteras av Flux - inga hemligheter finns i git.
Läs mer i [Utveckling och drift](/dokumentation/utveckling).

## Designprinciper

- **En ägare per data.** Plan API:t äger mål, körningar och godkännanden.
  Agenten och gränssnittet är klienter mot det.
- **Tunt lager i webbläsaren.** Affärslogik ligger i Go API:t. Nuxt
  vidarebefordrar och validerar inte på egen hand.
- **Kontraktet är versionsstyrt.** `src/contracts/plan-agent.yaml`
  beskriver faserna, tillstånden och anropen mellan Plan och agenten.
  Båda sidor testar mot samma fil - ändras den på ett sätt som kräver
  samtidig ändring på båda sidor, bumpas versionen.
- **Människan i loopen.** Agenten får ändra kod först efter att ett
  förslag har testats och en människa har godkänt det.
