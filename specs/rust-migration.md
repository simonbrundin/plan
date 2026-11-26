# Migrering: Go -> Rust

Målet: skriva om Plan-applikationens Go-tjänster i Rust. Nuxt-gränssnittet
(Nuxt 4, SPA) lämnas orört - det pratar HTTP mot API:t och påverkas inte.

## Omfattning

Två Go-tjänster finns i repot:

| Tjänst | Katalog | Omfattning | Storlek |
| --- | --- | --- | --- |
| Plan API | `src/api/` | REST-API som Nuxt och agenten pratar med | ~16 kLoC (utöver genererad swagger på 7,5 kLoC) |
| Agent | `src/agent/` | Agentkörningar: orchestrator, LLM-providers, k8s-jobb, scheduler, wiki, memory, verktyg | ~40 kLoC |

Båda migreras, men i ordningen **API först, agent senare**. Agenten pratar
med API:t via `src/contracts/plan-agent.yaml`, så API:t måste stå stabil
innan agenten rörs.

## Agenten (fas 4)

Agenten är två binärer: `agent-api` (HTTP-tjänsten) och `agent-worker`
(workern som utför körningarna). Den portas i den ordningen:

4a. **Planklient + kontrakt** - `planclient` mot Plan API, läst ur
    `src/contracts/plan-agent.yaml`. Kontrakstestet på båda sidor
    (Rust: `crates/plan-agent/tests/contract.rs`, Go:
    `src/agent/internal/planclient/contract_test.go`).
4b. **agent-api** - samma routes som `src/agent/internal/api/server.go`:
    status, workflow, samtal och meddelanden (inkl. ström), godkännanden,
    API-nycklar, verktygsavbrott, chatt-WebSocket.
4c. **LLM-providers** - `core/provider` (Anthropic, MiniMax, OpenAI) blir
    råa HTTP-anrop mot respektive API. `anthropic-sdk-go` försvinner.
4d. **Kärnan** - orchestrator (planner, checkpoint, gate), memory,
    secrets, usage, messages, selfimprove.
4e. **Workern** - `agentworker`, `definer`, `proposer`, `coder`: pollar
    Plan efter mål, definierar mål, föreslår kodändringar, kör git.
4f. **Kubernetes** - `jobcontroller`, `jobrunner`, `infrastructure/k8sjob`
    blir `kube-rs`. Jobb skapas, observeras och avslutas.
4g. **Wiki, workspace, scheduler** - resten av agentens utkant.

## Ordning

1. **Fas 0 - grund** (den här fasen): Cargo-workspace, config, DB-pool,
   startmigreringar, /health, router med samma vägar som Go, paritetstest.
2. **Fas 1 - auth**: sessioner (HMAC-cookie + DB), API-nycklar (scope),
   Zitadel-JWT/JWKS, agent-scope-middleware.
3. **Fas 2 - data**: modeller + repositories (goals, calendar_events,
   calendars, places, tags, agent_runs, agent_usage, agent_approvals,
   api_keys). Samma SQL som i Go, samma kolumnnamn och typer.
4. **Fas 3 - handlers**: alla endpoints, samma statuskoder och JSON-format.
   OpenAPI-specen genereras om med `utoipa` (Go använde swaggo).
5. **Fas 4 - agenten**: port av `src/agent/` till `plan-agent`-crate:
   planclient, orchestrator, providers (Anthropic), k8s-jobb (kube-rs),
   scheduler, wiki, memory, verktyg, secrets.
6. **Fas 5 - utfasning**: Go tas bort, CI byter till Rust, image byts i
   Flux, Go-modulen raderas.

## Regler

- **Nuxt rörs inte.** Endast om ett API-svar har en bugg som hindrar
  gränssnittet ändras Nuxt - aldrig för att Rust ska bli enklare.
- **Paritet är kravet.** Samma vägar, samma metoder, samma statuskoder,
  samma JSON-fält, samma felmeddelanden. Ett Rust-svar får inte räcka
  längre än ett Go-svar.
- **Paritetstest.** Rust-router-tabellen jämförs programmatiskt mot
  route-registreringarna i `src/api/cmd/api/main.go`, så en saknad
  endpoint failar i CI.
- **Go ligger kvar tills fas 5.** Båda kan köra parallellt mot samma
  databas under utvecklingen.
- **Inga hemligheter i repot.** Secrets kommer från samma källor som
  nu: SOPS/Flux i klustret, `NUXT_SESSION_PASSWORD` och `SESSION_SECRET`
  som env-variabler.

## Risks

- **RRULE.** Go använder `rrule-go`, gränssnittet använder `rrule` (JS).
  Rust-krigen `rrule` är svagt underhållt. Plan: egen liten RFC 5545-motor
  i Rust med testvektorer som jämförs mot Go-implementationen.
- **K8s i agenten.** `client-go` har ingen fullt motsvarande Rust-krust.
  `kube-rs` täcker det agenten använder (skapa/observera/avsluta Job).
- **Anthropic-SDK.** `anthropic-sdk-go` ersätts av råa HTTP-anrop mot
  Anthropic-API:et (`reqwest`) - mindre kod, inget beroende.
- **OpenAPI.** Swagger-filen i repot är stale (innehållerEndpoints från en
  gammal agent-UI, t.ex. `/workflow`, `/conversations`). Den regenereras i
  fas 3 och kontraktstestet körs mot båda filerna under fas 5.

## Status

| Del | Status |
| --- | --- |
| Cargo-workspace, config, DB, seed | klar |
| Router + route-paritetstest (API) | klar |
| Router + route-paritetstest (agent) | klar |
| Planklient + kontrakttest | klar |
| OpenAPI code-first (utoipa) + Scalar | klar |
| Auth (sessioner, API-nycklar, Zitadel, agent-scope) | klar |
| Modeller + repositories (statusar, mål, taggar, platser) | klar för de domänerna |
| Modeller + repositories (agentkörningar, händelser, godkännanden, budgetar, användning, API-nycklar) | klar |
| Versionsstyrd prislista (kostnad i Plan) | klar |
| Handlers | 76 av 76 portade (statusar, mål, taggar, relationer, beroenden, platser, kalendrar, kalenderhändelser, agentkörningar, godkännanden, budgetar, användning, API-nycklar, auth, kompletteringar) |
| Live-paritetsskript (Go vs Rust, samma databas) | klar |
| Agentens API, providers, kärna, worker | inte påbörjat |

### Vad som återstår i API:t

- Inget - alla Go-vägar är portade. Resterande arbete ligger i agenten
  (API-handlers, providers, kärna, worker) och i att Go-tjänsten tas
  bort.

### Noter om paritet

- Sessionstoken är byte-kompatibla med Go, så en cookie fungerar i båda
  tjänsterna medan de kör parallellt.
- Tidsstämplar renderas i UTC i Rust och i serverns tidszon i Go. Det är
  samma tidpunkt, men olika text - paritetsskriptet normaliserar dem.
- Ett uppdaterat mål svarar `null` för platser och taggar i båda
  tjänsterna, medan en lista svarar `[]`. Skillnaden syns i svaret och
  måste därför behållas.
- Återkommande serier expanderas av `plan-recurrence` - en egen motor för
  den delmängd av RRULE som appen använder (`FREQ`, `INTERVAL`, `COUNT`,
  `UNTIL`, `BYDAY`). Testvektorerna är desamma som i Go:s
  `recurrence_test.go`.
- Ett skapat override eller en delad serie svarar `places: null` i båda
  tjänsterna - Go fyller inte i platser i de svaren. Listor, enskilda
  händelser och skapade händelser svarar `places: []`.
- Gin svarar `404 page not found` i ren text för okända vägar och för
  metoder som inte stöds. Rust har därför en egen fallback utanför
  auth-lagret - annars hade okända vägar svarat 401.
- Ett terminalt tillstånd i livscykeln har `null` som övergångar i båda
  tjänsterna - Go lämnar en tom lista som null.
- Kostnaden räknas i Plan ur en versionsstyrd prislista. En okänd modell
  får dokumenterat fallbackpris med `price_source: "fallback"`, aldrig
  nollkostnad.
- Authfel svarar `{"error": ...}` medan övriga API-fel svarar
  `{"message": ...}` - det är Gons kontrakt, och det gäller även i Rust.
- En återkommande måls regel kan sättas vid skapande och vid uppdatering.
  Serien börjar på `recurrence_start_date` (annars idag), och de
  tidigare slutförandena rensas när regeln byts - de hörde till den
  gamla regeln.
- RRULE-felen skrivs med samma ordval som `rrule-go` ("undefined
  frequency: X", "INTERVAL måste vara minst 1" och så vidare), så att
  klienten får samma meddelande från båda tjänsterna.
- Varje route i tabellen måste ha en handler - annars svarar den 501.
  Route-tabellens test kontrollerar det, så en endpoint inte kan glömmas
  bort utan att testet fallerar.

## Layout

```
src/rust/
  Cargo.toml              # workspace
  crates/plan-api/        # REST-API
  crates/plan-agent/      # agenttjänsten (fas 4)
  crates/plan-parity/     # paritetsverktyg: jämför Rust-routes mot Go-källan
  crates/plan-recurrence/ # RRULE-motor - port av Go-tjänstens rrule-go-användning
```

`plan-parity` läser route-registreringarna ur Go-källan och jämför dem med
Rust-tabellerna. Det är det som fångar en glömd endpoint i CI, innan den
nått produktion.
