---
title: Rust-API:t kör i produktion - Go-implementationer syns inte i prod
type: note
tags: [rust, migration, produktion, projekt:plan]
created: 2026-11-28
updated: 2026-11-28
status: active
related: [notes/paritetstest-gatear-rust-migreringen-mot-go-kallan]
---

Sedan PR #172 (fix(prod): agenten i produktion kör Rust-imagen igen) kör plan-api i Rust i produktion, medan utvecklingsmiljön och API-testerna går mot Go-tjänsten.

Konsekvens för nya API-funktioner: en funktion som bara implementeras i Go finns i dev men inte i prod. Det syns inte i något test - paritetstestet jämför route-tabellerna (och det fångar nya vägar), inte beteendet. Vägarna läggs in i Rust-tabellen med 501 tills de portats.

Samma bildrad gäller för agenten: environments/prod/patches/api-deployment.yaml pekar på antingen Go- eller Rust-imagen, och bot-PR:erna byter den raden. Den senaste mergen vinner - en Go-build kan avinstallera Rust-imagen i prod.

Fasmaskinen för mål (PR #179) är just nu Go-only: i dev går fasövergångarna via maskinen, i prod svarar vägarna 501. Nästa uppgift är att porta maskinen och verifieringen till Rust, eller att besluta att Rust-API:t inte ska användas för mål.

## Relaterade

- [[notes/paritetstest-gatear-rust-migreringen-mot-go-kallan]]

