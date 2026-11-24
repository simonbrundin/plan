---
title: Lokal expansion kan verifieras mot serverns expansion
type: note
tags: [kalender, test, rrule, projekt:plan]
---

När kalendern expanderar serier lokalt (se notes/kalendern-hamtar-alla-handelser-en-gang-och-expanderar-serier-lokalt) är det viktigt att den lokala expansionen ger samma svar som serverns rrule-go-expansion.

Metod: hämta hela datasetet med scope=all, kör den lokala expandCalendarEvents() för samma from/to som ett serveranrop, och jämför de två listorna på titel+starttid. En skillnad på millisekunder är kosmetisk: servern formaterar förekomster med RFC3339 (sekundprecision) medan klienten behåller millisekunder. Ordinal, flyttade och avstängda förekomster måste matcha.

Det här är viktigt eftersom klienten och servern använder olika bibliotek (rrule och rrule-go) men samma datumalgoritm - en avvikelse i ett av dem ger olika kalendrar beroende på om klienten eller servern svarar.
