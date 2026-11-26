---
title: Dashboard importerar hela AA-katalogen (766 modeller)
type: note
tags: [dashboard, import, aa, ranking, projekt:dashboard]
---

Orsak till att bara 141 modeller importerades: isRankableModel-filtret i artificialAnalysis.ts tog bort modeller utan både Intelligence Index och tokenpris - den fria AA-endpointen returnerar hela 766 modeller (4 sidor). Fix: hela katalogen importeras; modeller utan index/priser får NULL i stället för fejkad nolla (intelligence_index och category är nu nullbara i DB och klient). Priser nulleras i normalizeAAModel så att saknat pris aldrig rankas som gratis, och estimateCostPerTask kräver båda priserna. Full sync (import.post) tvingar cache-uppdatering. Proxy-kostnads-scraping i fetch-more begränsas till modeller med intelligence_index, och stats needsPrices räknar samma villkor. Resultat efter import: 766 modeller, 155 med Intelligence Index, 467 med tokenpriser, 452 uppskattade kostnader, DeepSWE-par 14->58, AA Coding Agent-par 40, 100/766 med minst en uppmätt score+kostnad-kombination. Notera: AA-cachen på 6h gav först en inaktuell lista på 141.
