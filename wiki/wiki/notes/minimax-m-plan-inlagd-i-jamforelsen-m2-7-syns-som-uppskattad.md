---
title: MiniMax M Plan inlagd i jamforelsen, M2.7 syns som uppskattad
type: note
tags: [dashboard, minimax, subscriptions, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Användaren kör MiniMax M2.7 via 'MiniMax Starter Plan'. Faktakoll: (1) minimax-m2-7 finns i AA-katalogen (tokenpriser 0.3/1.2 $\/M, uppskattad agentkostnad 8.64 $\/task via max-profil) men AA har INTE mätt Coding Agent Index for den (liksom for hela M2.x-familjen - bara M3/M3.1-Flash har Intelligence Index). (2) MiniMax:s prenumerationsprodukt heter M Plan med nivåerna Go $22/mån, Explore $55, Build $132 (verifierat 2026-11-26 mot platform.minimax.io/docs/m-plan/monthly-offer; lanseringserbjudandet 50% forsta mån upphorde 2026-10-14). 'Starter' är den gamla Token Plan-nivåns namn - exakt pris for den kunde inte verifieras (sidor bot-blockerade/JS). (3) VIKTIGT: M Plan är KVOT-baserad - 'usage deducts from the included M Plan quota according to its pay-as-you-go price' med 5h- och veckofönster - sa ett 'obegränsat pris per uppgift' kan inte härledas ärligt. Implementation: SUBSCRIPTION_PLANS fick nya fältet quotaBased (Go/Explore/Build inlagda, priceVerified); compareSubscriptionPlans ger quotaBased-planer null i break-even/eff-kolumner (UI visar 'kvot'); buildRecommendation hoppar över quotaBased-planer i 'Svar' men API-optionslistan inkluderar nu ossakerade modeller märkta apiCostIsEstimated/qualityKnown=false (≈ + 'kodnivå ej mätt av AA') - de kan aldrig vinna rekommendationen. Test uppdaterad: ossakerad modell syns men vinner inte (40/40). Om användarens faktiska Starter-pris skiljer sig - prisfilen är ren data (app/data/subscriptions.ts).
