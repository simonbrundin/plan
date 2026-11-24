---
title: Prioriteringsordningen på /now sparades inte - bulk-update och spar-kö
type: note
tags: [bugg, go, nuxt, prioritetssida, prestanda, projekt:plan]
created: 2026-11-23
updated: 2026-11-23
status: active
related: [notes/sortering-av-grundmal-pa-goals-med-piltangenterna]
---

Symtom: att ändra ordningen på /now med ArrowUp/ArrowDown såg ut att fungera,
men efter en sidladdning var ordningen som den var innan.

Rotorsak 1 (dominerande): ReorderPriorityGoals i Go-repot skrev en UPDATE per
mål i en loop - en round-trip till databasen per mål. Dev-databasen ligger på
Neon över internet, och en användare har ~250 öppna mål, så varje omsortering
 tog 8-10 sekunder. En sidladdning (eller en tillfällig förlust av
anslutningen) avbröt requesten innan den kom fram, och det som såg ut som
"icke sparad sortering" var bara ett sparande som aldrig hann landa.

Fix: setPriorityOrder() i goal_repository.go gör EN bulk-update via
UNNEST($1::bigint[], $2::bigint[]). 8,5 s -> 0,33 s lokalt, ~0,4 s i dev.

Rotorsak 2: usePriorityMode skickade en request per tangenttryck, serialiserade
i en kedja. Med 10 sekunders svarstid hann en kö växa, och den sista ordningen
som sparades kunde vara en gammal tangenttryckning.

Fix: pendingOrder på modulnivå - bara den SENASTE ordningen sparas, och den
skickas så snart ett pågående sparande är klart. pagehide-handler med
keepalive sparar den sista ordningen om sidan laddas om direkt.

Test: TestReorderPriorityGoalsIntegration i goal_reorder_integration_test.go -
egen testanvändare i databasen, omsortering, läser listan igen och kontrollerar
att ordningen ligger kvar. Testet städar bort sina egna rader.

Lärdom: skriv inte en UPDATE per rad när databasen ligger över internet.
Bulk-updates (UNNEST) är skillnaden mellan 10 sekunder och 0,3 sekunder, och
den skillnaden syns direkt i gränssnittet som "sparat" respektive "inte sparat".

## Relaterade

- [[notes/sortering-av-grundmal-pa-goals-med-piltangenterna]]

