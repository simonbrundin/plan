---
title: NeuroQuant-vikterna räknas om dagligen - mejl triggas på publicering
type: note
tags: [neuroquant, bug, session, projekt:simon]
---

Bakgrund: 2026-11-03 kördes monitorn på pxe och skickade ett mejl 12:12 trots att portföljen inte hade publicerats om. Diffen visade Boliden 3,1 % -> 2,9 %, Boozt 3,4 -> 3,1, Investor B 7,5 -> 7,2, Proact 3,7 -> 3,5.

Varför: API:t (app.neuroquant.ai/api/trpc, strategy.getPortfolioData) returnerar fältet 'weight' = justerad vikt (InvVol 100 %, summerar till exakt 100). Den räknas om löpande när volatilitetsestimaten uppdateras, alltså utan att modellen publicerat om. Fältet 'target_weight' (summerar till ~96) är modellens råa målvikt och ändras bara vid publicering. updatedAt (t.ex. 2026-10-29T12:00:00.000Z) är publiceringsstämpeln.

Beslut: monitorn triggar nu på ny publicering (updatedAt/datum ändrat), inte på viktändringar ensamma. Mellan publiceringar jämförs vikterna och MIN_CHANGE_THRESHOLD (0,2) väljer vilka positioner som listas. Tröskeln är strikt (> 0,2) och flyttalsbrus rundas till 6 decimaler, så 3,1 - 2,9 räknas som exakt 0,2 och filtreras bort. Nya/borttagna positioner filtreras på samma sätt. state.json uppdateras bara vid ny publicering, så ändringar mäts mellan publiceringar.

Följd: ett mejl per publicering som faktiskt ändrats (var 4-5:e vecka), aldrig 20-70 mejl. Mottagare: simonbrundin@gmail.com och michael@brundins.se (bekräftat via Skickat-mappen).
