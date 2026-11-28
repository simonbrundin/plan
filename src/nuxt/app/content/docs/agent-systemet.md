---
title: Agentsystemet
description: Plan har en egen agent - en motor som kan specificera, planera, utföra och verifiera mål, med human gate, profiler och kostnadskontroll.
order: 4
icon: i-lucide-bot
updated: 2026-11-28
group: Agenten
---

# Agentsystemet

Utöver dina egna mål kan Plan arbeta på mål åt dig. En agentkörning är
ett spår: den får ett mål, går igenom fyra faser och rapporterar
tillbaka vad som hände.

**Plan äger körningen, agenten utför arbetet.** Körningen, händelserna
och godkännandena lagras i Plan - inte i agenten. Därför kan
gränssnittet visa en pågående körning även om agenttjänsten är nere, och
ett godkännande som du redan lämnat finns kvar.

## Faser

```text
define ──▶ plan ──▶ execute ──▶ verify ──▶ completed
```

| Fas | Vad som händer |
| --- | --- |
| `define` | Agenten skriver målets definition - vad som egentligen ska göras, vilka krav som gäller. Förslaget går via human gate innan något skrivs |
| `plan` | Definitionen bryts ner till checkpoints med uppgifter |
| `execute` | Checkpoints utförs, kod skrivs i en isolerad kopia av repot |
| `verify` | Testerna körs och resultatet verifieras |
| `completed` | Körningen är klar |

Faserna speglar målets status i databasen (`statuses`), så du ser samma
förlopp i mållistan som i agentvyn.

## Tillstånd

| Tillstånd | Betydelse |
| --- | --- |
| `queued` | Körningen väntar på ett Kubernetes Job |
| `running` | Agenten arbetar |
| `waiting_approval` | Agenten väntar på ett beslut från dig |
| `succeeded` | Körningen lyckades |
| `failed` | Körningen misslyckades |
| `cancelled` | Du avbröt körningen |

`queued`, `running` och `waiting_approval` är de tillstånd där en
människa kan ingripa. `succeeded`, `failed` och `cancelled` är
terminala - en körning lämnar dem inte.

## Triggers och profiler

| Trigger | Betydelse |
| --- | --- |
| `user` | Du startade den |
| `schedule` | Den schemalagdes |
| `webhook` | Den utlöstes av en händelse utifrån |
| `agent` | Agenten startade den själv |

Profilerna är **tekniska gränser**, inte instruktioner till modellen.
De gäller oavsett vad modellen säger.

| Gräns | `coding` | `life` |
| --- | --- | --- |
| Får ändra kod | Ja | Nej |
| Behöver klona repot | Ja | Nej |
| Kräver mänskligt beslut | Ja | Ja |
| CPU / minne | 1000m / 2Gi | 500m / 1Gi |
| Max antal rundor | 20 | 12 |

En okänd profil ger `coding` - den strängaste. Profilen väljs när
körningen skapas (`agent_profile`).

## Komponenterna

```text
src/agent/
├── internal/
│   ├── agentworker/   Agent-arbetsmotor (internt paketnamn, inte binär)
│   ├── definer/       Skriver definitionen (fas define)
│   ├── proposer/      Föreslår kodändringar
│   ├── coder/         Driver kodändringen från förslag till verifierat resultat
│   ├── executor/      Kör kommandon i workspace
│   ├── workspace/     Hanterar den isolerade kopian av repot
│   ├── jobcontroller/ Skapar Kubernetes Jobb per köad körning
│   ├── jobrunner/     Bygger och startar jobben
│   ├── profile/       Profilernas policies
│   ├── planclient/    Agentens klient mot Plan
│   ├── runusage/      Samlar tokenanvändning
│   ├── scheduler/     Tidsatt start av körningar
│   ├── wiki/          Kunskapswikin
│   └── core/
│       ├── orchestrator/ Checkpoint-exekvering och sub-agenter
│       ├── gate/     Gates: human, ui, behavior, adversarial
│       ├── tools/    Verktyg modellen kan använda
│       ├── provider/ LLM-leverantörer och fallback
│       ├── memory/   Agentens minne
│       ├── selfimprove/ Lärdomar från körningar
│       └── checkpoint/ Checkpoint-modellen
├── cmd/agent-worker/  Workerns ingång
└── cmd/wikictl/       Verktyget för kunskapswikin
```

### Verktygen

Modellen arbetar med verktyg, inte med fri text: `bash`, `read`,
`write`, `edit`, `glob`, `grep`, `read_multiple`. De är validerade
innan de körs - det finns separata validatorer för kärnan och för
säkerhet (hemligheter, kommandon utanför arbetsytan).

### Gates

En **gate** är en grind som en körning måste passera. Det finns fyra
sorters gates:

| Gate | Gör |
| --- | --- |
| `human` | Kräver ett beslut av en människa innan något skrivs |
| `ui` | Syns i gränssnittet och kan interaktas med |
| `behavior` | Kontrollerar att beteendet håller - regressioner, format |
| `adversarial` | En granskande agent letar efter problem i förslaget |

### Human gate

Human gate är kärnan i "människan i loopen". Den körs när agenten har
ett förslag som det vill skriva, och beslutet tas i Plan - i
gränssnittet, på sidan `/agent`.

Ett godkännande ska visa:

- vilket mål det gäller,
- vilka observationer agenten gjort,
- vad beslutet får till följd,
- hur resultatet ska verifieras.

Inte bara det tekniska kommandot. Beslutet sparas i `agent_approvals`
med `feedback`, `consequence` och `verification_plan`.

## Fasmaskinen - Plan bestämmer faserna

Faserna är inte bara en modell i agentens prompt. De är en maskin i
Plan (`src/api/internal/phase`), och alla vägar till en fas går genom
den - även den gamla `PATCH /goals/:id/status`.

```text
define ──▶ plan ──▶ execute ──▶ verify ──▶ completed
```

Människan och agenten har olika rättigheter, inte olika regler:

| Övergång | Människa | Agent |
| --- | --- | --- |
| Framåt | Får, även om den hoppar över faser | Får, men måste uppfylla fasens villkor |
| Bakåt | Får - loggas som avvikelse | Får inte |
| Avsluta | Får, från vilken fas som helst | Får inte - ska godkännas av en människa |

Villkoren för agenten är: en godkänd definition (`define` → `plan`),
planerade delmål (`plan` → `execute`), lösta delmål (`execute` →
`verify`).

En stoppad övergång ger `409` med en stabil skälkod - `phase_skipped`,
`backwards_transition`, `missing_planned_children`, `unresolved_children`,
`definition_not_approved` eller `approval_required`. Koderna är
kontraktet mellan agenten och Plan: agenten kan svara på dem i stället
för att gissa.

## Verifieringsplanen - valfri, men den ger skillnaden

En verifieringsplan anger **hur** ett mål ska räknas som klart: typ
(`checklista`, `delmal`, `kalender`, `bevis`, `matvarde`, `kod`),
villkor och styrka (`stark` / `svag`).

Planen är valfri. Utan plan fungerar målet precis som i dag - bara
räknas det aldrig som "verifierat", bara som "deklarerat".

```text
mål med plan ──▶ verifieringen körs ──▶ omdöme ──▶ människan beslutar
                                                              │
                                        godkänt ─────────────┴──▶ "verifierat"
                                        avslaget ────────────────▶ "deklarerat"
```

Agenten får lämna bevis och köra verifieringen. Den får **inte** godkänna
eller avslå omdömet - det är human gate, precis som att godkänna ett
förslag. API:t blockerar agentens nyckel från båda anropen.

## Åtgärdsprofilen - vad agenten får göra

Profilerna är tekniska gränser, inte instruktioner till modellen. De
väljs per körning och följer med i målets definition.

| Profil | Får |
| --- | --- |
| `planering` | Prata, bryta ner, föreslå - skriver ingenting |
| `research` | Läsa och samla - skriver ingenting |
| `arbete` | Arbeta i en isolerad kopia, skriva kod, föreslå ändringar |

Standardprofilen är `planering`. Kodaren körs bara när profilen får
ändra kod - så en livsgoal kan inte råka dra in kodaren.

## Måldefinitionen - vad agenten föreslår

I `define`-fasen föreslår agenten hur målet ska se ut. Det är inte bara
titel och beskrivning - även ikon, `now`, `eternal`, `agent`, startdatum,
due, uppskattad tid, status, taggar, platser och upprepning.

Förslaget lämnas som ett godkännande i Plan, med typen `goal_definition`.
Kodförslag har typen `code_change` - de är olika sorts beslut och blir
inte ihopblandade.

**Agenten fyller luckor.** Ett fält som redan har ett värde lämnas
orört - agenten skriver aldrig över det du själv har satt. Taggar och
platser slås ihop med målets egna, så inget användardata försvinner. Ett
värde som inte går att tolka (ett trasigt datum, en okänd ikon) hoppas
över och skrivs som en notering i förslaget.

När du godkänner skrivs exakt de värden du godkände till målet - ingenting
genereras om. Ett avslag stoppar körningen; agenten föreslår inte om något
utan ett nytt beslut från dig.

Målet flyttas till `plan` när definitionen är applicerad. Först då kan
agenten föreslå en kodändring.

## Verifiering före godkännande

Ett förslag testas med sitt eget testkommando innan det skickas till
en människa. Om testet inte går igenom får modellen först skriva om
testkommandot (innehållet behålls) och därefter skriva om hela
förslaget. Förslaget du ser är alltså självverifierat, och utfallet
redovisas i godkännandet.

```text
förslag ──▶ kör testet ──▶ går det igenom? ──ja──▶ human gate
                            │
                            nej
                            ▼
              skriv om testet ──▶ skriv om förslaget ──▶ kör testet igen
```

## Kostnad och budget

Agenten rapporterar tokenanvändning efter varje körning: input-,
output- och cache-tokens, antal verktygsanrop och vilka verktyg som
användes. Plan räknar sedan fram kostnaden ur en versionsstyrd
prislista, så att samma körning alltid ger samma summa.

Budgetar (`agent_budgets`) sätter gränser per period - dag, vecka eller
månad - och kan gälla en enskild agent eller alla agenter. En pågående
körning reserverar kostnad, så att samtidiga körningar inte
tillsammans passerar gränsen.

Kostnaden syns på `/agent-cost`, filtrerbar per agent och modell.

## Kontraktet mellan Plan och agenten

`src/contracts/plan-agent.yaml` beskriver allt agenten får göra:

- faserna och tillstånden,
- vilka triggers och profiler som finns,
- vilka API-anrop agenten gör - markerade med `agent: true`.

Filen är versionsstyrd och båda sidor testar mot den. Ändras
kontraktet på ett sätt som kräver samtidig ändring på båda sidor,
bumpas versionen. I produktion har agentens nyckel scope `agent`, så
Plan måste tillåta exakt de anrop som står med `agent: true` - annars
blir det 403 i prod men inte i dev (där nyckeln har scope `full`).

## Kunskapswikin

Alla repon delar en kunskaps-wiki som ligger i `wiki/` och hanteras av
`wikictl`. Agenten läser och skriver i den, precis som du gör:

```bash
wikictl search "bigram"        # sök i hela wikin
wikictl note --title "Rubrik" --tags "bug" --body "Vad som hände."
wikictl ingest                 # flytta inboxen in i grafen
wikictl lint                   # hitta trasiga länkar och gammal info
```

Wikin innehåller råkällor (`raw/`, immutable) och syntes (`wiki/`).
Den gör att kunskapen överlever mellan sessioner, agenter och modeller.
Mer om den i `~/wiki/AGENTS.md`.

## Miljövariabler

| Variabel | Betydelse |
| --- | --- |
| `PLAN_API_URL` | Plan API:t som agenten pratar med |
| `PLAN_API_KEY` | Agentens egen nyckel mot Plan |
| `AGENT_API_KEY` | Nyckel som Plan använder mot agent-API:t |
| `AGENT_REPO_PATH` | Sätt den för att kunna röra kod lokalt |
| `AGENT_GOAL_ID` | Kör workern mot ett enda mål |
| `AGENT_JOB_MAX_CONCURRENT` | Hur många jobb som får köras parallellt |
| `AGENT_JOB_NAMESPACE` | Namespace som jobben skapas i |
| `AGENT_JOB_DEADLINE_SECONDS` | Tidsgräns för ett jobb |
| `AGENT_WORKER_ENABLED` | Kör workern i agent-API:t (off i prod) |
| `AGENT_SCHEDULER_ENABLED` | Kör den gamla filbaserade schemalaggaren (off i prod) |
| `AGENT_JOB_CONTROLLER_ENABLED` | Skapa Jobb för köade körningar |

## Se också

- [Agentkörningar och godkännanden](/dokumentation/agent-korningar) - det
  kompletta flödet från mål till Pull Request.
- [API och autentisering](/dokumentation/api) - anropen agenten gör.
