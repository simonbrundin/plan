---
title: Rust är huvudspåret och PR-previewen deployas via Flux
type: note
tags: [ci, rust, flux, preview, arm64, projekt:plan]
---

Rust är nu det som körs i dev och i produktion. Go-koden finns kvar, men
bygger inga images och deployas inte - den testas bara, så att pariteten
mot Rust går att verifiera.

Varför: två pipelines som skriver till samma manifest-rad i
environments/prod/patches/api-deployment.yaml gör att "senaste mergen
vinner" - en Go-build kunde avinstallera Rust-imagen i prod. Nu är
commit.rust-images.yaml den enda workflowen som skriver in images i
prod-manifesten.

Tre lärdomar från PR-previewen (nu fungerande):

1. Klustret kör arm64. Previewbilderna byggdes amd64-only, vilket ger
   "no match for platform in manifest" vid pull. Nu byggs varje arkitektur
   på en runner av samma arkitektur (arm64 under QEMU tar timmar för en
   Rust-kompilering) och slås ihop till en manifestlista. Samma mönster
   gäller commit.rust-images.yaml.

2. Preview-overlays publiceras på branchen "previews", som Flux läser via
   GitRepository "plan-previews" och Kustomization "plan-preview"
   (path: environments/pr, prune: true). GitRepository "plan" läser bara
   main, och en PR-branch följs inte av en källkontroll.

3. Overlayen måste vara självständig. Den publiceras på en branch som bara
   innehåller overlays, så referenser till environments/base finns inte
   där - schemat kopieras in av skriptet i stället. En tom
   environments/pr gör dessutom att kustomize bygger "no resources" och
   Flux inte prunar något, så en README håller katalogen ifylld.

Buggen som rökade igång allt: Rust-API:t startade i mock-läge om
Postgres inte svarade vid start, och stannade där. I ett nytt kluster
(CNPG kör initdb, Atlas applicerar schemat efteråt) blev API:t ligga kvar
i mock-läge och svara "Failed to create user" trots frisk databas. Nu
väntar API:t på databasen (DB_CONNECT_TIMEOUT_SECONDS, default 120s) och
avslutar sedan, så att podden startar om.

Regler som gäller framöver:

- Ingen bild får byggas med en enda arkitektur. Klustret är arm64.
- Ingen workflow får skriva till prod-manifesten utom commit.rust-images.yaml.
- Deploy sker alltid via GitOps - pipeline committar, Flux applicerar.
