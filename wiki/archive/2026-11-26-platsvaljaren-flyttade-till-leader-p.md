---
title: Platsväljaren flyttade till leader + p
type: note
tags: [ui, tangentbord, platser, session, projekt:plan]
---

Kortkommandot för att sätta plats på ett mål är nu leader + p (Space, p) och fungerar på alla sidor där ett mål är markerat: /now, /goals och målsidan (/goal/:id).

Varför: 'p' var reserverat för platsväljaren på målsidan, vilket gjorde att 'p' inte kunde växla till/från Gör nu på den sidan. Nu är 'p' globalt sidbyte och platsväljaren ligger i leader-läget, som redan används för taggar (leader + t), datum (leader + d) och agentmärke (leader + a).

Hur det är löst: ny global composable useGoalPlacePicker (state utanför funktionen, som useGoalTagPicker) och ny komponent GoalPlacePicker som renderas i layouts/default.vue. Varje sida registrerar ett leader-kommando 'p' via registerPageCommands och pekar på sitt markerade mål. Efter sparande skrivs platserna tillbaka i sidans eget urval via placePicker.lastSaved/savedRevision.

Följd: isDoableAtPlaceOrUnplaced i app/utils/places.ts - mål utan någon plats är nu också 'görbara' och visas alltid i filtret 'Görbar här' på /now, eftersom de inte är bundna till en plats.
