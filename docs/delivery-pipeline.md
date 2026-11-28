# Delivery-pipeline: från dev till produktion

Den här dokumentationen beskriver hela vägen från en ändring i repot till
trafik i produktion: lokal utveckling, CI-grindar, imagebygge, publicering,
manifestuppdatering, GitOps-rekonsilering och verifiering.

Allt i den här beskrivningen är **hur det är idag**. Avsnittet
[Kända luckor](#kända-luckor) listar de ställen där kedjan inte håller ihop.

---

## Innehåll

1. [Översikt](#översikt)
2. [Delar och var de ligger](#delar-och-var-de-ligger)
3. [Utvecklingsmiljön](#utvecklingsmiljön)
4. [CI: grindarna på push](#ci-grindarna-på-push)
5. [Imagebygge och publicering](#imagebygge-och-publicering)
6. [Bot-PR: från image till manifest](#bot-pr-från-image-till-manifest)
7. [GitOps: Flux till produktion](#gitops-flux-till-produktion)
8. [Infrastrukturen i klustret](#infrastrukturen-i-klustret)
9. [Databas och schema](#databas-och-schema)
10. [Hemligheter](#hemligheter)
11. [Ingress, TLS och DNS](#ingress-tls-och-dns)
12. [Agentkörningar i drift](#agentkörningar-i-drift)
13. [Rust-migreringen: Rust är huvudspåret](#rust-migreringen-rust-är-huvudspåret)
14. [PR-preview](#pr-preview)
15. [Verifiering](#verifiering)
16. [Rollback](#rollback)
17. [Kända luckor](#kända-luckor)
18. [Referenser](#referenser)

---

## Översikt

```text
┌─ Lokal utveckling ──────────────────────────────────────────────┐
│  kind-kluster ("dev")  +  Tilt  →  bygger images lokalt         │
│  databas: Neon (branch "dev")                                  │
│  portar: :3100 (Nuxt) :4100 (API) :4200 (pgAdmin)              │
└─────────────────────────────────────────────────────────────────┘
                            │ git push main
                            ▼
┌─ GitHub Actions ───────────────────────────────────────────────┐
│  ci.yaml            test + typecheck-gate (Nuxt)                │
│  commit.api.yaml    go vet/test + lint  (Go, ingen image)       │
│  commit.nuxt.yaml  bun-test          → bygg Nuxt-image          │
│  commit.agent.yaml go vet/test       (Go, ingen image)          │
│  commit.rust.yaml  fmt/clippy/test   (Rust, ingen image)       │
│  commit.rust-images.yaml  → bygg rust-api + rust-agent          │
│                           (amd64 + arm64) → bot-PR              │
│  pr-preview.yaml    → bygg preview-images → previews-branchen   │
│                                                                  │
│  Varje byggworkflow → GHCR → bot-PR som byter image-raden i    │
│  environments/prod/patches/*.yaml                               │
└─────────────────────────────────────────────────────────────────┘
                            │ merge till main
                            ▼
┌─ Flux (i klustret) ─────────────────────────────────────────────┐
│  GitRepository "plan" (main, varje 1 min)                      │
│  Kustomization "plan"  path=environments/prod (varje 5 min)    │
│  → kustomize build → apply → SOPS-dekryptering → klustret       │
└─────────────────────────────────────────────────────────────────┘
                            ▼
┌─ Produktion (cluster1, namespace "plan") ───────────────────────┐
│  nuxt-app   → plan-service   → HTTPRoute plan-route            │
│  go-api     → go-api-service → HTTPRoute go-api-route (/api/v1) │
│              (namnet är legacy - imagen är rust-api)            │
│  agent-api  → agent-api      (intern)                          │
│  plan-db        (CloudNativePG, 3 instanser, Longhorn)          │
│  plan-agent-db  (CloudNativePG, 1 instans)                     │
│  plan.simonbrundin.com  (via Cloudflare → cloudflared → Envoy) │
└─────────────────────────────────────────────────────────────────┘
```

**Kärnprincip:** ingenting deployas direkt från en pipeline. Pipelines bygger
images och **committar** en ändring till repo. Flux är den enda som får röra
klustret. Det gör att produktionsläget alltid går att läsa ut av git.

---

## Delar och var de ligger

| Del | Lager | Repo/katalog | Ansvar |
| --- | --- | --- | --- |
| Gränssnitt | Nuxt 4 | `src/nuxt` | Byggs till `ghcr.io/simonbrundin/plan` |
| Plan API | Rust (Go under migrering) | `src/rust/crates/plan-api` | Byggs till `rust-api` |
| Agent | Rust (Go under migrering) | `src/rust/crates/plan-agent` | Byggs till `rust-agent` |
| Manifest (bas) | Kustomize | `environments/base` | Gemensam definition av alla resurser |
| Manifest (dev) | Kustomize | `environments/dev` | Tilt-overlay mot Neon |
| Manifest (PR) | Kustomize | `environments/pr-template/` | Mall för PR-overlays (genereras av `pr-preview-resources.py`) |
| Manifest (prod) | Kustomize | `environments/prod` | Kustomize-patcher + krypterade secrets |
| Pipelines | GitHub Actions | `.github/workflows/` | Test, bygg, publicera, bot-PR |
| Återanvända pipelines | GitHub Actions | `simonbrundin/deployment-pipeline` | `run-tests-bun`, `commit.build-publish-image` |
| Flux-konfiguration | Flux | `~/repos/infrastructure/infrastructure-flux/apps/plan` | Source + Kustomization mot plan-repot |
| Clusterkomponenter | Flux | `~/repos/infrastructure/infrastructure-flux/components/` | CNPG, Atlas, cert-manager, Envoy, Longhorn, Zitadel … |
| Dev-kluster | kind | `~/repos/devenv` | `dev.sh` + `Tiltfile` |

---

## Utvecklingsmiljön

Detaljer finns i [`development.md`](development.md) och
`environments/dev/README.md`. Kortfattat:

```bash
# 1. Klustret (kind "dev") + lokal registry
~/repos/devenv/dev.sh up      # eller motsvarande i devenv

# 2. Dev-miljön
cd environments/dev
tilt up
```

Tilt bygger tre images lokalt och applicerar `environments/dev`-overlayen:

| Image | Byggcontext | Byggs från |
| --- | --- | --- |
| `plan-nuxt` | `src/nuxt` | `src/nuxt/Dockerfile` |
| `plan-api` | `src` | `src/rust/Dockerfile` |
| `plan-agent` | repo-roten | `src/rust/Dockerfile.agent` |

Dev bygger de **samma Rust-images** som produktion kör - samma
Dockerfiles, samma binary. Skillnaden mot prod är databasen (Neon mot
CNPG) och inloggningen (dev-login mot Zitadel), inte koden.

Skillnader mot produktion som är viktiga att känna till:

| | Dev | Prod |
| --- | --- | --- |
| Databas | Neon (branch `dev`) | CloudNativePG `plan-db` (3 instanser) |
| Schema | Atlas Operator mot Neon | Atlas Operator mot CNPG |
| Hemligheter | `environments/dev/secrets/*.yaml` (okrypterade) | `environments/prod/secrets/*.enc.yaml` (SOPS) |
| API-tjänst | Rust (`plan-api`) | Rust (`rust-api`) |
| Agent | Rust (`plan-agent`) | Rust (`rust-agent`) |
| Inloggning | dev-login eller Zitadel | Zitadel |
| Ingress | port-forward via Tilt | Gateway API + Envoy + Cloudflare |

> Dev och produktion kör **samma Rust-images**, byggda ur samma
> Dockerfiles. Go-koden finns kvar och testas i CI, så att pariteten mot
> Rust går att verifiera - men den deployas inte. Se
> [Rust-migreringen](#rust-migreringen-rust-är-huvudspåret).

---

## CI: grindarna på push

Alla workflows triggas av `push` och filtrerar på sökvägar, så att en
ändring i `src/nuxt` inte bygger Rust-images och vice versa.

### `ci.yaml` (alla pushes)

| Jobb | Vad det gör |
| --- | --- |
| `unit-tests` | Kör `bun run test` i `src/nuxt` via den återanvända `run-tests-bun.yaml` |
| `nuxt-unit-tests` | Kör `bun run test:unit` (`bun:test`-sviter) |
| `nuxt-typecheck` | `bun run typecheck` – **typen är en gate, inte en rättelse**: antalet typfel per fil får inte öka |

### `commit.api.yaml` (`src/api/**`, `src/go.mod`, `src/contracts/**`)

`go vet ./api/...` → `go test ./api/...` → golangci-lint v2.

**Bygger ingen image och rör inga manifest.** Go-API:t är den gamla
koden under migreringen - den testas fortfarande, eftersom paritetstester
jämför den mot Rust, men den publiceras inte längre.

### `commit.nuxt.yaml` (`src/nuxt/**`)

Delade `run-tests-bun.yaml` → delad `commit.build-publish-image.yaml`
(bygg + push `ghcr.io/simonbrundin/plan`, amd64 + arm64) → bot-PR.

### `commit.agent.yaml` (`src/agent/**`)

`go vet ./agent/...` → `go test ./agent/...`.

**Bygger ingen image och rör inga manifest** - samma som för Go-API:t.
Agenten i drift är Rust-imagen (`rust-agent`).

### `commit.rust.yaml` (`src/rust/**`)

`cargo fmt --check` → `cargo clippy --all-targets -- -D warnings` →
`cargo test`. **Bygger ingen image** - det gör `commit.rust-images.yaml`.

Paritetstesterna i det här steget jämför Rust-tabellerna mot
route-registreringarna i Go-källan och kontraktet
`src/contracts/plan-agent.yaml` från båda sidor.

### `commit.rust-images.yaml` (`src/rust/**`)

Bygger och publicerar `ghcr.io/simonbrundin/rust-api` och
`ghcr.io/simonbrundin/rust-agent` → bot-PR.

**Båda arkitekturerna byggs.** Jobbet kör i en matris: `linux/amd64` på
`ubuntu-latest` och `linux/arm64` på `ubuntu-24.04-arm`. Arm64 byggs på
en riktig arm64-runner, inte under QEMU - en Rust-kompilering under
emulering tar timmar. Ett efterföljande jobb sätter ihop manifestlistan,
så en och samma tagg pekar på båda arkitekturerna.

Det är **enda workflowen som skriver in images i prod-manifesten**. Den
är därmed den som bestämmer vilken Rust-image som körs i prod - och
ingen Go-workflow kan längre avinstallera den.

### `agent-go.yml`

Kör `go vet`, `go test -race` med coverage och golangci-lint för
`src/agent/**`. Triggerar på push till `main` **och** på PR.

### `pr-preview.yaml`

Se [PR-preview](#pr-preview).

---

## Imagebygge och publicering

Alla images hamnar i **GHCR** (`ghcr.io/simonbrundin/...`) och publiceras med
GitHub Actions `GITHUB_TOKEN` (eller `REGISTRY_TOKEN` för den delade
byggworkflowen).

| Image | Byggcontext | Dockerfile | Plattformar | Referens i manifest |
| --- | --- | --- | --- | --- |
| Nuxt | `src/nuxt` | `src/nuxt/Dockerfile` | amd64, arm64 | `environments/prod/patches/deployment.yaml` |
| Rust API | `src` | `src/rust/Dockerfile` | amd64, arm64 | `environments/prod/patches/api-deployment.yaml` |
| Rust agent | repo-roten | `src/rust/Dockerfile.agent` | amd64, arm64 | `environments/prod/patches/agent-deployment.yaml` |
| Go API | `src` | `src/api/Dockerfile` | - | byggs inte längre |
| Go agent | repo-roten | `src/agent/Dockerfile` | - | byggs inte längre |

**Klustret kör arm64.** En image som bara finns för amd64 ger
`no match for platform in manifest` vid pull - därför byggs allt i
matris, per arkitektur, med egna runners.

Taggar:

- Nuxt: `:${github.sha}` (full sha) - manifestet refererar **digesten**:
  `ghcr.io/simonbrundin/plan@sha256:…`
- Rust: `:sha-<kort sha>` plus `:latest`
- Preview: `plan/pr-preview:<pr-nr>-<sha>` och
  `rust-api/pr-preview:<pr-nr>-<sha>`

Pinning:

- Base-images är pinnade (`node:24-alpine`, `golang:1.26-alpine`,
  `alpine:3.19`, `rust:1.99-alpine`, `gcr.io/distroless/static-debian12:nonroot`).
  Rörliga taggar som `node:lts-alpine` gjorde bygget icke-deterministiskt och
  tog ut CI-bygget i ett dygn - se wikinotet
  *Plan-repots CI- och prod-pipeline var trasig*.
- `bun install --frozen-lockfile --omit=peer` - peer-beroenden (t.ex.
  `better-sqlite3`) auto-installeras av bun och kräver native kompilering som
  är skör över Node-versioner. Appen använder dem inte.

---

## Bot-PR: från image till manifest

Varje byggworkflow gör samma sak till sist: den skriver image-referensen in i
det manifest som styr den tjänsten, och öppnar en PR med boten.

| Workflow | Manifest som ändras | Branch | PR-etikett |
| --- | --- | --- | --- |
| `commit.nuxt.yaml` | `environments/prod/patches/deployment.yaml` | `bot/update-plan-app-image` | `ci`, `skip-review` |
| `commit.rust-images.yaml` | `environments/prod/patches/api-deployment.yaml` + `.../agent-deployment.yaml` | `bot/update-rust-images` | `ci`, `skip-review` |

`commit.rust-images.yaml` är den **enda** workflow som skriver in images i
prod-manifesten. Go-workflows (`commit.api.yaml`, `commit.agent.yaml`) bygger
inte längre några images och öppnar inga bot-PR:er - de testar bara Go-koden
så att pariteten mot Rust går att verifiera.

Base-imagen (`environments/base/agent-deployment.yaml`) pekar fortfarande på
Go-agentens image, och prod patchar över den med Rust-imagen. Base-imagen
ändras först när Go-modulen tas bort helt.

Attributet `skip-review` är viktigt: `pr-preview.yaml` hoppar över PR:er med
den etiketten, annars skulle städjobbet skapa en ny PR för varje PR - en
oändlig kedja.

### Vad som händer när bot-PR:en mergas

1. Mergen hamnar i `main`.
2. Flux `GitRepository "plan"` hämtar den nya revisionen (varje minut).
3. Flux `Kustomization "plan"` bygger `environments/prod` och applicerar
   skillnaden (var femte minut).
4. Kubernetes startar om de pods som fått ny image.

Ingenting annat behöver göras - ingen `kubectl apply`, ingen manuell rollout.

---

## GitOps: Flux till produktion

Flux-konfigurationen för Plan ligger **inte** i plan-repot utan i
infrastrukturrepot:

`~/repos/infrastructure/infrastructure-flux/apps/plan/`

```yaml
# gitrepository.yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: plan
  namespace: flux-system
spec:
  interval: 1m
  url: https://github.com/simonbrundin/plan
  ref:
    branch: main
  secretRef:
    name: github-simonbrundin      # PAT, krypterad med SOPS
```

```yaml
# flux-kustomization.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: plan
  namespace: flux-system
spec:
  interval: 5m
  path: environments/prod
  prune: true
  sourceRef:
    kind: GitRepository
    name: plan
  targetNamespace: plan
  decryption:
    provider: sops
    secretRef:
      name: sops-age
  dependsOn:
    - name: bootstrap
      namespace: flux-system
```

Viktiga egenskaper:

- **HTTPS + PAT** i stället för SSH: repot är stort och arkivkopieringen via
  SSH avbröts med "unexpected EOF" - source-controller hann aldrig hämta en
  ny revision.
- **`prune: true`** - resurser som tas bort ur manifesten tas bort i klustret.
- **`decryption.provider: sops`** - kustomize-controller dekrypterar
  `.enc.yaml` med age-nyckeln i `sops-age` innan apply.
- **`dependsOn: bootstrap`** - CRD:erna (CNPG, Atlas, Gateway API) måste
  finnas innan Plan-overlayn appliceras.

Kontrollera läget:

```bash
kubectl --context admin@cluster1 get kustomization plan -n flux-system
kubectl --context admin@cluster1 get gitrepository plan -n flux-system
kubectl --context admin@cluster1 get events -n flux-system --sort-by=.lastTimestamp | tail
```

Villkor:

```bash
kubectl --context admin@cluster1 wait kustomization/plan -n flux-system \
  --for=condition=Ready --timeout=5m
```

---

## Infrastrukturen i klustret

Klustret (Talos, `cluster1`) byggs av `~/repos/infrastructure` och
`infrastructure-flux`. Bootstrap-steget installerar de CRD:er och operatörer
som Plan behöver:

| Komponent | Vad den gör för Plan |
| --- | --- |
| Flux | Reconciliation av allt ovanstående |
| CloudNativePG | `plan-db` och `plan-agent-db` |
| Atlas Operator | Applicerar `schema.sql` deklarativt |
| Envoy Gateway | `simons-gateway` i namespace `envoy` - ingången till Plan |
| cert-manager | TLS-certifikat för `*.simonbrundin.com` (DNS-01 via Cloudflare) |
| Longhorn | StorageClass `cnpg` för databaserna |
| Zitadel | Identitetsprovider för inloggning |
| MetalLB | VIP till gatewayn (`10.10.10.101`) |
| cloudflared | Tunnel från Cloudflare in till klustret |
| Kyverno, kube-prometheus-stack | Policy och metrics |

Ingressvägen:

```text
Browser
  → Cloudflare (proxy + DNS)
    → cloudflared-tunnel
      → Envoy Gateway "simons-gateway" (ns envoy, VIP 10.10.10.101)
        → HTTPRoute "plan-route"  (host plan.simonbrundin.com) → plan-service → nuxt-app
        → HTTPRoute "go-api-route" (/api/v1, /docs)                 → go-api-service → go-api
```

---

## Databas och schema

### Produktion

`plan-db` är ett CloudNativePG-kluster med **3 instanser**
(`environments/prod/patches/db-instances-patch.yaml`), storageClass `cnpg`
(Longhorn), 1 GiB per instans.

Bootstrapen (`environments/base/db/postgres-cluster.yaml`):

- databas `plan`, ägarroll `plan`
- `postInitSQL` ger `plan` rättigheter på `public`-schemat - **inte** i
  `schema.sql`, eftersom Atlas (community) avvisar `GRANT` med
  "permissions are available to logged-in users only"
- rollen `admin` (superuser) och `plan` (applikationsroll) med lösenord i
  Secrets

### Schemat

Schemat är **deklarativt** och ligger i
`environments/base/db/schema.yaml` (ConfigMap `plan-db-schema`).
`AtlasSchema` (`plan-db-schema`) pekar på ConfigMapen och på
`plan-db-user`-hemligheten.

```bash
kubectl --context admin@cluster1 get atlasschema -n plan
kubectl --context admin@cluster1 logs -n plan -l app.kubernetes.io/name=atlas-schema --tail=50
```

Viktigt: Atlas applicerar **bara DDL**. `INSERT`-raderna i `schema.yaml`
(statusar, Root-målet) körs inte - applikationen seedar själv vid start
(`src/api/internal/database/seed.go`).

**Varning:** Atlas-apply är deklarativ. Ett schema som saknar en tabell
*raderar* den. Töm aldrig `schema.sql` för att testa.

### Dev

Dev har **ingen CNPG-kluster** - `environments/dev/kustomization.yaml`
raderar `plan-db` med `$patch: delete` och pekar om Atlas mot Neon via
`plan-db-neon`. Kopplingarna finns i `environments/dev/db/neon.env`
(gitignoras) och synkas med `./environments/dev/db/sync-neon-env.sh`.

---

## Hemligheter

Det finns **ingen** Vault och ingen ESO i klustret. SOPS + age är den enda
secret-vägen.

| Hemlighet | Fil | Kommentar |
| --- | --- | --- |
| Zitadel + session | `environments/prod/secrets/auth.enc.yaml` | `NUXT_SESSION_PASSWORD`, OAuth-client, redirect-URI |
| DB-roller | `plan-db-user.enc.yaml`, `plan-db-admin.enc.yaml` | Användarnamn/lösenord/URI |
| GHCR-pull | `ghcr-secret.enc.yaml` | `imagePullSecrets` för Nuxt-imagen |
| Agent-nycklar | `agent-api-key.enc.yaml`, `agent-llm.enc.yaml`, `agent-git.enc.yaml` | Skapas av `create-agent-secrets.sh` |
| Agentens DB | `agent-db.enc.yaml` | `plan-agent-db-app` |

Flödet:

1. Skriptet `environments/prod/secrets/create-agent-secrets.sh` tar värdena från
   **miljövariabler** (aldrig argument - de hamnar annars i shell-historiken).
2. `sops --encrypt --age <mottagare>` krypterar varje fil.
3. Mottagaren är repots publika nyckel:
   `environments/base/sops-public-key.yaml` (samma som
   `infrastructure-flux/components/secrets/sops.age.pub`).
4. Den privata nyckeln ligger lokalt i `~/.config/sops-age/keys.txt` och
  Versionshanteras aldrig.
5. Flux dekrypterar filerna med `sops-age` (hemligheten `sops-age` i
   `flux-system`).

Kryptera om en hemlighet:

```bash
export SOPS_AGE_KEY_FILE=~/.config/sops-age/keys.txt
sops --decrypt environments/prod/secrets/auth.enc.yaml
```

---

## Ingress, TLS och DNS

- HTTPRouten `plan-route` ligger i `environments/base/http-route.yaml` och
  pekar på Gatewayn `simons-gateway` i namespace `envoy`.
- `go-api-route` skickar `/api/v1` och `/docs` till API-tjänsten. Namnet
  är legacy - tjänsten kör `rust-api`-imagen sedan migreringen.
- TLS: cert-manager utfärdar `simonbrundin` (wildcard) med
  `cloudflare-clusterissuer` (DNS-01 via Cloudflare-API-token).
- Cloudflare proxar trafiken in via cloudflared-tunneln.

---

## Agentkörningar i drift

Agentens körningar är **Kubernetes Jobs**, inte podar som agent-API:t startar
utan kontroll:

```text
Mål (agent=true)
  → AgentRun (status queued) i Plan-databasen
    → jobbcontroller i agent-API:t skapar ett Job per mål
      → Job: init-container klonar repot, worker kör agenten
        → rapporterar status, kostnad och resultat tillbaka till Plan
          → human gate (godkännande) innan kod ändras
```

Säkerhetsgränser (se `environments/base/agent-job.yaml`):

- `plan-agent-worker` - jobbens identitet, `automountServiceAccountToken: false`
  (jobbet kan inte prata med Kubernetes-API:t)
- `plan-agent-api` - agent-API:ts identitet, får bara `batch/jobs`
  (create/get/list/watch/delete) + läsa poddar och loggar
- resursgränser, `ActiveDeadlineSeconds`, TTL, `RestartPolicy: Never`
- global kapacitet `AGENT_JOB_MAX_CONCURRENT` (standard 4)

Granskning av loggarna:

```bash
kubectl --context admin@cluster1 get jobs -n plan
kubectl --context admin@cluster1 logs -n plan job/<job-namn>
```

---

## Rust-migreringen: Rust är huvudspåret

`specs/rust-migration.md` beskriver migreringen från Go till Rust.
**Rust är den spår som körs i dev och i produktion.** Go-koden finns kvar
som referens och testas i CI (därför att pariteten ska gå att verifiera),
men den byggs inte till images och den deployas inte.

| Spår | Status i CI | Deployar till prod | Används i prod just nu |
| --- | --- | --- | --- |
| Rust (`rust-api`, `rust-agent`) | bygg + publicerar | `api-deployment.yaml`, `agent-deployment.yaml` | Ja |
| Go (`go-api`, `plan-agent`) | testas bara | Nej | Nej |

Eftersom bara `commit.rust-images.yaml` skriver till
`environments/prod/patches/api-deployment.yaml` finns ingen konkurrens
om manifestraden längre - tidigare kunde en Go-build avinstallera
Rust-imagen i prod.

Pariteten mellan Go och Rust hålls av:

- `src/rust/crates/plan-parity` - jämför Rust-tabellerna mot
  route-registreringarna i Go-källan
- `src/contracts/plan-agent.yaml` - kontraktet mellan Plan och agenten,
  testat från båda sidor
- `scripts/compare-api-parity.sh`, `scripts/compare-agent-parity.sh`,
  `scripts/compare-worker-parity.sh`

Paritetstesterna körs i `commit.rust.yaml` (och `commit.api.yaml` /
`commit.agent.yaml` för Go-sidan). De är grinden - en endpoint som glöms
bort i migreringen failar i CI, inte i prod.

---

## PR-preview

Varje PR får en egen, kastbar miljö: eget namespace, egen databas, egen
URL. Flödet går genom GitOps - Actions bygger och committar, **Flux
applicerar**.

```
PR öppnas
   │
   ├── pr-preview.yaml: bygger Nuxt + Rust-API + Rust-agent
   │     (amd64 + arm64) → GHCR
   │
   ├── pr-preview-resources.py: skapar environments/pr/<nr>/
   │     (overlay: namespace, services, routes, secrets, CNPG, Atlas,
   │      agent-databas, agent-rättigheter)
   │
   └── commit → branchen "previews"
                  │
                  ▼
        Flux: GitRepository "plan-previews" (branch: previews)
               Kustomization "plan-preview" (path: environments/pr,
               prune: true)
                  │
                  ▼
        namespace plan-pr-<nr>, URL https://plan-pr-<nr>.simonbrundin.com
```

Detaljer:

- **Branchen `previews`** innehåller bara overlays - ingen kod. GitRepository
  `plan` läser bara `main`, och en PR-branch följs inte av en källkontroll.
- **Overlayen är självständig.** Den refererar inte till `environments/base`,
  eftersom basen inte finns på `previews`-branchen. Schemat kopieras in av
  skriptet från `environments/base/db/schema.yaml` - en enda källa.
- **Båda arkitekturerna byggs**, i en matris per arkitektur, eftersom
  klustret kör arm64.
- **Dev-login** aktiveras i previewen (`DEV_AUTH_ENABLED=true`), så en PR
  kan testas utan Zitadel. API:t och databasen är separata från prod.
- **Schemat appliceras av ett Job**, inte bara av Atlas. Atlas Operator
  räknar med att databasen redan svarar när den börjar - i en ny preview
  kör CNPG initdb, och Atlas kommer före. Operatören ger upp i stället för
  att försöka igen, och då står API:t och väntar på tabeller som aldrig
  skapas. Jobbet `plan-db-schema` väntar på databasen och kör samma
  `schema.sql` som dev och prod. Det är idempotent, så Atlas kan köra samma
  sak efteråt.
- **Städning:** när PR:en stängs tas overlayen bort från `previews`, och
  Flux prunar namespace, databas och routes. `environments/pr/README.md`
  håller katalogen ifylld - en tom katalog ger "no resources" i kustomize
  och ingenting prunas.

### Agenten i previewen

Previewen kör hela kedjan: API, Nuxt **och agent**.

| Del | Resurs i preview-ns | Vad det är |
|-----|--------------------|------------|
| Agent-API | `agent-api` (Deployment + Service `plan-pr-agent-service`) | Samma Rust-image som prod, byggd för PR:en |
| Agentens databas | `plan-agent-db` (CNPG) | Egen databas `agent`, skild från Plan-databasen |
| Rättigheter | `plan-agent-api`, `plan-agent-worker`, Role + RoleBinding | Jobb får skapas i namespace - inget annat |
| Nycklar | `plan-agent-api-key` | `PLAN_API_KEY` (agenten → API:t) och `AGENT_API_KEY` (Nuxt → agenten) |
| Modellnyckel | `plan-agent-llm` (kopieras in krypterat från prod) | `MINIMAX_API_KEY` - utan den kan agenten inte prata med modellen |

Sambandet mellan API:t och agenten:

- API:t får `AGENT_API_KEY` från samma secret som agenten får `PLAN_API_KEY`
  från. Nyckeln ger **agent-scope**: agenten får arbeta med mål och
  körningar, men aldrig besluta om ett godkännande.
- Nuxt får `NUXT_AGENT_API_INTERNAL_URL` (till `plan-pr-agent-service` i
  samma namespace) och `NUXT_AGENT_API_KEY`. Nyckeln skickas server-side,
  aldrig till webbläsaren.

Två avsiktliga begränsningar:

- **Inga git-hemligheter.** `AGENT_JOB_GIT_SECRET` sätts inte, så ett
  agent-jobb kan läsa koden (repot är publikt) men inte committa. Kodändringar
  ska ske i PR:en, inte i repot - en agent som committar till `main` från en
  preview vore en skrivning utanför PR:en.
- **Jobbens repo är PR:ens egen branch** (`AGENT_JOB_REPO_URL` /
  `AGENT_JOB_REPO_BRANCH` sätts från PR-head), så agenten ser samma kod som
  granskaren. Fallback är huvudrepoet.

Agentens databastabeller skapas av agenten själv vid start
(`CREATE TABLE IF NOT EXISTS`) - till skillnad från Plan-databasen, som får
sitt schema från `schema.sql`.

Flux-konfigurationen ligger i infrastructure-repot:
`apps/plan/preview-gitrepository.yaml` och `apps/plan/preview-flux-kustomization.yaml`.

Verifiera en preview:

```bash
kubectl --context admin@cluster1 get pods -n plan-pr-<nr>
curl -sS https://plan-pr-<nr>.simonbrundin.com/api/health
```

Om previewen inte dyker upp - kolla i den här ordningen:

```bash
# 1. Finns overlayen på previews-branchen?
git ls-tree --name-only origin/previews environments/pr/

# 2. Har Flux sett branchen?
kubectl --context admin@cluster1 get gitrepository plan-previews -n flux-system \
  -o jsonpath='{.status.artifact.revision}{"\n"}{.status.conditions[-1].message}{"\n"}'

# 3. Vad säger Flux om bygget?
kubectl --context admin@cluster1 get kustomization plan-preview -n flux-system \
  -o jsonpath='{.status.conditions[-1].message}{"\n"}'

# 4. Står poddarna i namespace?
kubectl --context admin@cluster1 get pods -n plan-pr-<nr>
```

**Om GitRepository svarar "context deadline exceeded"** är det source-controller
som fastnat, inte repot. En omstart av podden löser det:

```bash
kubectl --context admin@cluster1 delete pod -n flux-system -l app=source-controller
```

Symtomet är att *alla* HTTPS-repositories i klustret samtidigt slutar
uppdateras, medan SSH-repositories fungerar.

---

## Verifiering

### Efter att en bot-PR mergats

```bash
# 1. Flux har sett ändringen
kubectl --context admin@cluster1 get kustomization plan -n flux-system \
  -o jsonpath='{.status.lastAppliedRevision}{"\n"}'

# 2. Poddarna rullar ut
kubectl --context admin@cluster1 rollout status deploy/nuxt-app -n plan --timeout=5m
# API:t heter fortfarande "go-api" i manifesten, men kör Rust-imagen.
kubectl --context admin@cluster1 rollout status deploy/go-api   -n plan --timeout=5m
kubectl --context admin@cluster1 rollout status deploy/agent-api -n plan --timeout=5m

# 3. Vilken image körs?
kubectl --context admin@cluster1 get deploy -n plan \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[*].image}{"\n"}{end}'

# 4. Svarar applikationen?
curl -sS https://plan.simonbrundin.com/api/health
curl -sS https://plan.simonbrundin.com/api/v1/health
```

### Före merge - lokalt

```bash
# Manifesten byggs alls?
kustomize build environments/prod > /dev/null && echo OK
kustomize build environments/dev > /dev/null && echo OK

# Inga diffar i base som borde ligga i en overlay?
git diff --stat
```

### Schemat

```bash
# Vad Atlas har gjort
kubectl --context admin@cluster1 logs -n plan -l app.kubernetes.io/name=atlas-schema --tail=100

# Stämmer tabellerna?
kubectl --context admin@cluster1 exec -it plan-db-1 -n plan -- \
  psql -U plan -d plan -c '\dt'
```

---

## Rollback

Eftersom allt går via git är rollbacken en revert:

```bash
# 1. Revertera bot-PR:en (eller en tidigare fungerande image-rad)
git revert <merge-commit>
git push origin main

# 2. Flux applicerar revertet inom 5 minuter
kubectl --context admin@cluster1 wait kustomization/plan -n flux-system \
  --for=condition=Ready --timeout=5m

# 3. Verifiera
kubectl --context admin@cluster1 get deploy -n plan \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[*].image}{"\n"}{end}'
```

**Databasen rullas inte tillbaka.** Atlas är deklarativt: en revert av
manifestet applicerar det tidigare schemat, vilket innebär att kolumner och
tabeller som lades till efteråt **raderas**. Gör alltså en backup
innan du rullar tillbaka en ändring som rörde schemat:

```bash
kubectl --context admin@cluster1 exec -n plan plan-db-1 -- \
  pg_dump -U plan -d plan > plan-$(date +%F).sql
```

---

## Kända luckor

Saker som är trasiga, halvbreda eller riskfyllda just nu - samlat på ett
ställe så att de inte glöms.

1. **API:t startar i mock-läge om Postgres inte är klar.** Rust-API:t
   föll tillbaka till mock-läge när databasen inte svarade vid start, och
   stannade där. Nu väntar API:t i stället på databasen
   (`DB_CONNECT_TIMEOUT_SECONDS`, default 120s) och avslutar om den
   inte kommer - podden startar om och försöker igen. Utan fixen fick en
   nytt kluster (CNPG kör initdb, Atlas applicerar schemat efteråt) ett
   API som svarade "Failed to create user" trots att databasen var
   frisk.
2. **Ingen automatisk smoke-test efter rollout.** Verifieringen är manuell
   (se [Verifiering](#verifiering)).
3. **Go-flödena finns kvar utan att göra jobbet.** `commit.api.yaml` och
   `commit.agent.yaml` testar Go-koden men bygger inga images. De behövs
   så länge paritetstesterna jämför mot Go-källan - men de bör tas bort
   när migreringen är klar.
4. **`acceptance.yaml` är en död stub** - den refererar `frontend/`, en
   katalog som inte finns i repot.
5. **`.github/pr.yaml` är en inaktiv workflow** - GitHub läser bara
   `.github/workflows/`.
6. **`environments/kubernetes/overlays/dev`** är en gammal Flux-baserad
   dev-overlay som inte längre används - dev går via Tilt och kind.
7. **Atlas-apply raderar tabeller som saknas i `schema.sql`.** Ett misstag i
   schemafilen är en destruktiv ändring i prod.
8. **Kyvernos `block-nodeport` blockerade radering av Services.** CEL-uttrycket
   `has(object.spec) && object.spec.type != 'NodePort'` blir falskt vid
   DELETE, eftersom den gamla resursen inte fylls i - och policyn blockerade
   då radering av *alla* Services. Flux kunde inte prunna en preview. Fix:
   policyn validerar bara CREATE och UPDATE. Detta påverkade hela klustret,
   inte bara previewen.
9. **En agent i en preview kör mot riktig modell.** `plan-agent-llm`
   kopieras in krypterat från prod, så ett agent-jobb i previewen förbrukar
   samma modellnyckel som prod. Kostnaden är liten, men nyckeln är densamma.

---

## Referenser

| Dokument | Innehåll |
| --- | --- |
| [`docs/development.md`](development.md) | Utveckling i dev, Tilt, portar |
| [`docs/agent-platform.md`](agent-platform.md) | Agentarkitektur, körningsflöde, säkerhetsgränser |
| [`docs/checkpoint-gate-workflow.md`](checkpoint-gate-workflow.md) | Arbetsflöde för stora förändringar |
| [`specs/rust-migration.md`](../specs/rust-migration.md) | Go → Rust |
| `environments/dev/README.md` | Dev-miljön i detalj |
| `environments/prod/secrets/create-agent-secrets.sh` | Skapa agentens hemligheter |
| `src/contracts/plan-agent.yaml` | Kontraktet mellan Plan och agenten |
| `~/repos/infrastructure/infrastructure-flux/apps/plan/` | Flux-konfigurationen |
| `~/repos/deployment-pipeline/.github/workflows/` | Återanvända pipelines |
| Wiki: `notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg` | Historik: varför bas-images pinnas |
| Wiki: `notes/sops` | SOPS-modellen |
