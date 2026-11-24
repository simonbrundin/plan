---
title: Standardkalender i Plan
type: note
tags: [kalender, feature, standardkalender, projekt:plan]
created: 2026-11-18
updated: 2026-11-18
status: active
related: [notes/flera-kalendrar-med-egna-farger-i-plan]
---

Användaren kan nu sätta en av sina kalendrar som standardkalender. Nya
händelser hamnar då där utan att kalender behöver väljas varje gång.

## Design

- **En standardkalender per användare, garanterat av databasen.**
  `calendars.is_default` plus en unik index på `(user_id) WHERE
  is_default`. Servern avmarkerar dessutom den gamla standardkalendern i
  samma transaktion som den nya sätts, så två kan aldrig
  vara standardkalendrar samtidigt, inte ens vid race.
- **Standardkalendern tillämpas på servern, inte bara i klienten.** Ett
  `POST /calendar-events` utan `calendar_id` hamnar i standardkalendern.
  `calendar_id: 0` betyder däremot uttryckligen "ingen kalender" och
  skapar ingen kalenderkoppling - annars skulle det inte gå att skapa
  en kalenderlös händelse när en standardkalender finns.
- **Klienten preselectar.** Formuläret och snabbfältet väljer
  standardkalendern när de öppnas, så att det räcker att skriva.
- **Den första kalendern blir standard automatiskt**, så att det finns
  någotstans att landa direkt från början.

## Gotchas

- **Tre-state-semantik även här.** `PATCH` med `is_default: true` sätter,
  `false` avmarkerar och utelämnat lämnar oförändrat.
- **stjärnan i hanteraren växlar mellan "sätt" och "ta bort"** utifrån
  kalenderns aktuella `is_default`. Det gör att man måste hålla koll på
  vilken knapp som är vilken - därför bytades aria-labbet med ("Sätt X som
  standardkalender" / "Ta bort X som standardkalender") så att knappen
  går att nå via tangentbord och skärmläsare.
- **OpenAPI-specen måste genereras med `swag init -g cmd/server/main.go -o
  docs --parseDependency`.** Utan `--parseDependency` byter swag namn på
  alla definitioner (`plan-api_internal_models.X` -> `models.X`) och diffen
  blir orelaterad churn.

## Verification

Verifierat i browser (Playwright) mot dev-API:t och Neon: stjärnan sätter
och avmarkerar standardkalendern, legenden visar stjärnan, formuläret och
snabbfältet preselectar kalendern, och händelser skapade från båda vägarna
hamnar i standardkalendern och ärver dess färg. Go-integrationstest och
Nuxt-enhetstester täcker samma kedja.

## Relaterade

- [[notes/flera-kalendrar-med-egna-farger-i-plan]]

