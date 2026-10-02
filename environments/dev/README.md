# Plan local development environment

This directory contains the local-mode Tilt configuration and the **shared
dev gateway** that lets multiple git worktrees run side-by-side against
the same Zitadel OAuth application.

## Architecture

```
        Browser
           │
           │ http://localhost:3000/<worktree>/
           ▼
┌──────────────────────────────────────────────────────────┐
│  Caddy dev gateway (Docker, :3000)                        │
│   /<worktree>/*      → Nuxt  (per-worktree port)         │
│   /oauth/callback    → Go API (per-worktree port,         │
│                         selected via wt=<name> cookie)    │
└──────┬───────────────────────────────────┬─────────────────┘
       │                                   │
       ▼                                   ▼
┌─────────────────────┐          ┌──────────────────────┐
│  Nuxt dev server    │          │  Go API              │
│  :3100+hash         │ ──proxy──▶ :4080+hash            │
│  (Vue + Nitro)     │   via    │  (Gin + OAuth)       │
│                    │   Bearer │                      │
└──────────┬──────────┘   token  └──────────┬───────────┘
           │                              │
           └──────────────┬───────────────┘
                          ▼
                 ┌────────────────────┐
                 │  PostgreSQL        │
                 │  (Docker, :5432)  │
                 └────────────────────┘
```

Per-worktree port assignment is **deterministic** — the same worktree
always gets the same port across restarts, so cookies and bookmarks
survive `tilt down`.

```
hash = sum(ord(c) for c in WORKTREE_NAME)
Nuxt port = 3100 + hash % 100
API port  = 4080 + hash % 100
```

## Why a shared gateway?

Without one, every worktree needs its own Zitadel redirect URI registered
in the Zitadel admin UI, and port collisions make it impossible to run
more than one worktree at a time.

With the gateway:

- A single redirect URI: `http://localhost:3000/oauth/callback`
  is registered in Zitadel.
- The worktree identity is carried in a `wt=<name>` cookie that Go API
  sets when the OAuth flow starts.
- Caddy matches that cookie on `/oauth/callback` and routes the
  callback to the correct worktree's Go API.

## OAuth flow

1. User visits `http://localhost:3000/<worktree>/login`.
2. Caddy routes `/<worktree>/*` to Nuxt.
3. User clicks "Logga in med Zitadel" → Nuxt redirects to
   `http://localhost:<API_PORT>/api/v1/auth/login` (Nuxt
   server route `/auth/login`).
4. Go API builds the Zitadel auth URL with:
   - `redirect_uri=http://localhost:3000/oauth/callback`
   - PKCE `code_challenge` and `state`
   - Sets cookie `wt=<worktree>` so the gateway can route the callback
   - Sets cookie `oauth_state` (with verifier + worktree)
   - Redirects to Zitadel.
5. Zitadel authenticates and redirects to
   `http://localhost:3000/oauth/callback?code=...&state=...`.
6. Caddy matches path `/oauth/callback` AND cookie `wt=<worktree>` →
   routes to `127.0.0.1:<API_PORT>` (this worktree's Go API).
7. Go API exchanges `code` for tokens, fetches userinfo,
   `LookupOrCreateUser` + `CreateSession`, then redirects to
   `http://localhost:3000/<worktree>/api/auth/callback?session=...`.
8. Caddy routes `/<worktree>/*` to Nuxt. Nuxt
   `server/api/auth/callback.get.ts` sets `plan_session` + `plan_user`
   cookies and redirects to `/`.

## Files

- `Tiltfile` — dispatches to `Tiltfile.local` or `Tiltfile.kubernetes`.
- `Tiltfile.local` — runs Postgres + pgview via Docker Compose, the
  gateway via Docker Compose, Go API + Nuxt as `local_resource`s.
- `docker-compose.backend.yml` — Postgres + pgview containers.
- `gateway/` — the Caddy dev gateway.
  - `Caddyfile` — minimal HTTP listener on :3000. Per-worktree routes
    are added at runtime.
  - `Dockerfile` — `caddy:2` image.
  - `docker-compose.yml` — gateway container (ports 3000 and 2019).
  - `register-worktree.sh` — CLI to add/remove worktree routes via
    Caddy's admin API. Used by Tilt on startup.

## Setup (one-time)

1. Add a single OAuth redirect URI in the Zitadel application
   (`391431970426454645`):
   ```
   http://localhost:3000/oauth/callback
   ```
   This is the **only** redirect URI you register — all worktrees share it.

2. Start the gateway (once, persistent across `tilt down`):
   ```bash
   cd environments/dev/gateway
   docker compose up -d
   ```

## Running

```bash
cd environments/dev
tilt up
```

By default the worktree name is `session`. To run another worktree in
parallel:

```bash
TILT_WORKTREE=other tilt up -- --port=10351   # different Tilt UI port
```

Each Tilt instance will:

1. Start its own Go API on a deterministic port (e.g. `:4117`).
2. Start its own Nuxt on a deterministic port (e.g. `:3145`).
3. Register itself with the shared gateway via `register-worktree.sh`.
4. Be reachable at `http://localhost:3000/<worktree>/`.

Stop everything:

```bash
tilt down                       # stops worktree-specific resources
cd gateway && docker compose down  # stops gateway (only if you want it gone)
```

## OAuth login from a fresh browser

1. Open `http://localhost:3000/<worktree>/login`.
2. Click "Logga in med Zitadel".
3. Zitadel login page → authenticate.
4. Browser lands back on `http://localhost:3000/<worktree>/` logged in.

## Troubleshooting

### "duplicate ID" errors in Caddy logs

The gateway container reused an `autosave.json` from a previous run with
routes that were never cleaned up. Reset with:

```bash
docker compose down -v
docker compose up -d
```

### `Container plan-dev-postgres ... already in use`

A leftover container from a previous Tilt run. Remove it:

```bash
docker stop plan-dev-postgres && docker rm plan-dev-postgres
```

### Login redirect goes to the wrong worktree

Make sure the gateway admin API doesn't have stale routes from a
previous worktree. Inspect with:

```bash
curl -s http://localhost:2019/config/apps/http/servers/srv0/routes \
  | ./gateway/register-worktree.sh list
```

### Cookies not being sent

Browsers treat `localhost` cookie scope by host (not port). If you are
hitting Caddy at `:3000` and Nuxt at `:3172`, cookies set by Nuxt do
not flow back through Caddy unless they are scoped to `path=/`. Both
`plan_session` and `plan_user` are set with `path=/`.

### Zitadel rejects the redirect URI

Make sure the URI in the Zitadel application matches exactly:
```
http://localhost:3000/oauth/callback
```
(Schemal, host, port, path — all four.)

## Manual gateway operations

```bash
# Add a worktree manually (e.g. when not using Tilt)
./gateway/register-worktree.sh add <name> <nuxt_port> <api_port>

# Remove a worktree
./gateway/register-worktree.sh remove <name>

# List all registered routes
./gateway/register-worktree.sh list
```