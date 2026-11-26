---
title: Dashboard ranking utökad med benchmark- och kostnadstäckning
type: note
tags: [dashboard, ranking, benchmark, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Dashboardens modellranking använder nu separata, effort-matchade data för DeepSWE v1.1, Terminal-Bench 4.1/4.0, AA Coding Agent och BenchLM. Terminal-Bench/DeepSWE rankas på benchmarkets egna kostnad per lyckad uppgift; BenchLM är kvalitet-only. En separat uppskattad vy använder tokenprofiler och modellpriser och markeras som estimation. Importen verifierades med 141 modeller: 24 Terminal-Bench 4.1-par, 24 Terminal-Bench 4.0-par, 14 DeepSWE-par, 39 AA Coding Agent-par, 137 uppskattade kostnader och 55 modeller med minst en uppmätt score/kostnad-kombination. Tester, riktad lint och production build passerar; full typecheck/lint har kvar tidigare fel i Talos-koden.
