---
title: Fasmaskinen for mal ar nu implementerad i Plan API
type: note
tags: [agent, goals, implementation, verifiering, human-gate, projekt:plan]
---

VAD: fasmaskinen och den valfria verifieringen finns nu i Plan API (branch feat/goal-phase-machine, OpenSpec-change add-goal-verification-and-phase-transitions).

VARFÖR: designnoten "Malmaskin och verifierare" bestämde att faserna ska vara deterministiska villkor i API:t, inte modelltext. Det är nu genomfört - och två regler skärptes under implementeringen, eftersom de visade sig vara för lösa i designen:

1. Agenten får inte hoppa över faser (define -> execute direkt). Faserna är kontraktet mellan agenten och Plan, så en hoppning stoppas med koden phase_skipped. Designen hade bara sagt "aldrig bakåt".
2. En agent får aldrig avsluta ett mål själv - varken via fasendpointe eller via den gamla PATCH /status. Den får föreslå, en människa bestämmer (409 approval_required). Gränsnittet har fortfarande genvägen "jag är klar" för människan.

SÅ SER DET UT: maskinen är ren logik i src/api/internal/phase (inga beroenden till databas), bindningen ligger i src/api/internal/phaseservice. Både POST /api/v1/goals/:id/phase och den gamla PATCH /api/v1/goals/:id/status går via maskinen - annars hade den kunnat kringgås. Stoppade overgangar svarar 409 med en stabil skal-kod (definition_not_approved, missing_planned_children, unresolved_children, phase_skipped, backwards_transition, approval_required).

VERIFIERING: planen är valfri - mal utan plan fungerar precis som mal med plan. Data ligger i goal_verifications, goal_evidence, goal_verification_results och goal_phase_events; malets avslutande registreras i goals.completion_kind (deklarerat | verifierat). Ett omdome blir giltigt forst nar en manska godkanner det - agent-scope blockerar approve-endpointen i API:t, inte i prompten. Nar planen andras ogiltigforklaras tidigare omdomen.

GALLER FRAMOT: agenten (block 4 i tasks.md) ska ga via fasendpointe, respektera 409 och skapa godkannanden i stallet for att avsluta mal. Gransnittet (block 6) ska visa "jag ar klar", verifieringsplan, bevis och omdome. Kodmalens verifierare (CI eller mansklig granskning) definieras fortfarande inte - typen "kod" finns i databasen, men innehallet kommer i en egen change.
