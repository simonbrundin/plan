---
title: Månadskostnad räknas på körningar, inte på lösta uppgifter
type: note
tags: [dashboard, modeller, enheter, bugg, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Buggen hittades när användaren frågade vad "agentkörningar" betyder - frågan avslöjade
att samma sak räknades på två sätt.

Enhetsmodellen på modelsidan:
- agentkörning = ett försök på en uppgift, kan misslyckas (det är det du betalar för)
- lyckad uppgift = en körning som löste uppgiften
- anrop = ett API-anrop; en körning består av många anrop

Felet: beslutstabellen räknade månadskostnaden som "kostnad per lyckad uppgift × 150"
medan planerna spred planpriset över "150 försök". Det överdrev API-kostnaden med
1/pass rate - 1,5x på Terminal-Bench ($1081/mån i stället för $699).

Fix: allt räknas över antal körningar, och antalet lösta uppgifter redovisas som en
egen siffra ("97 lösta/mån"). Kostnad per lyckad uppgift och månadskostnad står i
varsin kolumn.

Lärdom: när ett begrepp på en sida är oklart ("agentkörningar") är det ofta ett tecken
på att samma storhet räknas på olika sätt på olika ställen. Läs orden, inte bara
siffrorna.
