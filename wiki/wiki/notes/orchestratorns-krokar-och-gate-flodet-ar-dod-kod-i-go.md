---
title: Orchestratorns krokar och gate-flödet är död kod i Go
type: note
tags: [agent, rust, paritet, gates, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
---

# Orchestratorns krokar och gate-flödet är död kod i Go

Go-agentens orchestrator har en hel uppsättning krokar som aldrig körs:

- `NewAgentWithOrchestrator` anropas ingenstans. Det är den enda vägen
  till `PrepareNextTurn` och till `injectLearningsIntoPrompt`, så
  lärdomarna når aldrig systemprompten och checkpointens framsteg
  uppdateras aldrig under en chatt.
- `orchestrator.ExecuteCheckpoint` - det som kör gatearna - anropas inte
  heller. `checkpoint_executor.go`, `subagent.go` och `test_planner.go`
  saknar därför körande anropare.
- `NewCheckpointExecutor` har inga anropare alls.

Det är samma slags död kod som gate-paketet: reachability-analysen
(`scripts/go-reachability.py`) når paketen, men inte funktionerna.

## Varför det spelar roll i migreringen

Rust ska svara precis som Go. Därför har Rust samma mekanik, men
avstängd som standard:

- `AGENT_ORCHESTRATOR_HOOKS=true` slår på lärdomarna i systemprompten,
  checkpointens framsteg efter varje tur och instruktionen att använda
  `ask_human_approval` när en human-gate väntar.
- Utan variabeln är beteendet identiskt med Go: inga lärdomar i prompten,
  ingen framstegsuppdatering.

Texterna är ordagrant Go:s `buildCheckpointProgress`,
`updateCheckpointProgress`, `buildHumanApprovalInstruction` och
`hasPendingHumanGate`.

## Gate-utförandet behöver inte portas

Eftersom inga gates körs i Go behöver Rust inte implementera
gate-utförningen, subagenter eller testplannern. Det räcker att Rust
bevar samma no-op-beteende: en plan skapas utan checkpoints, eftersom
ingen checkpointplanner är konfigurerad.

## Typfellen i databasen

Skuggtestet mot den riktiga databasen hittade en panic: Rust läste
`goal_dependencies.id` som i64, men kolumnen är INTEGER. sqlx avkodar
INT4 som INT8 med `unwrap` → panic → anropet hängde. Go läser samma rad
med `lib/pq`, som klarar båda typerna.

Två lärdomar:

1. Ett skuggtest mot *riktig* data hittar fel som en lokal testdatabas
   inte gör - en tom tabell avkodas aldrig.
2. Rust läser kolumner med exakt typ. Go är flexibelt. Varje kolumn ska
   matcha schemat, och en panic ska inte få ett anrop att hänga - det
   finns nu en panikfångare i API:ts middleware-lager som svarar med 500.

## Kalenderhändelsernas ordning

Go lägger de overrides som inte matchar någon expanderad förekomst sist i
listan, i Go-kartans iterationordning - som är slumpmässig. Rust gör
samma sak, men i en annan slumpordning. Paritetsskripten sorterar därför
kalenderhändelser före jämförelsen, annars är jämförelsen flitig.
