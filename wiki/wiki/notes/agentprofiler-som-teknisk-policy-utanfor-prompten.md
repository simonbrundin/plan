---
title: Agentprofiler som teknisk policy utanför prompten
type: note
tags: [plan, agent, profil, policy, behörighet, säkerhet, projekt:simon]
created: 2026-11-24
updated: 2026-11-24
status: active
---

I ~/repos/plan (Fas 4 i mål 497) ligger agentprofilerna i src/agent/internal/profile och är tekniska gränser - inte text i prompten.

Varför: en profil som bara beskrivs i prompten är bara en önskemål. Modellen kan glömma, och en bugg i prompten blir en säkerhetslucka. Profilen ska därför slå igenom i klustret.

Profilen tillämpas i tre lager:
1. Jobbet (jobrunner) - kodprofilen får klona repot och får GIT_AUTH_HEADER, livsprofilen får inte det. Livsprofilens jobb har dessutom mindre resursgränser.
2. Miljön - AGENT_REPO_PATH sätts bara för profiler som får röra kod, så workern inte ens har en sökväg till repot.
3. Workern och codern - codern sätter sig själv ur vägen om profilen inte tillåter kodändringar, och körningen avslutas med en tydlig anledning i stället för att lämnas öppen.

En okänd profil ger kodprofilen (den strängaste). Det är viktigt: ett fel ska aldrig ge fler rättigheter än avsett.

Lärdom: när workern byggdes för första gången med en ny profil panikade den - kodfasen anropades utan att codern fanns. Profilens gränser måste gälla i varje lager, inte bara i jobbet.

Relaterat: varje körning får sin egen branch (agent/goal-<mål>-run-<körning>). En gemensam branch per mål krockar mellan försök, och en push som inte är fast-forward avvisas - då sågs ändringen som "gick inte att publicera" trots att allt annat fungerade.
