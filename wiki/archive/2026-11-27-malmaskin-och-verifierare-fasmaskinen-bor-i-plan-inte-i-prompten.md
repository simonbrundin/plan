---
title: Målmaskin och verifierare: fasmaskinen bor i Plan, inte i prompten
type: note
tags: [agent, goals, design, verifiering, human-gate, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/anvanda-agenter-i-plan, notes/anvandarytan-for-agenten-ska-vara-event-driven-och-asynkron, notes/beslut-atgardsprofil-per-uppgift-valfri-verifiering-och-enhetlig-malterminologi]
---

ÖNSKEMÅL: ett agentiskt system som tar mål från define till completed OCH verifierat.

VARFÖR: Plan har redan maskineriet (mål, faser, agent_runs, godkännanden, human gate via API-scope, kostnad i Plan, kontraktet src/contracts/plan-agent.yaml). Det som saknas är inte plattformen utan loopen för icke-kodmål: workern går idag define -> plan och sedan sköter coder kodändringar. profile.Life finns som policy (src/agent/internal/profile/profile.go) men har ingen exekveringsmotor bakom sig.

BESLUT (design):
1. Fasmaskinen (define -> plan -> execute -> verify -> completed) ska ligga i Plan API som deterministiska transitions med villkor, inte som modelltext. Modellens svar ska aldrig kunna skriva statusen fritt - annars är faserna tafelexempel.
2. Varje mål kan fa en VALFRI verifieringsplan (typ: checklista / delmal / kalender / bevis / matvarde / kod, plus villkor t.ex. "3 av 4 denna vecka" och styrka stark/svag). MAL UTAN VERIFIERARE AR OK - da avslutar mansniskan med en handling ("jag ar klar", t.ex. malet "Byt globdlampa"). Nar en plan finns blir den villkoret for att AGENTEN far foreslå completed, och beviset visas i godkannandet. Det gor "verifierat" till en egenskap i datamodellen (t.ex. goals.verification) i stallet for en agents kansla. INGEN Home Assistant - bevisen kommer fran anvandaren, kalendern, incheckningar och for kodmal repot/CI.
3. Fyra roller i samma motor (samma kontrakt, budget och human gate): definera (definition + verifierare), planera (nedbrytning till delmal med datum), utfora (paminnelser, forberedelser, deadline-justering), verifiera (samla bevis, kor verifieraren, foresl completed eller "vad saknas").
4. Human gate forblir strukturellt (agent-scope kan inte besluta). Huvudregel: agenten far inte agera pa ett mal innan anvandaren godkant forslaget - sedan far agenten arbeta inom det. ALLT som lamnar datorn (bokning/kop/ringa/skicka) krav ALLTID eget godkannande, varje gang. Valfri spets per mal: varje atgard godkannas separat.
5. Forbud: ingen separat chatbot bredvid Plan, ingen agentdatabas for mal (Plan ager), ingen kontinuerlig polling - event-driven i stallet.

RISK: en agent som flyttar mal kan stressa anvandaren. Därfor max N aktiva mal i agentens arbete och agenten ska hellre dela upp an dela pressa.

Relaterat: [[notes/anvandarytan-for-agenten-ska-vara-event-driven-och-asynkron]], [[notes/beslut-atgardsprofil-per-uppgift-valfri-verifiering-och-enhetlig-malterminologi]] och [[concepts/anvanda-agenter-i-plan]].

## Relaterade

- [[concepts/anvanda-agenter-i-plan]]

