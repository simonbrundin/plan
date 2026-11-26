---
title: Utveckling och drift
description: Köra Plan lokalt, köra testerna, ändra schemat och deploya - steg för steg.
order: 7
icon: i-lucide-wrench
updated: 2026-11-25
group: Utvecklare
---

# Utveckling och drift

## Lokalt

### Gränssnitt och API

```bash
# Nuxt dev server (port 3000)
cd src/nuxt && bun run dev

# Go API (port 8080)
cd src && go run ./api/cmd/api
```

Dev kör **inte** mot en Postgres i klustret. `environments/dev/db/neon.env`
(gitignoras) innehåller anslutningssträngarna till Neon-branch `dev` och
blir Secret `plan-db-neon` via kustomize.

### Hela stacken med Tilt

Plan och agenten deployas från **en** Tilt - två Tilt-instanser som
deployar samma manifests fightar om k8s-objekten.

```bash
cd environments/dev
tilt up
```

Tilt bygger tre images:

| Image | Byggcontext | Vad det är |
| --- | --- | --- |
| `plan-nuxt` | `src/nuxt` | Plan-gränssnittet |
| `plan-go-api` | `src` | Plan API - äger mål, körningar, godkännanden |
| `plan-agent` | repo-roten | Agent-API:t och workern |

| Tjänst | Port i dev |
| --- | --- |
| Plan (Nuxt) | http://localhost:3100 |
| Plan API | http://localhost:4100/api/v1 |
| Agent-API | http://localhost:8080 (port-forward) |

## Tester

```bash
# Gränssnittet - Cucumber-funktionstester
cd src/nuxt && bun run test

# Enhetstester (Bun)
cd src/nuxt && bun run test:unit

# Typkontroll
cd src/nuxt && bun run typecheck

# Plan API
cd src && go test ./api/...

# Agenten
cd src && go test ./agent/...
```

CI kör samma tester via `.github/workflows/ci.yaml`,
`acceptance.yaml` och `agent-go.yml`. Commit-meddelandet styr
versionssättningen av de tre delarna (semantic-release):

| Fil | Del |
| --- | --- |
| `commit.nuxt.yaml` | Gränssnittet |
| `commit.api.yaml` | Plan API |
| `commit.agent.yaml` | Agenten |

## Bygga

```bash
cd src/nuxt && bun run build
```

`bun.lock` måste hållas synkad med `package.json` - Dockerfilen kör
`bun install --frozen-lockfile`. Runtime i produktion är Node.js
(`node:lts-alpine`), inte Bun: Bun gav problem med HTTP-svar i
produktion, vilket ledde till bytet i PR #72.

## Ändra databasen

1. Ändra SQLen i `environments/base/db/schema.yaml` (nyckeln
   `schema.sql`).
2. Deploya till dev och testa.
3. Samma för PR-miljön.
4. Låt Flux applicera den deklarativa schemat i produktion.

Atlas applicerar DDL - seeden (statuser och rotmålet) ligger i
`internal/database/seed.go` och körs av applikationen vid start.

## Deploy

| Miljö | Kluster | Deploy | Database |
| --- | --- | --- | --- |
| dev | kind-dev | Tilt | Neon, branch `dev` |
| pr | Kubernetes | GitHub Actions + Flux | egen Neon-branch |
| prod | cluster1 (Talos) | Flux (GitOps) | CNPG i klustret |

Flux synkar `environments/*` till klustret. Det finns ingen
manuell deploy i produktion - en merge till `main` är allt som behövs.

## Hemligheter

Inga credentials i repot. Hemligheter krypteras med **SOPS + age** och
dekrypteras av Flux (`spec.decryption.provider: sops`).

| Plats | Innehåll |
| --- | --- |
| `environments/prod/secrets/*.enc.yaml` | Krypterade hemligheter för prod |
| `environments/base/sops-public-key.yaml` | Publika age-nyckeln |
| `~/.config/sops-age/keys.txt` | Privata nyckeln (aldrig commit) |

Övriga hemligheter (Zitadel `CLIENT_SECRET`, `SESSION_SECRET`) finns i
1Password och hämtas med:

```bash
op read "op://Private/<item>/<field>"
```

Git-hemligheten för agentens Pull Requests (`GIT_AUTH_HEADER`) skapas
utanför git:

```bash
TOKEN=$(gh auth token)
HEADER="Authorization: Basic $(printf 'x-access-token:%s' "$TOKEN" | base64 -w0)"
kubectl --context kind-dev -n plan create secret generic plan-agent-git \
  --from-literal=GIT_AUTH_HEADER="$HEADER"
```

## Klusteråtkomst

```bash
kubectl --context admin@cluster1 get pods -n plan
```

**Kör aldrig `talosctl kubeconfig` rakt in i `~/.kube/config`** - det
mergar in och skapar `cluster1-N`-dubbletter med olika CA. Använd
`talos-kubeconfig` (i `~/.local/bin`) som hämtar med `--merge=false`, eller
`kubectl --context admin@cluster1` direkt.

## Kunskapswikin

Wikin är gemensam för alla repon och ligger i `wiki/`. Bygg verktyget
efter ändringar:

```bash
cd src && go build -o ~/.local/bin/wikictl ./agent/cmd/wikictl
```

```bash
wikictl search "bigram"        # sök
wikictl note --title "Rubrik" --tags "bug" --body "Vad som hände."
wikictl ingest                 # flytta inboxen in i grafen
wikictl lint                   # hitta trasiga länkar och gammal info
```

Reglerna för wikin - råkällor, syntes, schema - finns i `wiki/AGENTS.md`.

## Konventioner

- **Bun för bygg och installation**, Node.js som runtime i produktion.
- **Hjälpmodalen uppdateras alltid** när en tangent ändras -
  `src/nuxt/app/components/HelpModal.vue` är det enda stället där
  användaren ser vilka tangenter som fungerar. Modalen är data-driven:
  nya tangenter läggs in i `sections` i script-blocket.
- **Skriv varför, inte bara vad**, i noterna till wikin. Koden visar
  vad - noten ska visa varför.
- **Verifiera innan du deklarerar klart.** Kör testerna, kolla
  loggarna, jämför före/efter.

## Se också

- [Arkitektur](/docs/arkitektur) - tjänsterna och deras ansvar.
- [API och autentisering](/docs/api) - endpoints, sessioner, scopes.
- [Agentsystemet](/docs/agent-systemet) - faserna och profilerna.
