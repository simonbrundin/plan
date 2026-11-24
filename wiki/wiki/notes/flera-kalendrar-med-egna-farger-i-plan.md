---
title: Flera kalendrar med egna färger i Plan
type: note
tags: [kalender, feature, färg, projekt:plan]
created: 2026-11-17
updated: 2026-11-18
status: active
related: [notes/handelsen-har-ingen-egen-farg, notes/standardkalender-i-plan, notes/valfri-fargkod-pa-kalendrar, notes/vertikal-kalender-med-ikonstod-i-plan]
---

Användaren kan nu skapa hur många kalendrar som helst (t.ex. Jobb, Fritid),
var och en med eget namn och egen färg. Händelser läggs i en kalender och
har kalenderns färg. (Händelser kunde ha en egen färg - den möjligheten togs
bort senare, se notes/handelsen-har-ingen-egen-farg.)

## Varför så här

- **Färgen bor i kalendern, inte bara i händelsen.** En händelse har alltid
  en `calendar_id`, och klienten löser upp färgen så: kalenderns färg ->
  standardfärgen #6366f1. Logiken ligger i
  src/nuxt/app/utils/calendarColors.ts så att kort, prickar och formulär
  använder samma regel.
- **Kalendern har ingen egen vy.** En vy per kalender hade blivit ett
  större projekt. I stället visas alla kalendrar i samma tidslinje, och en
  legendrad över tidslinjen togglar synlighet per kalender.
- **`calendar_id: 0` betyder "ingen kalender".** PATCH har inte samma
  tre-state-semantik för null som för ikon/färg, så 0 används för att koppla
  bort en händelse från sin kalender.
- **ON DELETE SET NULL på kopplingen.** När en kalender tas bort försvinner
  inte händelserna - de blir kalenderlösa. Det är mindre destruktivt än
  att radera händelser.

## Gotchas som kostade tid

- **Snabbfältet var den enda fungerande vägen att skapa händelser.**
  Händelseformuläret i VerticalCalendar.vue var kodat men aldrig öppnat
  (openCreateEvent saknade anropare). Utan kalenderväljare i snabbfältet hade
  det inte gått att placera händelser i en kalender alls. Formuläret är nu
  nåbart via "+ Händelse" i legendraden.
- **Node 26 tappar alla response-headrar i Nitro:s devserver.** Content-Type,
  Set-Cookie och Location försvinner, så SPA:n laddar inga scripts och
  inloggningen via dev-login tappar sessioncookien. Med Node 22 fungerar det
  som vanligt. Kör devservern med Node 22 (node:lts-alpine i produktion).
- **Fältet för kalendernamn visade en tom sträng** i hanteraren, eftersom
  v-modellen var bunden till en draft som aldrig fylldes från kalenderns
  nuvarande namn.
- **`calendar_event_repository_integration_test.go` raderar alla användarens
  händelser** i sin defer (`DELETE FROM calendar_events WHERE user_id=$1`).
  Kör den aldrig mot dev-databasen - den raderar användarens riktiga data.

## Verification

Verifierat i riktig browser (Playwright + systemets chromium) mot dev-API:t
och Neon: skapa kalender, ärva färg, egen färg, "Följ kalendern", dölja och
visa kalender, flytta händelse mellan kalendrar, ta bort kalender samt
snabbfältets kalenderväljare. Inga konsolfel.

## Relaterade

- [[notes/valfri-fargkod-pa-kalendrar]]
- [[notes/handelsen-har-ingen-egen-farg]]
- [[notes/standardkalender-i-plan]]
- [[notes/vertikal-kalender-med-ikonstod-i-plan]]
