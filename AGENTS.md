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
- Inga credentials i detta repo. Hemligheter krypteras med **SOPS + age**
  och dekrypteras av Flux (`spec.decryption.provider: sops`).
- `environments/prod/secrets/*.enc.yaml` - krypterade, appliceras av Flux.
- `environments/base/sops-public-key.yaml` - publika age-nyckeln (samma som
  `infrastructure-flux/components/secrets/sops.age.pub`).
- Privata nyckeln ligger lokalt i `~/.config/sops-age/keys.txt` (aldrig commit).
- Övriga hemligheter (Zitadel CLIENT_SECRET, SESSION_SECRET) finns i
  1Password - hämta med `op read "op://Private/<item>/<field>"`.

## Cluster Access
- Kubectl: `kubectl --context admin@cluster1 ...` (kubeconfig i
  `~/.kube/config`). `/tmp/kubeconfig-certificate` fungerar inte - det
  pekar mot en gammal CA. Kontexten `admin@cluster1` pekar mot VIP
  `https://10.10.10.10:6443` med aktuell CA.
- **Kör ALDRIG `talosctl kubeconfig` rakt in i `~/.kube/config`** - det mergar
  in och skapar `cluster1-N`-dubbletter med olika CA. Använd
  `talos-kubeconfig` (i `~/.local/bin`) som hämtar med `--merge=false` och
  installerar `admin@cluster1` på rätt sätt, eller
  `talosctl kubeconfig /tmp/kc.yaml --merge=false` och kopiera in CA+cert
  manuellt. Shell-guarden i `dotfiles/bash/.bashrc` tvingar `--merge=false`.
- `talosctl`: `talosctl -n <nod-ip> ...` med `~/.talos/config`.
- **Ingen Vault och ingen ESO (External Secrets Operator)** i klustret - det
  var en ArgoCD-era-modell. SOPS via Flux är enda secret-väg. Se
  `~/wiki/wiki/concepts/sops.md`.
- Klustret (Talos) byggdes om 2026-11-23: FluxOperator måste installeras
  manuellt i ett nytt kluster, och noder med trasig pod-nätverk
  (controlplane-2, worker-8) har taint `node.kubernetes.io/broken`.

## Stack

| Layer        | Technology       | Docker Image        | Port | Purpose                     |
| ------------ | --------------- | ------------------- | ---- | --------------------------- |
| **Frontend** | Nuxt 4 (SPA)    | `node:lts-alpine`  | 3000 | Interface, SPA routing      |
| **API**      | Go + Gin        | `gcr.io/distroless` | 8080 | REST API, business logic    |
| **Database** | PostgreSQL      | Neon (dev) / CNPG (prod) | 5432 | Data storage           |
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

### Kortkommandon - uppdatera ALLTID hjälpmodalen (viktigt)

**Varje gång ett kortkommando läggs till, tas bort eller ändras måste
`src/nuxt/app/components/HelpModal.vue` uppdateras i samma PR.** Hjälpmodalen
är det enda stället där användaren kan se vilka tangenter som fungerar, så en
ny tangent som inte finns i modalen är osynlig för användaren.

Var tangenterna ligger just nu:

| Tangenter | Fil |
| --- | --- |
| Globala (`?`, `Ctrl+K`, `n`, `z`, `p`, `Space`) | `app/layouts/default.vue` |
| Leader-läge + globala leader-kommandon | `app/composables/useLeaderKey.ts` |
| Leader-kommandon per sida | `app/pages/goal/[id].vue`, `app/pages/goals/index.vue`, `app/pages/now.vue` |
| Målsidan (d, e, t, c, b, m, r, x, p, i, a, n, j, k, h, l) | `app/pages/goal/[id].vue` + `app/composables/useGlobalNavigation.ts` |
| Mållistan (/goals) (leader + a/n) | `app/pages/goals/index.vue` |
| Gör nu-listan (/now) (leader + a/n) | `app/pages/now.vue` + `app/composables/usePriorityMode.ts` |
| Taggar på markerat mål (leader + t) | `app/composables/useGoalTagPicker.ts` + `app/layouts/default.vue` |
| Toggle "påbörjat" (leader + s) | `app/pages/goal/[id].vue`, `app/pages/goals/index.vue`, `app/pages/now.vue` |
| Sidorbyte till prioritetssidan (`p`) | `app/composables/usePriorityToggle.ts` |
| Kalendern (t, r, j/k, h/l, +/−) | `app/components/VerticalCalendar.vue` |
| Bekräftelsedialog (h/l/Enter/Esc) | `app/components/ConfirmDeleteGoalModal.vue` |

Att verifiera ändringen:

```bash
cd src/nuxt && bun run build
```

Öppna modalen med `?` och jämför mot listan ovan. Hittar du en tangent i
koden som saknas i modalen - eller tvärtom - är de ur synk.

Modalen är **data-driven**: nya tangenter läggs in i `sections` (script-blocket),
inte i markupen. Sökfiltret i modalen filtrerar automatiskt utifrån samma
data - inget extra att göra där.

### Run locally

```bash
# Nuxt dev server
cd src/nuxt && bun run dev

# Go API
cd src && go run ./api/cmd/api

# PostgreSQL (if local)
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=secret postgres:16
```

### Databas i dev = Neon

Dev kör **inte** mot en Postgres i klustret. `environments/dev/db/neon.env`
(gitignoras) innehåller anslutningssträngarna till Neon-branch `dev` i
projektet `fragrant-pine-27757562`, och blir Secret `plan-db-neon` via
kustomize. Se `environments/dev/README.md` och `environments/dev/db/README.md`.

### Database Schema

Schema is managed declaratively via Atlas Operator:

- `environments/base/db/schema.sql` - SQL definition
- `environments/base/db/atlas-schema.yaml` - Atlas CRD

Changes to schema start in `environments/base/db/schema.sql`, then are tested in dev/pr, and finally applied to production via Flux.

### Testdata i dev-databasen (viktigt)

Dev-databasen i Neon innehåller **användarens riktiga data**. Följande regler
gäller:

- **Rör aldrig rader som inte är testdata.** Radera aldrig "alla rader i ett
  tidsintervall" - det tar användarens data med.
- Testhändelser **ska** ha titeln prefuxad med `[test]`.
- Städa bort testdata med `scripts/cleanup-test-events.sh` (kräver `--force`).
  Verktyget raderar bara `[test]`-rader och listar det som lämnas orört.
- Kolla alltid efter användardata innan en `DELETE` körs mot dev.

### Återställning av raderad data i Neon

Neon-branchen `dev` har 6 timmars WAL-historik
(`history_retention_seconds`). Rader som försvinner utanför det fönstret går
inte att återställa. Kontrollera därför alltid historiken **innan** data
raderas:

```bash
neon connection-string "dev@2026-11-17T09:00:00Z" --project-id fragrant-pine-27757562
psql "<uri>" -tAc "SELECT count(*) FROM calendar_events"
```

Det finns inga snapshots och ingen lokal backup - tidsresan är enda
återställningsvägen.

<!-- wiki:start -->
## Ägande av kunskaps-wikin

Det här repot **äger den globala kunskaps-wikin** (`~/wiki` → `wiki/` här)
och verktyget `wikictl` (`src/rust/crates/plan-wiki`). Alla andra repor
använder wikin men äger den inte.

**Ägareansvar:**

- Ändra inte `wiki/AGENTS.md`-schemat utan att följa de tre lagren:
  råkällor (`raw/`), syntes (`wiki/`) och schema (`AGENTS.md`).
- Råkällor är immutabla - skriv ny syntes i stället för att redigera dem.
- Kör `wikictl lint` efter att du ändrat strukturer eller länkar.

Wikin flyttades hit från `agents`-repot 2026-11-24 i samband med att
agentens backend lades in under `src/agent`. Verktyget byggs med:

```bash
cd src/rust && cargo build -p plan-wiki --bin wikictl
cp src/rust/target/debug/wikictl ~/.local/bin/wikictl
```

Go-versionen finns kvar under `src/agent/cmd/wikictl` tills migreringen
är klar. Båda ger samma svar - `scripts/compare-wiki-parity.sh`
jämför dem kommando för kommando.

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
