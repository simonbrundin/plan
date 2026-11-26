---
title: Dashboard: en synk-knapp i stället för tre förvirrande
type: note
tags: [dashboard, ux, sync, projekt:dashboard]
---

Användaren upplevde de tre knapparna (Synka modeller/Hämta priser/Uppdatera priser) som otydliga. Fix: fetch-more (SSE-endpointen som redan visade progress-bar) är nu den enda syncen - den tvingar AA-cache-uppdatering (forceRefresh), upsertar hela katalogen, tar bort modeller som AA lagt ner (DELETE WHERE NOT slug = ANY), och scrapar proxy-kostnad för modeller med Intelligence Index som saknar task-kostnad. Klart-meddelandet summerar: 'Synkat klart: N modeller, X nya, Y borttagna, Z priser hämtade'. UI har nu en knapp 'Synka modeller (totalInDb)'; useModelActions förlorade importModels/refreshPrices (endpoints /api/models/import och /api/models/refresh-prices finns kvar server-side som underhållsverktyg). Verifierad körning: 766 modeller, 0 nya, 0 borttagna; scraping-fasen gav ingen träff för de 15 - de är gratis eller saknar publicerad task-kostnad och får fortfarande null.
