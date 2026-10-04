# Development Environment

Plan använder Tilt för att hantera en komplex utvecklingsmiljö med flera worktrees, Kubernetes-resurser och en delad dev-gateway.

## Arkitektur

```
┌─────────────────────────────────────────────────────────────────────┐
│                          DEV GATEWAY (Caddy)                        │
│                     http://localhost:3000                            │
│                                                                      │
│  Routes:                                                             │
│    /session/*      → Nuxt dev server (per worktree, olika portar)    │
│    /session        → Nuxt dev server                                 │
│    /oauth/callback → Go API (per worktree, via wt-cookie)           │
│    /health         → 200 OK                                          │
└─────────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│ Nuxt (session)│    │ Go API        │    │ PostgreSQL    │
│ :3172         │    │ :4152        │    │ :5432         │
└───────────────┘    └───────────────┘    └───────────────┘
```

## Två Lägen

### 1. Lokalt Docker-läge (`Tiltfile.local`)

Alla tjänster kör lokalt med Docker Compose:
- **Gateway**: Caddy i Docker
- **Nuxt**: Lokalt via `npm run dev`
- **Go API**: Lokalt via `go run`
- **PostgreSQL**: Docker Compose

```bash
# Starta
cd environments/dev
tilt up

# Eller explicit:
tilt up -- --mode=local
```

### 2. Kubernetes-läge (`Tiltfile.k8s`)

Dev-miljö i Kubernetes med live-update:
- **Images**: Byggs och pushas till ttl.sh
- **Resources**: CNPG för PostgreSQL, Custom deployments för app/api
- **Live Update**: Kod ändras synkas direkt till pods

```bash
# Starta
cd environments/dev
tilt up -- --mode=kubernetes

# Eller använd den dedikerade K8s-Tiltfile:
tilt -f Tiltfile.k8s up
```

## Flera Worktrees

Tanken är att kunna köra flera worktrees parallellt, t.ex.:
- `prod-plan` för produktionsfix
- `feat/new-feature` för nya features

### Port-allokering

Varje worktree får deterministiska portar baserat på namnet:

| Worktree    | Nuxt  | Go API |
|-------------|-------|--------|
| session     | 3172  | 4152   |
| prod-plan   | 3173  | 4153   |
| feat-xyz    | 3174  | 4154   |

### Routing via Dev Gateway

Gatewayen (Caddy) routar trafik baserat på path:

```
http://localhost:3000/session/     → Nuxt på :3172
http://localhost:3000/prod-plan/     → Nuxt på :3173
http://localhost:3000/feat-xyz/      → Nuxt på :3174
```

## OAuth med Worktrees

### Problemet

Zitadel (OAuth-provider) kräver en whitelistad redirect-URL. Med många worktrees på olika portar blir det snabbt ohanterligt.

### Lösningen: Centraliserad OAuth via Gateway

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Användare klickar "Logga in" i worktree "session"          │
│ 2. Go API genererar OAuth-URL med redirect till gateway:         │
│    https://auth.simonbrundin.com/authorize?...                 │
│    &redirect_uri=http://localhost:3000/oauth/callback          │
│ 3. Go API sätter cookie: wt=session                            │
│ 4. Zitadel redirectar till gateway                              │
│ 5. Gateway läser wt-cookie och proxyar till rätt worktree       │
└─────────────────────────────────────────────────────────────────┘
```

**Key insight**: Alla worktrees använder SAMMA redirect-URL (`http://localhost:3000/oauth/callback`). Worktree-cookien (`wt`) talar om för gatewayen vilken backend som ska ta emot callback.

### Registrera en ny Worktree

```bash
./gateway/register-worktree.sh add <worktree-name> <nuxt-port> <api-port>

# Exempel:
./gateway/register-worktree.sh add session 3172 4152
./gateway/register-worktree.sh add prod-plan 3173 4153
```

Detta lägger till routes i Caddy:
- `/<worktree>/*` → Nuxt dev server
- `/oauth/callback` (med `wt=<worktree>` cookie) → Go API

## Miljövariabler

### Nuxt
```bash
DATABASE_URL=postgresql://plan:plan@localhost:5432/plan
NUXT_SESSION_PASSWORD=<secret>
NUXT_PUBLIC_APP_URL=http://localhost:3000/session
NUXT_PUBLIC_GO_API_URL=http://localhost:4152/api/v1
```

### Go API
```bash
DATABASE_URL=postgresql://plan:plan@localhost:5432/plan
PORT=4152
APP_DOMAIN=localhost:3000
ZITADEL_DOMAIN=auth.simonbrundin.com
ZITADEL_CLIENT_ID=<id>
WORKTREE_NAME=session
```

## Vanliga Problem

### "No such file or directory: nuxt"

`npm install` kördes inte. Lösning:
```bash
cd src/nuxt
npm install
```

### Gateway returnerar tom respons

Caddy-containern kan inte nå host-services. Kontrollera:
1. Gateway kör med `network_mode: host` i docker-compose
2. Routes pekar på `127.0.0.1:<port>`

### OAuth redirect-fel

1. Verifiera Zitadel har korrekt redirect-URL: `http://localhost:3000/oauth/callback`
2. Kontrollera `wt`-cookie sätts korrekt i browser

## Struktur

```
environments/dev/
├── Tiltfile           # Huvud-Tiltfile (väljer local/k8s)
├── Tiltfile.local     # Lokalt Docker-läge
├── Tiltfile.k8s       # Kubernetes-läge
├── docker-compose.yml  # Backend (postgres, pgview)
├── gateway/
│   ├── Caddyfile      # Caddy-konfiguration
│   ├── docker-compose.yml
│   └── register-worktree.sh
├── db/
│   └── schema.sql
└── kubernetes/
    └── overlays/dev/  # K8s manifests för dev
```

## Nästa Steg

- [ ] Lägg till HTTPS-stöd i dev-gateway via Let's Encrypt
- [ ] Automatisk worktree-registrering vid `tilt up`
- [ ] Delad schema.sql mellan local och K8s
