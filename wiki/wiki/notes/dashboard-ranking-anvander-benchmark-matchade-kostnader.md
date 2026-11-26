---
title: Dashboard ranking använder benchmark-matchade kostnader
type: note
summary: Prisvärdhetsrankingen isolerar varje benchmark och undviker blandade score/kostnad-pairs.
tags: [benchmark, dashboard, metodik, projekt:dashboard, projekt:simon, ranking]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [notes/aa-data-api-v2-data-pensionerad-dashboard-migrerar-till-api-v2-language-models]
---

2026-11-26 ändrades modellrankingen i dashboarden så att pris och kvalitet inte längre blandas mellan olika benchmarkar. Före ändringen kunde exempelvis en BenchLM-familjepoäng divideras med en DeepSWE-kostnad och visas som pts/$. Terminal-Bench- och DeepSWE-poängen används redan som pass rate vid kostnad per lyckad uppgift; att dividera poängen en gång till dubbelräknar därför framgångsgraden.

Ny modell: välj benchmark separat. Terminal-Bench 4.0 och DeepSWE v1.1 rankas på kostnad per lyckad uppgift och visar sin egen score separat. AA Coding Agent Index använder bara sin egen score/kostnads-pair. BenchLM är score-only eftersom en matchad BenchLM-kostnad saknas. Pareto-frontier är per benchmark. BenchLM-familjematchningar och Estimated-resultat är opt-in; endast exakta Supported-varianter ingår som standard.

Benchmarkväljaren i Nuxt UI 4 måste använda USelect-propen items, inte options. Browser smoke test verifierade 14 DeepSWE-pairs, 11 Terminal-Bench-pairs, 4 BenchLM-rader som standard och 89 med familj/Estimated-valet aktiverat.

DeepSWE-score sparas i deep_swe_score så att den kan paras med samma benchmarkkostnad. Lokal verifiering efter import: 141 modeller; 0 AA Coding Agent-pairs (scrapern hittade inga agents den här gången). AA-vyn är därför tom tills dess källa levererar matchade värden.
