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
13. [Rust-migreringen: två parallella spår](#rust-migreringen-två-parallella-spår)
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
│  commit.api.yaml    go vet/test + lint  → bygg Go-API-image     │
│  commit.nuxt.yaml  bun-test          → bygg Nuxt-image          │
│  commit.agent.yaml go vet/test       → bygg Go-agent-image      │
│  commit.rust.yaml  fmt/clippy/test   (Rust, ingen image)       │
│  commit.rust-images.yaml             → bygg rust-api + rust-agent│
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
│  go-api     → go-api-service → HTTPRoute go-api-route (/api/v1)│
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
| Plan API | Go → Rust | `src/api` → `src/rust/crates/plan-api` | Byggs till `go-api` / `rust-api` |
| Agent | Go → Rust | `src/agent` → `src/rust/crates/plan-agent` | Byggs till `plan-agent` / `rust-agent` |
| Manifest (bas) | Kustomize | `environments/base` | Gemensam definition av alla resurser |
| Manifest (dev) | Kustomize | `environments/dev` | Tilt-overlay mot Neon |
| Manifest (PR) | Kustomize | `environments/pr/<nr>` | Bot-skapade overlays |
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
| `plan-go-api` | `src` | `src/api/Dockerfile` |
| `plan-agent` | repo-roten | `src/agent/Dockerfile` |

Skillnader mot produktion som är viktiga att känna till:

| | Dev | Prod |
| --- | --- | --- |
| Databas | Neon (branch `dev`) | CloudNativePG `plan-db` (3 instanser) |
| Schema | Atlas Operator mot Neon | Atlas Operator mot CNPG |
| Hemligheter | `environments/dev/secrets/*.yaml` (okrypterade) | `environments/prod/secrets/*.enc.yaml` (SOPS) |
| API-tjänst | Go (`plan-go-api`) | Rust (`rust-api`) |
| Agent | Go (`plan-agent`) | Rust (`rust-agent`) |
| Inloggning | dev-login eller Zitadel | Zitadel |
| Ingress | port-forward via Tilt | Gateway API + Envoy + Cloudflare |

> Dev kör fortfarande Go-tjänsterna medan produktion kör Rust. Pariteten
> hålls av paritetstesterna i `src/rust/crates/plan-parity` och
> `scripts/compare-*-parity.sh` - se [Kända luckor](#kända-luckor).

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

`go vet ./api/...` → `go test ./api/...` → golangci-lint v2 →
bygg och push `ghcr.io/simonbrundin/go-api` (amd64 + arm64) →
bot-PR.

### `commit.nuxt.yaml` (`src/nuxt/**`)

Delade `run-tests-bun.yaml` → delad `commit.build-publish-image.yaml`
(bygg + push `ghcr.io/simonbrundin/plan`, amd64 + arm64) → bot-PR.

### `commit.agent.yaml` (`src/agent/**`)

`go vet ./agent/...` → `go test ./agent/...` → bygg och push
`ghcr.io/simonbrundin/plan-agent` → bot-PR.

### `commit.rust.yaml` (`src/rust/**`)

`cargo fmt --check` → `cargo clippy --all-targets -- -D warnings` →
`cargo test`. **Bygger ingen image** - det gör `commit.rust-images.yaml`.

Paritetstesterna i det här steget jämför Rust-tabellerna mot
route-registreringarna i Go-källan och kontraktet
`src/contracts/plan-agent.yaml` från båda sidor.

### `commit.rust-images.yaml` (`src/rust/**`)

Bygger och publicerar `ghcr.io/simonbrundin/rust-api` och
`ghcr.io/simonbrundin/rust-agent` (bara amd64 - arm64 under QEMU tar
timmar för en Rust-kompilering) → bot-PR.

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
| Go API | `src` | `src/api/Dockerfile` | amd64, arm64 | `environments/prod/patches/api-deployment.yaml` |
| Go agent | repo-roten | `src/agent/Dockerfile` | amd64, arm64 | `environments/base/agent-deployment.yaml` |
| Rust API | `src` | `src/rust/Dockerfile` | amd64 | `environments/prod/patches/api-deployment.yaml` |
| Rust agent | repo-roten | `src/rust/Dockerfile.agent` | amd64 | `environments/prod/patches/agent-deployment.yaml` |

Taggar:

- Nuxt: `:${github.sha}` (full sha) - manifestet refererar **digesten**:
  `ghcr.io/simonbrundin/plan@sha256:…`
- Go/Rust: `:sha-<kort sha>` plus `:latest`

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
| `commit.api.yaml` | `environments/prod/patches/api-deployment.yaml` | `bot/update-plan-api-image` | `ci`, `skip-review` |
| `commit.agent.yaml` | `environments/base/agent-deployment.yaml` | `bot/update-plan-agent-image` | `ci`, `skip-review` |
| `commit.rust-images.yaml` | `environments/prod/patches/api-deployment.yaml` + `.../agent-deployment.yaml` | `bot/update-rust-images` | `ci`, `skip-review` |

Notera att `commit.agent.yaml` skriver **Go-agentens** image till `base/`, medan
produktion patchar över med Rust-imagen i `environments/prod/patches/agent-deployment.yaml`.
Base-imagen ändras bara när Go-modulen tas bort helt.

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
        → HTTPRoute "go-api-route" (/api/v1, /docs)                → go-api-service → go-api
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
- `go-api-route` skickar `/api/v1` och `/docs` till API-tjänsten.
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

## Rust-migreringen: två parallella spår

`specs/rust-migration.md` beskriver migreringen från Go till Rust. Under
migreringen lever **två** pipelines parallellt:

| Spår | Bygger | Deployar till prod | Används i prod just nu |
| --- | --- | --- | --- |
| Go | `go-api`, `plan-agent` | `api-deployment.yaml`, `base/agent-deployment.yaml` | Nej - patchas över |
| Rust | `rust-api`, `rust-agent` | `api-deployment.yaml`, `agent-deployment.yaml` | Ja |

Eftersom båda spåren skriver till `environments/prod/patches/api-deployment.yaml`
(rad för rad med `sed`) är det **senaste mergen som vinner**. Det är den
viktigaste konflikten i kedjan just nu - se [Kända luckor](#kända-luckor).

Pariteten mellan Go och Rust hålls av:

- `src/rust/crates/plan-parity` - jämför Rust-tabellerna mot
  route-registreringarna i Go-källan
- `src/contracts/plan-agent.yaml` - kontraktet mellan Plan och agenten,
  testat från båda sidor
- `scripts/compare-api-parity.sh`, `scripts/compare-agent-parity.sh`,
  `scripts/compare-worker-parity.sh`

---

## PR-preview

`pr-preview.yaml` ska ge varje PR en egen miljö:

1. Bygger Nuxt-imagen med taggen `pr-<nr>-<sha>` och publicerar den.
2. Kör `.github/scripts/pr-preview-resources.py`, som skapar
   `environments/pr/<nr>/kustomization.yaml` (namespace `plan-pr-<nr>`,
   egen hostname `plan-pr-<nr>.simonbrundin.com`) och lägger in PR:en i
   `environments/pr/kustomization.yaml`.
3. Committar till PR-branchen och kommenterar i PR:en.
4. Vid stängning rensar `.github/scripts/pr-preview-cleanup.py` bort overlayen
   via ett PR med etiketten `skip-review`.

**Verkligheten idag:** det finns ingen Flux `Kustomization` som pekar på
`environments/pr`, och klustret har inga `plan-pr-*`-namespaces. Ingenting
deployas. Se [Kända luckor](#kända-luckor).

Dessutom är `.github/pr.yaml` en **inaktiv** workflow - GitHub läser bara
`.github/workflows/*.yml`. Den pekar dessutom på
`environments/kubernetes/overlays/template-ephemeral`, som inte finns.

---

## Verifiering

### Efter att en bot-PR mergats

```bash
# 1. Flux har sett ändringen
kubectl --context admin@cluster1 get kustomization plan -n flux-system \
  -o jsonpath='{.status.lastAppliedRevision}{"\n"}'

# 2. Poddarna rullar ut
kubectl --context admin@cluster1 rollout status deploy/nuxt-app -n plan --timeout=5m
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

1. **PR-preview deployar ingenting.** `pr-preview.yaml` skapar overlayer i
   `environments/pr/`, men ingen Flux `Kustomization` pekar på den sökvägen.
   Klustret har inga `plan-pr-*`-namespaces. Fix: lägg till en
   `Kustomization` (t.ex. `apps/plan/flux-kustomization-pr.yaml`) med
   `path: environments/pr` och `prune: true`.
2. **Go- och Rust-flöden konkurrerar om samma manifest-rad.**
   `commit.api.yaml` och `commit.rust-images.yaml` skriver båda image-raden i
   `environments/prod/patches/api-deployment.yaml`. Den senaste mergen vinner,
   och en Go-build kan alltså avinstallera Rust-imagen i prod. Fix: låt
   `commit.api.yaml` sluta röra prod-manifesten när Rust är i drift (eller
   peka om den på en egen patch).
3. **Dev kör Go, prod kör Rust.** Paritetsrisken bärs av paritetstesterna, men
   dev-miljön verifierar inte exakt det som körs i prod.
4. **Ingen byggning av images på PR.** Bara Nuxt byggs (för previewen, som
   inte fungerar). Rust- och agent-imagerna testas inte före merge.
5. **`.github/workflows/acceptance.yaml` är en död stub** - den refererar
   `frontend/`, en katalog som inte finns i repot.
6. **`.github/pr.yaml` är en inaktiv workflow** - GitHub läser bara
   `.github/workflows/`.
7. **Ingen automatisk smoke-test efter rollout.** Verifieringen är manuell
   (se [Verifiering](#verifiering)).
8. **`environments/kubernetes/overlays/dev`** är en gammal Flux-baserad
   dev-overlay som inte längre används - dev går via Tilt och kind.
9. **Atlas-apply raderar tabeller som saknas i `schema.sql`.** Ett misstag i
   schemafilen är en destruktiv ändring i prod.

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
