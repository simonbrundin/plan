# Agent Preferences

This file contains preferences and requirements for AI coding agents working on this project.

## Runtime Preferences

### Nuxt Application
- **Bun används för build och dependency-installation** (`bun install`, `bun run build`)
- **Runtime i produktion är Node.js** (`node:lts-alpine`) - ändrat i PR #72 eftersom Bun
  gav problem med HTTP-svar i produktion. Ändra inte tillbaka utan att verifiera.
- `bun.lock` måste hållas synkad med `package.json` - Dockerfile:n kör
  `bun install --frozen-lockfile`

## Important Context

### Auth (2026-11)
- **Authen ligger i Go API:t**, inte i Nuxt. Go skapar sessionen, sätter
  `plan_session` (httpOnly) + `plan_user` (läsbar hint) och validerar sessionen
  från cookie *eller* `Authorization: Bearer`.
- Nuxt är en tunn proxy (`server/utils/goApi.ts`): vidarebefordrar cookien och
  kopierar tillbaka `Set-Cookie` + `Location` från Go.
- Sessionstoken får inte cirkula i URL:er - det var den gamla callback-flödet.

### Current Production Issues
- plan.simonbrundin.com not responding to HTTP requests (upstream timeout)
- Readiness probe failing: `/api/health` endpoint not responding inside Kubernetes
- GitHub Actions cache causing stale builds with missing dependencies

## Secrets
- SESSION_SECRET: `3eDqjso1WefD4doxeG2nR4ZQvg4SjOMe669ac4B8UQE=`
- Zitadel CLIENT_SECRET: `pba!gvj5dqk7hut@NYR` (in 1Password item "Auth - Zitadel")

## Cluster Access
- Use certificate login: `KUBECONFIG=/tmp/kubeconfig-certificate kubectl...`
- Vault is down (no pods in vault namespace) - ESO can't sync secrets
- Auth secret patched manually as workaround

## Stack

| Layer        | Technology       | Docker Image        | Port | Purpose                     |
| ------------ | --------------- | ------------------- | ---- | --------------------------- |
| **Frontend** | Nuxt 4 (SPA)    | `node:lts-alpine`  | 3000 | Interface, SPA routing      |
| **API**      | Go + Gin        | `gcr.io/distroless` | 8080 | REST API, business logic    |
| **Database** | PostgreSQL      | CNPG               | 5432 | Data storage (Kubernetes)  |
| **Schema**   | Atlas           | Operator           | -    | Schema migration (K8s)     |

## Communication

### Frontend → API

```typescript
// Use runtimeConfig for API URL
const config = useRuntimeConfig();
const data = await fetch(`${config.public.apiBase}/endpoint`);
```

### API → PostgreSQL

```go
// Standard database/sql with pq driver
db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

## Environment Variables

| Variable               | Nuxt | Go  | Description                  |
| ---------------------- | ---- | --- | ---------------------------- |
| `DATABASE_URL`         | -    | ✅  | PostgreSQL connection string |
| `NUXT_PUBLIC_API_BASE` | ✅   | -   | URL to Go API                |
| `PORT`                 | -    | ✅  | API port (default: 8080)     |

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Browser   │────▶│  Nuxt SPA   │────▶│   Go API    │
│             │◀────│             │◀────│             │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                                │
                                                ▼
                                        ┌─────────────┐
                                        │  PostgreSQL │
                                        └─────────────┘
```

## Development

### Run locally

```bash
# Nuxt dev server
cd src/nuxt && bun run dev

# Go API
cd src/go && go run ./cmd/api

# PostgreSQL (if local)
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=secret postgres:16
```

### Database Schema

Schema is managed declaratively via Atlas Operator:

- `environments/base/db/schema.sql` - SQL definition
- `environments/base/db/atlas-schema.yaml` - Atlas CRD

Changes to schema start in `environments/base/db/schema.sql`, then are tested in dev/pr, and finally applied to production via Flux.

<!-- wiki:start -->
## Kunskaps-wiki

Det här repot delar en gemensam kunskaps-wiki med alla andra repor.
Wikin ligger i `~/wiki` och är en persistent kunskapsgraf som
agenterna underhåller - kunskapen härleds inte om vid varje fråga.

**Fem regler:**

1. **Läs wikin innan du frågar användaren.** Svaret finns ofta redan
   dokumenterat - leta först.
2. **Skriv en not i wikin inbox när något stort händer.** Incident, beslut,
   ny feature eller en icke-obvious bugg.
3. **Lägg in råkällor** du lär dig av med `wikictl source add`.
   De är immutabla och ger spårbarhet.
4. **Håll wikin aktuell.** Ett påstående som en ny källa motsäger ska
   ändras, inte ligga kvar. `wikictl lint` hittar trasiga länkar,
   ödelappar och föråldrad info.
5. **Fila tillbaka värdefulla svar** som sidor, så att de inte försvinner
   i chatthistoriken.

```bash
wikictl search "bigram"                  # sök i hela wikin
wikictl search "fel" --project <repo>    # sök inom ett projekt
wikictl note --title "Rubrik" --tags "bug" --body "Vad som hände och varför."
wikictl page --type concept --title "Rubrik" --body "..."   # skriv en sida
wikictl link --from concepts/a --to concepts/b               # korslänka
wikictl source add --file kalla.md --title "Rubrik"         # råkälla
wikictl ingest                          # flytta inboxen in i grafen
wikictl lint                            # hälsa
```

Anteckningar taggas automatiskt med `projekt:<repo>`.
Schemat för wikin står i `~/wiki/AGENTS.md`.
<!-- wiki:end -->
