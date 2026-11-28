---
title: Från dev till produktion
description: Hela leveranskedjan - lokal utveckling, CI, imagebygge, bot-PR, Flux och verifiering.
order: 8
icon: i-lucide-rocket
updated: 2026-11-28
group: Utvecklare
---

# Från dev till produktion

Beskrivningen gäller **hur kedjan ser ut idag**. Avsnittet
[Kända luckor](#kända-luckor) listar ställen där den inte håller ihop.

## Kedjan i korthet

```text
┌─ Lokal utveckling ──────────────────────────────────────────────┐
│  kind-kluster ("dev") + Tilt → bygger images lokalt             │
│  databas: Neon (branch "dev")                                  │
│  portar: :3100 (Nuxt) :4100 (API) :4200 (pgAdmin)              │
└─────────────────────────────────────────────────────────────────┘
                            │ git push main
                            ▼
┌─ GitHub Actions ───────────────────────────────────────────────┐
│  ci.yaml             test + typecheck-gate (Nuxt)               │
│  commit.rust.yaml    fmt/clippy/test  (Rust)                    │
│  commit.nuxt.yaml    bun-test → bygg Nuxt-image                 │
│  commit.rust-images  bygg rust-api + rust-agent (amd64+arm64)  │
│  pr-preview.yaml     bygg preview-images → previews-branchen    │
│                                                                  │
│  Varje byggworkflow → GHCR → bot-PR som byter image-raden i    │
│  environments/prod/patches/*.yaml                               │
└─────────────────────────────────────────────────────────────────┘
                            │ merge till main
                            ▼
┌─ Flux (i klustret) ─────────────────────────────────────────────┐
│  GitRepository "plan" (main, varje minut)                      │
│  Kustomization "plan"  path=environments/prod (var 5:e minut)  │
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
images och **committar** en ändring till repot. Flux är den enda som får röra
klustret - därför går det alltid att läsa av produktionsläget i git.

## Lokal utveckling

```bash
cd environments/dev
tilt up
```

Tilt bygger Plan och agenten och deployar dem mot Neon-branch `dev`. Se
[Utveckling och drift](/docs/utveckling) för portar och testkommandon.

## CI: grindarna på push

| Workflow | Triggas av | Vad den gör |
| --- | --- | --- |
| `ci.yaml` | alla pushes | tester + typecheck-gate för Nuxt |
| `commit.rust.yaml` | `src/rust/**` | `cargo fmt`, `clippy`, tester - ingen image |
| `commit.api.yaml` / `commit.agent.yaml` | `src/api/**`, `src/agent/**` | Go-tester för paritet - ingen image |
| `commit.nuxt.yaml` | `src/nuxt/**` | bun-tester, bygger Nuxt-imagen |
| `commit.rust-images.yaml` | `src/rust/**` | bygger `rust-api` + `rust-agent` → bot-PR |
| `pr-preview.yaml` | PR | bygger preview-images → `previews`-branchen |

Rust är huvudspåret: Go-koden testas fortfarande för paritet mot
route-implementationen, men bygger inga images.

## Imagebygge och publicering

Båda arkitekturerna byggs, och varje arkitektur byggs på en runner av
**samma** arkitektur - arm64 under QEMU tar timmar för en
Rust-kompilering.

| Image | Registry | Byggs av |
| --- | --- | --- |
| `ghcr.io/simonbrundin/plan` | GHCR | `commit.nuxt.yaml` |
| `ghcr.io/simonbrundin/rust-api` | GHCR | `commit.rust-images.yaml` |
| `ghcr.io/simonbrundin/rust-agent` | GHCR | `commit.rust-images.yaml` |

Images taggas med commit-shas (`sha-<sha>`) eller med PR-nummer för
previews. Taggen är också **build-identiteten**: API:t och agenten
rapporterar vilken build som kör (`PLAN_BUILD_SHA`), vilket rökprovet
använder för att veta att den nya versionen faktiskt svarar.

## Bot-PR: från image till manifest

Byggworkflowen skapar en PR som bara byter image-raderna i
`environments/prod/patches/`:

```yaml
image: ghcr.io/simonbrundin/rust-api:sha-4b952cf
```

Bot-PR:en har alltid `base: main` - annars hamnade manifeständringen i
feature-branchen. `commit.rust-images.yaml` är den enda workflow som får
skriva Rust-images till prod-manifestet: två pipelines som skriver samma
rad gör att "senaste mergen vinner", och Go-builden kunde avinstallera
Rust-imagen i produktion.

## GitOps: Flux till produktion

Flux läser repot och applicerar `environments/prod`:

| Resurs | Vad den gör |
| --- | --- |
| `GitRepository/plan` | klonar `main` (varje minut) |
| `Kustomization/plan` | `kustomize build environments/prod` → apply (var 5:e minut) |

Manifesten är Kustomize-bas med overlays: `environments/base` innehåller
den gemensamma definitionen, `environments/prod` patchar den. Hemligheter
ligger som SOPS-krypterade filer (`environments/prod/secrets/*.enc.yaml`)
och dekrypteras av Flux vid apply - inga hemligheter i klartext i git.

## Vad som körs i produktion

| Deployment | Image | Nås via |
| --- | --- | --- |
| `nuxt-app` | `ghcr.io/simonbrundin/plan` | `plan.simonbrundin.com` |
| `go-api` | `ghcr.io/simonbrundin/rust-api` | `/api/v1`, `/docs`, `/health` |
| `agent-api` | `ghcr.io/simonbrundin/rust-agent` | internt + via Nuxt |

`go-api` heter så i manifesten än idag, men imagen är `rust-api` - Rust
har ersatt Go i produktion.

## PR-preview

Varje PR får en egen, kastbar miljö:

```text
namespace   plan-pr-<nr>
URL         https://plan-pr-<nr>.simonbrundin.com
databas     egen CloudNativePG-instans (plan-db + plan-agent-db)
inloggning  dev-login (ingen Zitadel)
```

Överlappen genereras av `.github/scripts/pr-preview-resources.py` och
publiceras på branchen `previews`, som Flux läser via `GitRepository
"plan-previews"` (`Kustomization "plan-preview"`, path `environments/pr`,
`prune: true`). När PR:en stängs tas överlappen bort och Flux rensar
namespacet.

## Verifiering

Efter varje rollout körs ett rökprov:

```text
preview   bygg-version → root → /api/health → /health → dev-login
          → /api/v1/goals → agentens endpoint
prod      bygg-version → root → /api/health → /health
          → skyddade endpoints ska ge 401 utan session
```

| Flöde | När | Var |
| --- | --- | --- |
| `smoke-test.yaml` | när en bot-PR med images mergas | kommentarar på PR:et |
| `pr-preview.yaml` | efter att previewen publicerats | kommenterar på PR:et |

Manuellt:

```bash
.github/scripts/smoke-test.sh --url https://plan.simonbrundin.com --mode prod
.github/scripts/smoke-test.sh --url https://plan-pr-<nr>.simonbrundin.com --mode preview
```

## Rollback

Revertera bot-PR:en (eller sätt tillbaka en tidigare fungerande
image-rad). Flux applicerar revertet inom fem minuter.

**Varning:** manifest-revert kan få Atlas att radera senare tillagda
tabeller och kolumner - schemat är destruktivt i båda riktningar. Ta
databasbackup före schemaändringar.

## Kända luckor

- Rökprovet når inte databasen i produktion - där finns ingen dev-login,
  så auth ger 401. I previewläget testas databasen på riktigt.
- Go-workflows finns kvar utan att göra jobbet (bara paritetstester).
  De bör tas bort när migreringen är klar.
- `acceptance.yaml` är en död stub och `.github/pr.yaml` är en inaktiv
  workflow.
- Previewens agent använder samma modellnyckel som produktion.
- Ett rökprov testar inte regressioner - det bekräftar bara att den
  nya versionen svarar.

## Se också

- [Utveckling och drift](/docs/utveckling) - lokalt, tester, bygga.
- [Arkitektur](/docs/arkitektur) - tjänsterna och deras ansvar.
- [API och autentisering](/docs/api) - endpoints, sessioner, scopes.
