---
title: Kameo-sidan flyttad till toppnivå och delad i flikar
type: note
tags: [pappa, nuxt, kameo, ui, projekt:pappa]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Kameo-utbetalningar flyttades från /mail/kameo-utbetalningar till /kameo-utbetalningar eftersom sidan inte är en del av mailagenten - den läser Kameos lån direkt via /api/kameo/loans. Sidan har nu två flikar: Aktiva lån (status ongoing) och Återbetalade lån (status repaid). Fliken byter automatiskt om ett sökträffar ett lån på den andra fliken. Dev-klustret deployas via Tilt (environments/dev/Tiltfile), så ändringar syns efter automatisk ombyggnad.
