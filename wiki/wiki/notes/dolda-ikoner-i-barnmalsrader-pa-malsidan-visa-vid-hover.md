---
title: Dolda ikoner i målrader - visa bara det målet använder
type: note
tags: [ui, ikon, målsida, drag, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Varje målrad (barnmål på målsidan, huvudmålets ikonrad, /goals-listan) hade en handfull ikoner även om målet inte använde agent, plats, deadline, upprepning, tidsåtgång, startdatum eller "påbörjat".

Beslut: en ikon visas bara om målet faktiskt använder funktionen. Det gäller även start/paus-ikonen (circle-play) - den syns bara när målet är påbörjat, och evigt-märket bara när målet är evigt.

Teknik: en gemensam CSS-regel i src/nuxt/app/assets/css/main.css:

    .opt-icon { display: none; }
    .opt-icon.is-on,
    .opt-row:hover .opt-icon { display: inline-flex; }

Raden märks med .opt-row, ikonkapslarna med .opt-icon, och .is-on när funktionen används. display:none (inte opacity) gör att dolda ikoner inte lämnar några luckor - radens ikoner packas till höger. När raden svävas över visas de dolda ikonerna ändå, så att musvägen finns kvar. Alla tangenter (c, r, p, a, m, s, e, b) fungerar oförändrat.

Varför inte v-if: då skulle man inte kunna lägga till en plats/deadline med musen, bara via tangent - en funktionsförlust användaren inte ville.

Drag och beroende: länkikonen (orange drag-handle) togs bort helt. Istället öppnar ett släpp av ett barnmål över ett annat en modal med valet "Flytta in som undermål" eller "Skapa beroende" (blir det "Ta bort beroendet" om beroendet redan finns). Den gamla separata mouse-drag-implementationen (handleDepMouseDown/Move/Up, depDragChildId) togs bort.

Verifierat headless (playwright + chromium, lokal Nuxt-build mot Go-API på port 8099 -> Neon dev): i vila är alla oanvända ikoner display:none, använda ikoner ligger flush höger i raden (916 -> 895 med 16px padding), och en syntetisk drag+drop över ett annat barnmål gav modalen "Vad vill du göra? | Avbryt | Flytta in som undermål | Skapa beroende". Reparent och beroende verifierades mot API:t med [test]-mål som sedan raderades.
