---
title: Flyttalsparitet och orchestratorns döda gate-flöde
type: note
tags: [rust, paritet, agent, projekt:plan]
---

Under migreringen av Go-tjänsterna till Rust visade paritetsskripten två
saker som var lätta att missa.

**Flyttal skrivs olika.** Go skriver ett heltal flyttal som `1`, medan
serde_json skriver `1.0`. Det syntes i kostnader (`reserved_cost_usd`),
budgetgränser (`hard_limit_usd`) och i agentens framsteg i procent.
Paritetsskriptet jämför svaren fält för fält, så Rust behövde en egen
serialiserare: `src/rust/crates/plan-api/src/models/go_num.rs` och
`src/rust/crates/plan-agent/src/go_num.rs`. Samma lösning i båda
tjänsterna - de är separata Go-moduler, och de är separata Rust-crates.

**Orchestratorns gate-körning är död kod.** Reachabilityanalysen
(`scripts/go-reachability.py`) visar att `internal/core/gate` *nås* från
orchestratorn, men ingen körande tjänst anropar `Orchestrator.Execute` -
det är bara `executePlan` som gör det, och den kallas ingenstans.
`SubmitHumanDecision` är dessutom en no-op i Go, och
`NewAgentWithOrchestrator` (det enda stället där lärdomarna skulle
injiceras i prompten) kallas inte heller. Därför portades varken
gate-motorn eller subagenterna: de hade blivit död kod i Rust också, och
Rust hade betett sig annorlunda än Go om de börjat köra.

Det som faktiskt är nåbart - `CreatePlan`, `ApprovePlan`, `RejectPlan`,
`GetCurrentCheckpoint`, `GetCheckpointProgress` och `ListPlans` - finns i
Rust och gav samma svar som Go när de jämfördes live.

Lärdomshanteraren däremot porterades (`src/rust/crates/plan-agent/src/learnings.rs`),
eftersom den har egen tillstånd att spara. Filens format är Go:s: rubriken
är `## datum: titel` och den senaste lärdomen läggs först. Utskrifterna
jämfördes rad för rad mot Go - även de knepiga fallen, som att
`GetLearningsByCategory` bara returnerar `# Learnings`-raderna.

**Varför:** paritet handlar om beteende, inte om kodmängd. Att porta kod
som aldrig körs ger bara två risker: den rostar, och den riskerar att
beteendet glider isär från Go utan att någon märker det.

**Hur man fortsätter:** när Go-modulen tas bort raderas de döda
paketen med - de behöver ingen Rust-motsvarighet. Om gate-flödet någon
gång ska köras måste det beslutas först, och då portas det med sin egen
paritetskontroll.
