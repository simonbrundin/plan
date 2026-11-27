---
title: Beslut: atgardsprofil per uppgift, valfri verifiering och enhetlig malterminologi
type: note
tags: [agent, goals, beslut, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/anvanda-agenter-i-plan, notes/anvandarytan-for-agenten-ska-vara-event-driven-och-asynkron, notes/malmaskin-och-verifierare-fasmaskinen-bor-i-plan-inte-i-prompten]
---

ANVANDARENS SVAR (2026-11-27) pa designfragorna for det agentiska malsystemet, och justeringar de gav:

1. MALEN SPANNER HELA LIVET plus manga KODPROJEKT, och malen har manga olika typer. FORSTA FORSLAGET (goals.execution_kind = kod | liv) AVVISADES av anvandaren - delningen kändes onaturlig och konstig. JUSTERING: malet har ingen typ. MALETS AMNE ar en TAGG (taggar finns redan: "ute", "hem", "vinter"). Det som styr i stallet ar AGENTENS ATGARDSPROFIL per KORNING: planering (default - pratar, bryter ned, foreslar, skriver ingenting), research (laser och samlar underlag) eller arbete (isolerad kopia, skriver kod, branch + PR). Profilen foreslas av agenten i definitionen och godkanns av anvandaren. Den aldrig vaxer utan godkannande. profile.Coding/profile.Life blir implementationdetail under den.
2. VERIFIERINGEN SKILJER SIG PER MAL OCH AR VALFRI. Beviskallor (ingen Home Assistant - uttryckligen ratat): checklista, delmal, kalenderhändelse, bevis (text/fillank), matvarde, kodstatus (branch/PR/CI). Varje verifierare har ett villkor och en styrka (stark/svag). MAL UTAN VERIFIERARE AR OK - da far anvandaren bara trycka "jag ar klar" (t.ex. malet "Byt globdlampa": du byter lampan och sager att den ar bytt, sa ar den det). Verifieringsmaskineri ska aldrig gora en enkel mal tungt.
3. AVSLUTANDE HAR TVA VAGAR: (A) mansniskan avslutar - en handling, valfri fas, ingen verifiering kravs, registreras som "deklarerat"; (B) agenten foreslar - da kravs att verifieringen (om malet har en plan) gatt igenom, och forslaget ska godkannas av mansniska. completed registrerar ALLTID hur det avslutades (deklarerat | verifierat).
4. GODKANNANDE: huvudregeln - agenten far inte agera pa ett mal innan anvandaren godkant forslaget. Allt som LAMNAR DATORN (kopa, boka, skicka, ringa) kräver ALLTID eget godkannande, varje gang, oavsett installning. Valfri spets per mal: approval_mode = plan (default) | varje_atgard.
5. AGENTEN KOR I KLUSTRET. Ingen HA-koppling, inga sensorer, inga extra hemligheter i klustret.
6. ATERKOMMANDE MAL: verifieringen galler VARJE FOREKOMST, medan verifieringsplanen ligger pa malet (malet har redan goal_completions per forekomst).

TERMINOLOGI: allt agenten skapar ar ett MAL. Agentens nedbrytning skapar undermal via goal_relations - ingen separat checkpoint-modell i malmodellen (checkpoint-ordningen i docs/checkpoint-gate-workflow.md forblir en intern detalj i kodagentens gate-flode). Anvandaren sager själv att han kallar allt mal oavsett var i hierarkin det ligger - da far agenten inte introducera ett annat ord for samma sak.

SPEC: openspec/changes/add-goal-verification-and-phase-transitions (fasmaskin + valfri verifiering + atgardsprofil + godkannandemode). Validerad med openspec validate --strict.

Relaterat: [[notes/malmaskin-och-verifierare-fasmaskinen-bor-i-plan-inte-i-prompten]] och [[notes/anvandarytan-for-agenten-ska-vara-event-driven-och-asynkron]].

## Relaterade

- [[concepts/anvanda-agenter-i-plan]]
- [[notes/malmaskin-och-verifierare-fasmaskinen-bor-i-plan-inte-i-prompten]]
