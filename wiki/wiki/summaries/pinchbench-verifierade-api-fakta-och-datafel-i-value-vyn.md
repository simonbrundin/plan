---
title: PinchBench: verifierade API-fakta och datafel i Value-vyn
type: summary
tags: [projekt:dashboard]
sources: [raw/2026-11-26-pinchbench-verifierade-api-fakta-och-datafel-i-value-vyn.md]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [notes/pinchbench-egen-value-score-ska-inte-kopieras-rakna-kostnad-per-lyckad-uppgift-sjalv]
---

PinchBench (Kilo Code) mäter LLM:er som OpenClaw-agenter. Verifierat mot det publika API:t: Value Score = Success % / Best Cost blandar två olika körningar och kostnaden är inte normaliserad per uppgift (gemma-4-26b-a4b-it har 100 % från en 1-uppgifts community-körning, officiellt 77,06 % för $2,03). CPST är 'N/A' för alla rader. Poängen är medel av delpoäng, inte pass rate. Användbar för oss: publikt API, metadata.suite=='all' för att filtrera delkörningar, per-uppgift-poäng för riktig pass rate och CI, harness- och benchmarkversion per körning, thinking-level som dimension.

## Relaterade

- [[notes/pinchbench-egen-value-score-ska-inte-kopieras-rakna-kostnad-per-lyckad-uppgift-sjalv]]

