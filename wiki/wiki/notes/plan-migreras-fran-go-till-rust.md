---
title: Plan migreras från Go till Rust
type: note
tags: [migration, rust, api, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Beslut 2026-11-26: Plan-applikationens Go-tjänster skrivs om i Rust. Nuxt-gränssnittet (Nuxt 4, SPA) lämnas orört - det påverkas inte av migreringen.

Varför: enhetlig stack, bättre typgaranti och Rust-arbetsflödet (cargo, clippy, rustfmt) matchar agentens verktygskedja.

Omfattning: två Go-tjänster - src/api (REST-API, ~16 kLoC) och src/agent (agentkörningar, ~40 kLoC). Ordningen är API först, agent senare, eftersom agenten pratar med API:t via kontraktet src/contracts/plan-agent.yaml.

Nytt workspace: src/rust/ (Cargo) med crates/plan-api och senare crates/plan-agent. Go-tjänsten ligger kvar tills fas 5 (utfasning), så båda kan köra parallellt mot samma databas.

Paritet är kravet: samma vägar, metoder, statuskoder och JSON-format. Ett test i Rust jämför route-tabellen mot route-registreringarna i src/api/cmd/api/main.go, så en endpoint som glöms bort failar i CI.

Kända risker: RRULE (rrule-go -> egen motor behövs, rrule-krigen är svagt underhållt), k8s i agenten (client-go -> kube-rs), Anthropic-SDK (anthropic-sdk-go -> råa HTTP-anrop). swagger.json i repot är stale (innehåller gamla agent-UI-endpoints som /workflow och /conversations) och måste regenereras i fas 3.

Spec: specs/rust-migration.md.
