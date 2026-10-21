# Agent Preferences

This file contains preferences and requirements for AI coding agents working on this project.

## Runtime Preferences

### Nuxt Application
- **Use Bun for runtime**, NOT Node.js
- The Nuxt app uses `nitro preset: "node-server"` (for the server-side code), but runs on Bun (`oven/bun:1-slim` container)
- Dockerfile MUST use `oven/bun:1-slim` for production runtime, not `node:lts-alpine`

## Important Context

### Authentication Fix (fix-goals-401-auth-error)
- Zitadel returns opaque/JWE tokens that the JWT library can't parse
- Solution: Use session tokens with `introspectToken()` for Zitadel introspection
- Session tokens use HMAC-SHA256 signing with `SESSION_SECRET`
- Users auto-provisioned via `LookupOrCreateUser()` on first login

### Current Production Issues
- plan.simonbrundin.com not responding to HTTP requests (upstream timeout)
- Readiness probe failing: `/api/health` endpoint not responding inside Kubernetes
- Bun runtime may have issues with HTTP responses in production
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
| **Frontend** | Nuxt 4 (SPA)    | `oven/bun:1-slim`  | 3000 | Interface, SPA routing      |
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
Wikin ligger i `~/wiki`.

**Två regler:**

1. **Läs wikin innan du frågar användaren.** Svaret finns ofta redan
   dokumenterat - leta först.
2. **Skriv en not i wikin inbox när något stort händer.** Incident, beslut,
   ny feature eller en icke-obvious bugg.

```bash
wikictl search "bigram"                  # sök i hela wikin
wikictl search "fel" --project <repo>    # sök inom ett projekt
wikictl note --title "Rubrik" --tags "bug" --body "Vad som hände och varför."
wikictl ingest                          # flytta inboxen in i grafen
```

Anteckningar taggas automatiskt med `projekt:<repo>`.
Schemat för wikin står i `~/wiki/AGENTS.md`.
<!-- wiki:end -->
