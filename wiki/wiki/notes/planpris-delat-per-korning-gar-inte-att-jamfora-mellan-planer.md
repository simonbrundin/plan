---
title: Planpris delat per körning går inte att jämföra mellan planer
type: note
tags: [dashboard, modeller, kvoter, enheter, presentation, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Anledning: användaren frågade "menar du att Gemini 4 Neon är lika billig som MiniMax
M3.1 Flash?" efter att ha sett fyra prenumerationer som alla visade "≤ ≈$0.01 per försök".

Varför de såg identiska ut: planpriset delades med antalet körningar. Google AI Pro
$19,99, Claude Pro $20, ChatGPT Plus $20 och MiniMax M Plan Go $22 delat med 1800
körningar/månad ger $0,011-$0,012, som avrundas till $0.01 för alla.

Varför det var fel: det är inte ett pris. En körning i Google AI Pro förbrukar en helt
annan kvot än en körning i MiniMax M Plan Go, så "pris per körning" från två planer
med olika kvotmekanik går inte att jämföra. Siffran säger ingenting om vad du får.

Sådan ser det ut nu:
- Planer med kvot publicerad i dollar (Copilot credits): riktigt pris per lyckad uppgift.
- Planer med mjuka gränser eller anropskvoter: bara planpris + kvot + brytpunkt mot
  API-priset. "går inte att räkna ut" i kolumnen.
- Planer vars modell inte är mätt i valt benchmark: egen sektion med pris och kvot,
  ingen prisjämförelse alls. Tidigare försvann de helt om modellen saknade
  tokenpris-estimat - nu visas de ändå.

Generell regel: en siffra som fås genom att dividera en fast kostnad med en antagande
volym är en brytpunkt, inte ett pris. Visa den som brytpunkt.

Commit 909f61b.
