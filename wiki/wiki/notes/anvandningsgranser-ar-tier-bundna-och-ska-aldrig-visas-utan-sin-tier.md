---
title: Användningsgränser är tier-bundna och ska aldrig visas utan sin tier
type: note
tags: [dashboard, modeller, granser, kvoter, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Dashboardens modellsida visar nu publicerade användningsgränser per modell: timgräns
(anrop/min, token/min, samtidiga), dagsgräns (token/dag) och månadsgräns för spend.

Varför detta behövdes: gränser är inte en siffra utan ett paket - de gäller en viss
tier hos en viss leverantör, och de flesta leverantörer låter dem växa med spend. En
gräns utan sin tier är vilseledande: Kimis 3 anrop/min gäller Tier 0 och är betydligt
högre på nästa nivå, och Anthropics \$500/mån är en spend-cap för organisationen, inte
en inkluderad kvot.

Principer som följdes:
- Ej publicerade gränser skrivs "ej publicerade", aldrig 0 och aldrig "obegränsad".
- "Obegränsat" används bara när leverantören uttryckligen säger det (t.ex. Kimi
  Tier 1+, där dagsgränsen faller bort).
- Prenumerationens kvot hålls skild från API-gränserna: kvoten gäller agentanvändningen
  i apparna, API-gränserna gäller API-nivån.
- En prenumeration täcker modellen i alla reasoning-effort-nivåer, eftersom
  katalogens "-high"/"-max"-rader är samma modell i olika inställningar. Detta hade
  tidigare gjort att effort-varianterna saknade plan i gränsvyn.
- Tierinformationen visas alltid bredvid siffran, inklusive när en högre tier är
  obegränsad ("1,5M tokens/dag (Tier 0) → obegränsat vid högre nivå").

Kod: src/nuxt/app/data/usageLimits.ts (data), src/nuxt/app/utils/usageLimits.ts
(logik), src/nuxt/app/components/models/ModelsUsageLimits.vue (vy),
src/nuxt/tests/usageLimits.test.ts (69 tester totalt gröna).
