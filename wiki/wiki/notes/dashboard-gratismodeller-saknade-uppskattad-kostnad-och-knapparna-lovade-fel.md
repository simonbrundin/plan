---
title: Dashboard: gratismodeller saknade uppskattad kostnad och knapparna lovade fel
type: note
tags: [dashboard, bug, estimated-cost, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Efter katalogimporten (766 modeller) upptaktes tva UX-buggar: (1) estimateCostPerTask returnerade null nar kostnaden blev 0, sa 15 gratismodeller (AA listar 0/0-priser, t ex gemma) fick ingen uppskattad kostnad trots att metodiken sager att noll kostnad ar gratis. Fix: estimate >= 0 ar giltig; saknade priser ger fortfarande null. (2) 'Hamta priser (N utan)' visade modeller utan kodningskostnad (~726) men scraping kan bara fylla proxy-kostnad for modeller med Intelligence Index (15) - knappens räknare matchade inte vad knappen gor. Fix: nya client-raknare modelsWith/WithoutProxyCost (aaIntelligenceCostPerTask/costPerTask + intelligenceIndex-villkor) styr knapparna, och ModelsHeader visar nu 'matt task-kostnad / uppskattad kostnad / utan priser' i stallet for det tvetydiga 'utan kostnad'. Datamangd efter fix: 467 uppskattade kostnader (467 med priser, 15 gratis->0), 140 matt proxy-kostnad, 299 modeller utan nagon prisdata alls (AA publicerar ingen pricing for dem - arligt tomt, inte en bugg).
