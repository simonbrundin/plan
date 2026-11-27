---
title: Definitionsassistent för mål (define-fasen)
type: source
added: 2026-11-27
origin: https://github.com/simonbrundin/plan
immutable: true
---

# Definitionsassistent för mål (define-fasen)

## Problem
Plan har fem målfaser (`define`, `plan`, `execute`, `verify`, `completed`) och en
agent som kan definiera mål. Men agentens definitionsmotor är hängd på
`goals.agent = true` och körs via `agent-worker` -> Kubernetes-Job -> polling.
Det är maskineri byggd för att en *kodagent ska arbeta på målet*, inte för att
*människan ska definiera ett mål*.

Följder:
- Ett nytt mål hamnar i `define` och står där - ingen hjälp att fylla fasen.
- Hjälp kräver agentflagga + agent-API igång + worker-poll + godkännande i
  agent-docken. Tungt för en snabb, mänsklig handling.
- Godkännandet är allt-eller-inget: ett `approve` applicerar hela
  values-paketet, inte fält för fält.

## Lösning (design)
En **definitionsassistent** i den vanliga appen:
1. Panel (från inboxens "Bearbeta" och från målsiden med `Space+f`).
2. Luckinventering - vilka fält saknar värde.
3. Förslag från LLM som fyller ENDAST luckorna (samma policy som agentens
   `goaldefinition` idag: skriv aldrig över användarens värde).
4. Diff-vy `från -> till` per fält; acceptera/ändra/avvisa per fält.
5. Apply - server skriver EXAKT de godkända värdena, inget annat.

Nyckelbeslut:
- **LLM:anropet sker i Plan API:t** (`src/api/internal/llm`), synkront. Inte via
  agent-tjänsten - att definiera är en snabb mänsklig handling och ska inte stå
  i vägen för en worker och ett Kubernetes-jobb. `src/api` och `src/agent` ligger
  i samma Go-modul (`src/go.mod`, `module plan`) och `anthropic-sdk-go` finns
  redan i require-listan, så det är billigt.
- **Förslaget lever i klienten** tills godkännande skickas. "Godkänna allting"
  blir bokstavligt. Appliceringen loggas i `goal_definition_applies`.
- **Definitionen blockerar INTE fasbytet.** `define -> plan` går alltid att
  göra; en varning ("3 fält saknas") visas i stället. Användaren bestämde detta
  uttryckligen.
- **Verifieringskriterium är en del av definitionen** (`goals.criteria`,
  fritext). För kodmål ≈ acceptanskriterium/TDD ("testet är grönt"), vilket
  binder ihop define/plan/execute/verify. Fältet är fritt - tomt är inte fel.
- **Tangent `Space+f`** (leader + f). `d` var upptaget ("toggle slutfört" på
  målsidan, `Space+d` = startdatum/deadline). `f` är ledig; leader-läget är
  avsett för sekundära åtgärder.
- Policyn flyttas till ett gemensamt paket `src/internal/goaldefinition` så att
  Plan API:t och agenten inte får två definitioner av "fylla luckor".

## Leverans
OpenSpec-change `add-goal-definition-assistant` i plan-repot:
proposal.md, design.md, tasks.md, specs/goal-definition-assistant/spec.md.
Validerad med `openspec validate add-goal-definition-assistant --strict`.

## Koppling
`add-goal-verification-and-phase-transitions` (öppen change) föreslår
strukturerade verifieringsplaner (`goal_verifications`) och en fasmaskin där
`define -> plan` kräver "godkänd agentdefinition". Den här changeen ger samma
definition men som ett godkännande i gränssnittet, och fasmaskinen ska INTE
kräva den. Inte blockerande - de två förslagen ska samordnas.
