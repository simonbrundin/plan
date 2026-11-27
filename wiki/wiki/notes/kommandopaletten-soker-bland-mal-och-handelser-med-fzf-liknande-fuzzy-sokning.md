---
title: Kommandopaletten söker bland mål och händelser med fzf-liknande fuzzy-sökning
type: note
tags: [feature, nuxt, sök, kalender, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
---

2026-11-27: Ctrl+K (Cmd+K) öppnade en palett som bara sökte bland mål och navigeringslänkar. Nu söker den bland mål, kalenderhändelser och navigering, och sökningen är fzf-liknande.

Varför egen fuzzy-matchning (app/utils/fuzzy.ts) i stället för Fuse.js: Fuse med höjd tröskel matchar för lätt ("mk" träffar "Läkare"), och Fuse med låg tröskel hittar inte "mt" i "Mötet i rummet". Egen matchning ger prefix-, ordgräns- och camelCase-bonus, hoppstraff och smart case, och matchar "mål" mot "mal" (diakriter plockas bort). Enhetstestad med bun test (test/unit/fuzzy.test.ts).

Varför paletten filtrerar själv: UCommandPalette filtrerar med Fuse och ignorerar grupper med `ignoreFilter`. Därför rankas träffarna i useCommandPalette och skickas in redan filtrerade - annars skulle palettens egen filtrering krocka med fuzzy-träffarna.

Varför kalenderhändelser hämtas via useCalendarSearch: paletten och kalendern delar ett useState-cachat eventdatasät, så sökningen inte laddar data i onödan men ändå hittar händelser direkt.

Varför fokus sker via useCalendarFocus: kalendern ligger i en egen komponent (och i layoutens högerspalt), så paletten lämnar över händelsen i en delad state. Kalendern zoomar dit och öppnar händelsen för redigering. Varje begäran har ett id, så att två kalendrar (layout + /calendar på mobil) inte öppnar samma händelse två gånger.

Begränsning: återkommande serier söks som en träff (masterraden), inte varje förekomst - annars skulle sökningen svämma över.

Verifiering: bun run build, bun test (251 tester), cucumber (89 scenarier), typecheck-gate grön, samt headless chromium mot den byggda appen (dev-login-session): Ctrl+K visar mål/händelser/navigering, "sömn" hittar Sömn-händelserna, val av händelse öppnar "Redigera händelse" i kalendern, val av mål navigerar till /goal/<id>, och söktexten nollställs när paletten öppnas igen.
