---
title: Taggar på mål i plan - tangenten g
type: note
tags: [feature, goals, tags, ui, go, nuxt, db, projekt:plan]
created: 2026-11-23
updated: 2026-11-23
status: active
---

ÖNSKEMÅL: kunna lägga till taggar på mål (t.ex. "ute") och kunna göra det med ett enkelt kortkommando.

LÖSNING:
- DB: nya tabeller "tags" (user_id, name) och "goal_tags" (goal_id, tag_id). Namnen normaliseras (gemener, "#" tas bort), så "Ute" och "ute" blir samma tagg. Taggar hör till användaren - en unik index på (user_id, name). Schemat ligger i environments/base/db/schema.yaml (kanonisk fil, Atlas applicerar den).
- API: GET /api/v1/tags (användarens taggar + antal mål), GET/PUT /api/v1/goals/:id/tags (sätter målets taggar utifrån namn - nya namn skapas), DELETE /api/v1/goals/:id/tags/:name. Målen i listan, målsidan och målgraphen bär nu med fältet "tags".
- Gränssnitt: tangenten "g" på målsidan öppnar taggväljaren (GoalTagPicker) för det markerade målet - huvudmål, undermål eller förälder, samma mönster som tangenten "p" för platser. Man skriver ett taggnamn och trycker Enter: finns taggen läggs den på (tryck igen för att ta bort), annars skapas den. Tomt fält + Enter gör ingenting - en oavsiktlig tangent ska inte ändra målet. Taggar visas som text bredvid taggikonen i målets rubrikrad och i undermålsraderna.

VARFÖR NAMN ISTÄLLET FÖR ID: tangentflödet ska vara "skriv - Enter". Med id skulle klienten först behöva slå upp taggen, och varje nytt namn skulle kräva ett extra anrop. Namn gör att ett anrop räcker, och normaliseringen sköter dubbletter.

FÄLLA SOM UPPTÄCKTES: isSelected() i väljaren jämförde strings som om de vore objekt (tag.name.toLowerCase()), vilket kastade TypeError när målet redan hade en tagg - modalen försvann då. Selecterade taggar är strings, inte Tag-objekt.

VERIFIERING: go test ./... och bun test test/unit (216 tester) gröna, bun run build går igenom. Headless Chromium mot lokal Nuxt-build + Go-API mot Neon dev: 'g' öppnade väljaren med fokus i fältet, "ute" + Enter skapade taggen (API svarade ["ute"] och taggen syntes i sidan), Enter igen tog bort den, Escape stängde. API:t testat separat: normalisering ("Ute", " hem ", dubbletter), skapa/uppdatera/lista/graph, 403 för en annan användares mål. Testdata ([test] taggtest + tagtest-användare) raderades ur dev-databasen efteråt.
