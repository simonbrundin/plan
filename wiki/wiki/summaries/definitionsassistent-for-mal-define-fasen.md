---
title: "Sammanfattning: definitionsassistent för mål (define-fasen)"
type: summary
summary: En egen definitionsassistent i Plan - synkront LLM-anrop i Plan-API:t, förslag som fyller bara luckorna, godkännande per fält, och en mjuk varning i stället för ett hinder när fasen byts.
tags: [goals, agent, human-gate, feature, projekt:plan]
sources: [raw/2026-11-27-definitionsassistent-for-mal-define-fasen.md]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [notes/agentens-maldefinition-gar-via-human-gate, notes/definitionsassistent-for-mal-fylla-define-fasen-med-godkannande-per-falt]
---

## Källa

`raw/2026-11-27-definitionsassistent-for-mal-define-fasen.md` (origin: https://github.com/simonbrundin/plan)

## Huvudpunkter

1. **Diagnosen:** agentens definitionsmotor är byggd för kodagenten
   (`goals.agent = true` -> `agent-worker` -> Kubernetes-Job -> polling), inte
   för människan som vill definiera ett mål. Därför står nya mål kvar i
   `define` utan hjälp.
2. **Godkännandet är fel nivå:** dagens `approve` applicerar hela
   values-paketet. Det som behövs är godkännande **per fält**.
3. **LLM:anropet flyttas till Plan API:t** (`src/api/internal/llm`), synkront.
   `src/api` och `src/agent` ligger i samma Go-modul och
   `anthropic-sdk-go` finns redan i require-listan.
4. **Förslaget lever i klienten** tills användaren godkänner; servern
   applicerar exakt de godkända värdena. Appliceringen loggas i
   `goal_definition_applies`.
5. **Definitionen blockerar inte fasbytet** - bara en varning ("N fält
   saknas"). Detta var ett uttryckligt beslut.
6. **Verifieringskriterium** (`goals.criteria`, fritext) är en del av
   definitionen; för kodmål ≈ acceptanskriterium/TDD. Fältet är fritt.
7. **Tangent `Space+f`** - `d` var upptaget ("toggle slutfört").
8. **Policyn flyttas till `src/internal/goaldefinition`** så att Plan API:t och
   agenten inte får två olika definitioner av "fylla luckor".

## Kopplingar

- [[notes/definitionsassistent-for-mal-fylla-define-fasen-med-godkannande-per-falt]]
- [[notes/malfaserna-ar-gemensamma-for-alla-mal-i-plan]]
- [[concepts/anvanda-agenter-i-plan]]

## Relaterade

- [[notes/agentens-maldefinition-gar-via-human-gate]]

