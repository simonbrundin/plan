---
title: Användarytan för agenten ska vara event-driven och asynkron
type: note
tags: [agent, ui, usability, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/anvanda-agenter-i-plan, notes/malmaskin-och-verifierare-fasmaskinen-bor-i-plan-inte-i-prompten, notes/skrivbordswidget-som-visar-gor-nu-malet]
---

ÖNSKEMÅL: agenten ska vara användbar i vardagen, inte bara bra i teorin.

VARFÖR: agentens kvalitet är sekundär - om det tar 40 sekunder att svara pa agenten anvands den tre ganger. Risk: agenten ligger och vantar pa att man oppnar en webblasarflik, och da forlorar man malen mitt i allt.

BESLUT om ytor:
- Fangst-lage: en rad ("jag vill kopa en cykel") -> inbox -> agenten foreslar definition med verifierare -> godkannande med Enter. Ingen form, ingen wizard.
- /now som cockpit: dagens 1-3 checkpoints, nasta handling, knappar Klar / Skjuta upp / Hjalp. Skrivbordswidgeten (Quickshell) visar samma data.
- Asynkront: agenten ska aldrig blockera. Notis/tost i stallet for att man ska polla.
- Event-driven i stallet for kontinuerlig polling: vid faskifte, morgonpuls (t.ex. 07:00), kvallsreflektion (t.ex. 20:00 med "vad gick idag?"), och vid händelser utifran (t.ex. HA: dammsugare klar -> verifiera stadmaltet; kalender: traning passerat -> checka av).
- Liten mandat: agenten far flytta deadlines, foresla mindre delmal och uppdatera prognos utan att fraga - synlig logg, budgetgrans.
- Forklarlighet som krav: varje statusandring och varje forslag far en rad "varfor" i handelseloggen. utan fortroende slutar man anvanda systemet.

Relaterat: [[notes/malmaskin-och-verifierare-fasmaskinen-bor-i-plan-inte-i-prompten]] och [[notes/skrivbordswidget-som-visar-gor-nu-malet]].

## Relaterade

- [[concepts/anvanda-agenter-i-plan]]

