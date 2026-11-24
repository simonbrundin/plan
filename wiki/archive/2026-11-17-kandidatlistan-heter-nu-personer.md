---
title: Kandidatlistan heter nu /personer
type: note
tags: [pappa, ui, navigering, omdöping, projekt:pappa]
---

Sidan som tidigare hette /kandidater heter nu /personer (2026-11-17).

VARFÖR: registret innehåller både rekryteringskandidater och personer som redan anstälts eller som inte är aktuella just nu. 'Kandidat' var alltså ett fel laddord för sidan - fliken 'Ej aktuella nu' visar ju just personer som INTE är kandidater.

GENOMFÖRANDET: sidan heter nu /personer, men gamla länkar fungerar - src/nuxt/app/pages/kandidater.vue är en redirect-sida som behåller query-parametern (?person=<id>), så bokmärkta djuplänkar till detaljvyn går att öppna. Menyn, startsidan (/ redirectar till /personer), /admin, /mail/rules och /relationer pekar om.

OBS: internt i koden heter fortfarande PersonCard, CandidateDetail och 'kandidat' i kommentarer - det är fritt om manuella kandidater och rekryteringsspråket lever vidare. Bara sidans namn och URL ändrades.
