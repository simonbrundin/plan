---
title: Paritetstest gatear Rust-migreringen mot Go-källan
type: note
tags: [rust, migration, test, ci, projekt:plan]
---

Under migreringen från Go till Rust jämförs Rust-tabellerna mot route-registreringarna i Go-källan programmatiskt, i CI.

Varför: en endpoint som glöms bort i migreringen ger 404 i gränssnittet eller i agenten - och det upptäcks först i produktion. Ett test som läser Go-källan gör felet till ett rött CI-svar istället.

Hur: crate:en src/rust/crates/plan-parity läser Go-källan (src/api/cmd/api/main.go och src/agent/internal/api/server.go), plockar ut grupperna (api := r.Group("/api/v1")) och deras routes, och jämför dem med Rust-tabellerna. Båda vägar testar: en route som saknas i Rust och en route som finns i Rust men inte i Go.

Fällor som kostade tid:
- Go registrerar routes över flera rader (r.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{). Rad-för-rad-parsning missar dem - parsningen måste ske över hela källan.
- Gruppprefixet måste tas med: Go-registreringarna är relativa till gruppen ("/goals"), Rust har fullständiga sökvägar ("/api/v1/goals").
- Handler-anrop (api.GET("/goals", goalHandler.GetGoals)) är inte routes - bara gruppernas egna anrop ska räknas.

Kontraktet mellan Plan och agenten (src/contracts/plan-agent.yaml) testas nu från båda sidor i Rust: agentens planklient (crates/plan-agent/tests/contract.rs) och Plan API:t (crates/plan-api/tests/contract.rs).

CI: .github/workflows/commit.rust-api.yaml kör fmt, clippy -D warnings och cargo test för hela workspace:t.
