---
title: Dashboard ranking använder benchmark-matchade kostnader
type: note
tags: [dashboard, ranking, benchmark, metodik, projekt:simon]
---

2026-11-26 ändrades modellrankingen i dashboarden så att pris och kvalitet inte längre blandas mellan olika benchmarkar. Före ändringen kunde t.ex. en BenchLM-familjepoäng divideras med en DeepSWE-kostnad och visas som pts/$. Terminal-Bench- och DeepSWE-poängen används redan som pass rate vid kostnad per lyckad uppgift; att dividera poängen en gång till dubbelräknar därför framgångsgraden.

Ny modell: välj benchmark separat. Terminal-Bench 4.0 och DeepSWE v1.1 rankas på kostnad per lyckad uppgift och visar sin egen score separat. AA Coding Agent Index använder bara sin egen score/kostnads-pair. BenchLM är score-only eftersom en matchad BenchLM-kostnad saknas. Pareto-frontier är per benchmark. BenchLM-familjematchningar och Estimated-resultat är opt-in; endast exakta Supported-varianter ingår som standard.

DeepSWE-score sparas nu i  så att den kan paras med samma benchmarkkostnad. Lokal verifiering efter import: 141 modeller, 14 DeepSWE-score/kostnad-pairs, 11 Terminal-Bench-pairs, 4 exakta Supported BenchLM-rader och 0 AA Coding Agent-pairs (scrapern hittade inga agents den här gången). AA-vyn är därför tom tills dess källa levererar matchade värden.
