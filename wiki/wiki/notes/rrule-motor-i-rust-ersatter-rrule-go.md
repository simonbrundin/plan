---
title: RRULE-motor i Rust ersätter rrule-go
type: note
tags: [rust, migrering, kalender, paritet, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Återkommande serier expanderades i Go med rrule-go (teambition/rrule-go v1.8.2). Rust har ingen motsvarighet, så repot fick ett eget delat crate: `src/rust/crates/plan-recurrence`.

Det stöder den delmängd av RFC 5545 som appen använder: FREQ (DAILY/WEEKLY/MONTHLY/YEARLY), INTERVAL, COUNT, UNTIL och BYDAY. Testvektorerna är desamma som i Go:s `recurrence_test.go` - samma fall, samma förväntade värden - så en avvikelse i någon port syns direkt i testerna.

**Varför inte använda ett färdigt crate?** rrule-cratet är svagt underhållt och Go-tjänstens beteende är facit i migreringen. En egen motor är dessutom delad mellan plan-api och den framtida plan-agenten, och den kan växa när återkommande mål portas.

**Tre fällor som pariteten tvingade fram:**

1. Go fyller inte i platser i svaret för ett skapat override eller en delad serie - där svarar `places: null`, medan listor och enskilda händelser svarar `places: []`. Fältet är därför `Option<Vec<Place>>` i modellen, inte `Vec<Place>`.
2. `"all_day": null` betyder false i Go (fältet har en standard). Med `Option<bool>` och en naiv bindning blev det ett NOT NULL-fel i databasen - alltså 500 i stället för 201.
3. Go:s RRULE-motor räknar i sekundprecision (`Between` dokumenterar only
