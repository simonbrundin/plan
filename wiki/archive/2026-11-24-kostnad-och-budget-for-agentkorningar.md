---
title: Kostnad och budget för agentkörningar
type: note
tags: [kostnad, agent, budget, lesson, projekt:plan]
---

Plan äger kostnaden, agenten rapporterar bara användningen.

**Varför:** samma prislista ska gälla för alla körningar, och kostnaden ska kunna reproduceras senare. Därför versioneras prislistan (2026-11-24.1) och varje usage-rad sparar version, valuta och om priset kom från listan eller från en schablon. En okänd modell får aldrig kosta noll - noll vore en falsk kostnad.

**Lärdomar:**
- MiniMax strömmar utan användning ("usage": null i varje chunk). Att läsa fler chunks efter finish_reason räcker inte - man måste använda det vanliga svaret. Lösningen blev StreamingOptions.Usage: när användningen efterfrågas används ett icke-strömmat svar. (Att bara sätta stream=false i requesten räcker INTE - då kommer SSE tillbaka och JSON-parsningen fallerar med 'invalid character d'.)
- Kostnadskontroll och reservation måste ske i samma transaktion som att körningen skapas, annars kan två samtidiga körningar tillsammans överskrida gränsen. Budgetraderna låses med FOR UPDATE.
- Den väntande körningens egen reservation måste räknas med i kontrollen, annars kan N körningar tillsammans överskrida gränsen.
- En budget som nått sin gräns måste kunpausas, annars kan agenten inte köra igen förrän perioden går ut. Därför finns PATCH på agent-budgets.
- Terminala körningar får alltid en usage-rad; markerad incomplete om agenten inte hann rapportera.

**Verifierat i dev:** körning #81 (MiniMax-M2.7) gav 36737 input- och 1728 output-tokens, 4 verktygsanrop (run_test, write_file) och 0.0080386 USD med price_source=price_list. Hård budget ger 409 reason=limit_reached, varning ger budget_warning utan stopp, okänd modell ger 409 reason=unknown_price, och två samtidiga körningar mot samma budget ger en 201 och en 409.
