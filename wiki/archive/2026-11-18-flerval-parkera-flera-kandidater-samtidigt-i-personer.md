---
title: Flerval: parkera flera kandidater samtidigt i /personer
type: note
tags: [feature, pappa, personer, potential, projekt:pappa]
---

Ny funktion i pappa: på /personer kan man markera flera personkort och parkera hela gänget med samma skäl och samma genomgångsdatum.

Varför: enskild parkning var teduggt - varje kandidat krävde egen dialog och eget datum, och det var lätt att glömma parkera en i gruppen. Flerval gör det till en satsning.

Hur det fungerar: rutan i varje PersonCard markerar kandidaten (selectable/selected-props + emit toggle-select). 'Markera alla synliga' markerar aktuell flik, och byte av flik nollställer urvalet så att man inte kan parkera personer man inte ser. Ett flytande fält längst ner visar antalet markerade och knappen '⏸ Markera ej aktuella' + 'Rensa val'. Dialogen är samma som för en enskild kandidat (skäl, valfri not, datum).

Ny endpoint: PATCH /api/potential-bulk (kräver adminsession) - uppdaterar alla valda i EN transaktion och loggar i audit_log med bulk: true, en rad per person. Delade valideringen ligger i shared/utils/potential.ts (parseBulkSelection, isRealIsoDate) så att webbläsaren och servern använder samma regler; max 500 personer per satsning.

Verifierat: bun test (147 grön), nuxt typecheck (bara fel i parallellt arbete om marknad), samt end-to-end mot dev-klustret via x-admin-token - parkade en person, kontrollerade is_potential/skal/datum och aktiverade tillbaka den så databasen är oförändrad.
