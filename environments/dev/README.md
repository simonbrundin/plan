# Plan dev-miljö

Dev-miljön körs i ett **kind-kluster** (namespace `plan`) via Tilt.
Tilt bygger images, applicerar kustomize-manifesten och port-forwardar
tjänsterna till fasta portar på din laptop.

## Adresser

| Tjänst  | URL                        | Kommentar                          |
| ------- | -------------------------- | ---------------------------------- |
| Nuxt    | http://localhost:3100      | Appen - logga in här               |
| Go API  | http://localhost:4100      | REST-API, `/health`, `/swagger`    |
| pgAdmin | http://localhost:4200      | Databasadmin                       |

Portarna är **fasta** (`Tiltfile`). Tidigare allokerades de dynamiskt
från 3000, vilket gav nya URL:er varje gång Tilt startade - det gjorde
det omöjligt att ha en stabil OAuth-redirect-URI och bröt bookmarks och
cookies mellan omstarter.

## Arkitektur

```
        Browser
           │  http://localhost:3100
           ▼
┌─────────────────────┐   cookie + Bearer   ┌──────────────────────┐
│  Nuxt (Nitro)       │ ───────────────────▶ │  Go API (Gin)        │
│  :3100              │   tunn proxy        │  :4100               │
│  /api/proxy/*       │                     │  /api/v1/*           │
└──────────┬──────────┘                     └──────────┬───────────┘
           │                                           │
           │  sessionkoken i cookie                    ▼
           │  (sätts av Go)                   ┌────────────────────┐
           ▼                                 │  PostgreSQL (CNPG) │
    plan_session (httpOnly)                  │  plan-db-rw        │
    plan_user    (läsbar hint)               └────────────────────┘
```

### Authen ligger i Go API:t

Go äger hela inloggningen:

| Steg | Vem gör vad |
| ---- | ----------- |
| `/auth/login` (Nuxt) | Vidarebefordras till Go, som bygger Zitadel-URL:en och sätter `oauth_state` |
| Zitadel → `/api/v1/auth/callback` (Go) | Byter code mot token, skapar session, **sätter `plan_session` + `plan_user`** och redirectar till apparoten |
| `/api/proxy/*` (Nuxt) | Vidarebefordrar cookien till Go - Go validerar sessionen |
| `/api/v1/auth/session` (Go) | Returnerar användaren bakom sessionen |
| `/api/v1/auth/logout` (Go) | Raderar sessionen och rensar cookierna |

Sessionen godkänds via `plan_session`-cookien **eller** `Authorization: Bearer`
(API-klienter). Sessionstoken cirkulerar aldrig i en URL.

`plan_user` är en läsbar hint-cookie (sub + email) som bara gör att UI:t kan
rendera direkt - servern är alltid auktoritativ. Den sätts också av Go.

Nuxt:s proxy ligger i `server/utils/goApi.ts` och sköter tre saker:
vidarebefordrar `Cookie`, kopierar tillbaka `Set-Cookie` och `Location`, och
följer inte redirects själv (annars skulle OAuth-flödet köras server-side).

### URL:er

- Serverinternt: `NUXT_GO_API_INTERNAL_URL` (`http://go-api-service.plan:8080/api/v1`)
- Webbläsaren: `NUXT_PUBLIC_GO_API_URL` (`http://localhost:4100/api/v1`)

## Inloggning i dev

### Dev-login (rekommenderat, kräver ingen Zitadel)

Öppna http://localhost:3100/login och klicka **"Dev-login (utan Zitadel)"**.

Flödet:

1. Nuxt:s route `/auth/dev-login` anropar `POST /api/v1/auth/dev-login`
   på Go API:t (intern URL).
2. Go API:t skapar/findar användaren `dev-user` (`dev@plan.local`) och
   skapar en riktig session i tabellen `sessions`.
3. Nuxt sätter `plan_session` (httpOnly) + `plan_user` och redirectar till `/`.

Endpointen är bara aktiverad när `DEV_AUTH_ENABLED=true` (sätts i
`patches/api-deployment.yaml` för dev). I prod svarar den 404.

### Zitadel-inloggning

Fungerar bara om **Zitadel är nåbart** (`auth.simonbrundin.com` pekar på
hemlabet - laptop och kluster måste nå det nätverket) och redirect-URI:t
är registrerat i Zitadel:

```
http://localhost:4100/api/v1/auth/callback
```

Flödet: `/auth/login` (Nuxt) → Go bygger authorize-URL:en med
`redirect_uri` (= `ZITADEL_REDIRECT_URI`) och PKCE → Zitadel → callback
till Go → Go skapar sessionen, sätter cookierna och redirectar till
`FRONTEND_URL`.

`redirect_uri` måste vara **identisk** i authorize- och token-bytet,
annars avvisar Zitadel bytet. Båda läser `ZITADEL_REDIRECT_URI`.

Cookie-sättning: Go sätter `plan_session` för värdet `localhost` - porten
spelar ingen roll för cookies, så även om cookien sätts i svaret från
API:t (port 4100) skickas den till Nuxt (port 3100).

## Kör

```bash
cd environments/dev
tilt up
```

Stoppa:

```bash
tilt down
```

## Filer

- `Tiltfile` - bygger images, applicerar `kustomize(".")`, port-forwardar
  de fasta portarna.
- `kustomization.yaml` - dev-overlay: base + pgadmin + secrets + patches.
- `patches/nuxt-deployment.yaml` - env för Nuxt (URL:er, dev-auth).
- `patches/api-deployment.yaml` - env för Go API:t (dev-auth, OAuth-URL:er).
- `secrets/` - dev-secrets (DB-användare, auth).
- `pgadmin/` - pgAdmin för devdatabasen.
- `gateway/` - **används inte** i nuvarande läge. En gammal Caddy-gateway
  från en tidigare "local mode"-uppsättning med flera worktrees. Kan
  tas bort.

## Databasen

- CNPG-kluster `plan-db`, databasen `plan`, rollen `plan` äger tabellerna.
- Schemat appliceras av Atlas Operator (`AtlasSchema` `plan-db-schema`) -
  som rollen `plan`, så API:t har rätt att ändra tabellerna.
- **Atlas applicerar schemat deklarativt (bara DDL)** - `INSERT`-raderna i
  `schema.yaml` körs alltså *inte*. Applikationen seedar därför själv
  (statusarna + Root-målet + goals-sekvensen) i
  `src/go/internal/database/seed.go` vid start.
- `sessions`-tabellen ingår i `environments/base/db/schema.yaml`; annars
  vill Atlas radera den (den skapas också av API:t vid runtime).
- `GRANT`-satser får **inte** ligga i `schema.yaml` - Atlas (community)
  avvisar dem med "permissions are available to logged-in users only".
  Rättigheterna sätts i stället via `postInitSQL` i
  `environments/base/db/postgres-cluster.yaml`.
- Varning: en Atlas-apply är deklarativ. Ett schema som saknar tabeller
  **raderar** dem - testa aldrig genom att tömma `schema.yaml`.

## Felsökning

### "404 på /auth/login"

Kontrollera att du använder **http://localhost:3100** och inte en gammal
dynamiskt allokerad port. Port-forwarden ändras inte när Tilt startar om
- men den gamla porten dör med den gamla processen.

### 401 från API:t trots inloggad

Kontrollera i Go API:t:s logg: `ValidateSession failed: ...`. Om det
står "session expired" är det tokenets tidsstämpar som inte avkodas -
`signSession` och `Session`-structen måste använda samma format
(JSON-taggar + RFC3339).

### "Mål kan inte skapas" (duplicate key på goals_pkey)

Root-målet seedas med explicit `id = 1`, vilket lämnar identitetssekvensen
efter. Applikationen fixar sekvensen vid start (`seed.go`).

### AtlasSchema är "Stalled"

Kontrollera att `DATABASE_URL`/Atlas-URI:n använder **FQDN**
(`plan-db-rw.plan.svc.cluster.local`) - korta namn fungerar bara från
samma namespace som tjänsten.
