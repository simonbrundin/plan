---
title: Dashboard: tydlighetslyft av jamforelserna
type: note
tags: [dashboard, ux, projekt:dashboard]
---

Användaren upplevde jamforeelserna som 'sjukt otydliga'. Tydlighetslyft: (1) formatModelName i modelFormatters kortar namn: '(Max, Default Fallback)' -> ' - Max' sa att tabeller visar produktnamn + qualifier istallet for gror. (2) ModelsBudgetFinder omdesignad: stor 'Svar:'-rad med det direkta svaret (billigast som klarar referensnivån, med %-besparing), sjalvdokumenterande kolumnrubriker ('Kodpoäng (AA Coding Agent Index, 0-100)', 'Kostnad per agentuppgift (mätt av AA)'), separata gruppen för token-estimat med tydlig 'inte en mätning'-text, max 6 rader + visa-alla-knapp. (3) ModelsSubscriptionComparison omdesignad: 'Svar:'-rad vid vald användningsnivå, svenska kolumner ('Lönar sig redan vid', 'Kostnad per uppgift med planen'), plan-notes som tooltip i stället för celltext. (4) Språk-sweep: statskort, jämförelsetabellens rubriker (Modell/Kategori/Hastighet/Öppen) och status-texter till svenska. 33/33 tester, lint, typecheck rent utöver talos, build, Tilt Ready.
