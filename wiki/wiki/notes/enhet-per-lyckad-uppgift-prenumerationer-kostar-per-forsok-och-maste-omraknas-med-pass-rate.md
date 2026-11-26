---
title: Enhet per lyckad uppgift: prenumerationer kostar per försök och måste omräknas med pass rate
type: note
tags: [dashboard, kostnad, enhet, metodik, benchmark, prenumeration, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Ett byte som är lätt att missa: planpriset sprids över ANTALET AGENTKÖRNINGAR (försök), medan API-kostnaden jämförs som kostnad per LYCKAD UPPGIFT. Att jämföra /bin/bash.13 per försök (plan) med .7 per lyckad uppgift (API) blandar enheter - det siffror ut som om planen vore ~60x billigare än den är.

Fixa: planen prissätts per försök (planpris / antal körningar per månad), sedan divideras med pass rate från samma benchmark för att få kostnad per lyckad uppgift. Pass rate kommer ur samma mätning (Terminal-Bench 65.2 %, DeepSWE), sa den ar en giltig omrakningsfaktor. Saknas pass rate (AA Coding Agent Index) far allt sta som per forsok och markeras sa.

Beslutet galler: computePlanCost() tar ALLTID en kostnad per forsok. recommendation.ts raknar om planens resultat med costPerSolvedTaskFrom(). En enhetslucka i den har en gang givit dubbelmaat 'per lyckad uppgift per lyckad uppgift' - enheterna ska skrivas pa ett stalle, inte tvast.

Relaterat: basketet for enhetsvyn ar 1M input + 1M output, och cachevyn anvander 90 % cache-hit som dokumenterat standardantagande for modeller utan egen matning.
