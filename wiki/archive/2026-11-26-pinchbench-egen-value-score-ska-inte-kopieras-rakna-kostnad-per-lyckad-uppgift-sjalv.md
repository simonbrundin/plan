---
title: PinchBench: egen Value Score ska inte kopieras - räkna kostnad per lyckad uppgift själv
type: note
tags: [dashboard, modeller, benchmark, pinchbench, research, projekt:dashboard]
---

Anledning: användaren visade pinchbench.com/?view=value och frågade vad jag tycker.
Svaret blevi: bra benchmark, men Value-vyn är metodiskt svag.

Verifierat mot api.pinchbench.com (offentligt API, ingen auth):
- /api/leaderboard?official=true&version=2.0.0
- /api/submissions/:id med per-uppgift-data, usage_summary och metadata.suite

Kallan: notes/pinchbench-verifierade-api-fakta-och-datafel-i-value-vyn (raw).

Det viktiga fyndet: Value Score = Success % / Best Cost, och de två "best"-värdena
kommer från två olika körningar. Kostnaden är inte normaliserad per uppgift, så en
1-uppgifts community-körning (1/147 av kostnaden) ser extremt billig ut.
Verifierat: gemma-4-26b-a4b-it har best_score 100 % från en 1-uppgifts körning, medan
den officiella 147-uppgiftskörningen gav 77,06 % för $2,03 (113,3 lösta -> 55,8 lösta/$).
PinchBench visar 173,1. CPST-kolumnen är "N/A" för alla rader - de räknar aldrig ut
kostnad per lyckad uppgift.

Grundproblemet är detsamma som vi redan löst i dashboarden: en siffra som blandar
mätning och antagande, eller som cherry-pickar två olika körningar, får inte användas
som beslutsunderlag.

Varför PinchBench ändå är intressant för oss:
- publikt API utan auth, filtrerbart på official + version
- metadata.suite == "all" gör att delkörningar kan filtreras bort
- per-uppgift-poäng -> vi kan räkna en riktig pass rate (score >= max_score) och ett
  riktigt konfidensintervall. Det vore den första benchmarken i dashboarden där CI-kolumnen
  fylls på riktigt (AA saknar varians, DeepSWE har bara en).
- harnessversion och benchmarkversion per körning
- thinking-level som egen dimension
- täcker DevOps/k8s/Dockerfile/CI-debug - komplementärt till repo-nivaens SWE-benchmarks

Plan: importera som femte benchmark "PinchBench 2.0.0 (OpenClaw)", med egna matt:
pass rate fran fullpoang, kostnad per lyckad uppgift fran samma koring, CI fran
per-uppgiftsfordelningen, endast official + suite=all + en benchmarkversion, egen
harnessrad (aldrig blandad med Terminal-Bench/DeepSWE).
