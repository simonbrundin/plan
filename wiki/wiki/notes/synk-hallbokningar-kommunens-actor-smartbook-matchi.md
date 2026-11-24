---
title: Synk hallbokningar: kommunens Actor Smartbook -> matchi
type: note
tags: [stk, actorsmartbook, bokning, integration, projekt:stk]
created: 2026-11-06
updated: 2026-11-17
status: active
related: [notes/gron-markering-av-tider-som-ar-bokade-i-matchi-skrivlasande-jamforelse, notes/krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force, notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas, notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass, notes/query-parametrar-i-actor-smartbook-endpointen-url-avkodas-till-mellanslag, notes/stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts]
---

Nytt projekt i stk: synka hallbokningar från Sundsvalls kommuns bokningssystem (Actor Smartbook, https://sundsvall.actorsmartbook.se) till klubbens eget bokningssystem matchi.

Första steget (2026-11-06): Go-paketet src/go/internal/actorsmartbook med endpointen GET /actorsmartbook/appointments (parametrar from/to, max 31 dagar). Klienten använder resurfiltret 628,629,630,631,632,751,633,634,635,636,637 (Baldershallen 1-5 + squashbanor 1-5).

Varför det här är viktigt att veta:
1) Kommunens API returnerar tider som /Date(1793642400000)/ där UTC-komponenterna utgor LOKAL VAGGKLOCKA (TimeZoneOffset: 0). Ett exempel: /Date(1793642400000)/ motsvarar 18:00 i Sundsvall, inte 18:00 UTC. Klienten tolkar därfor UTC-komponenterna som tid i Europe/Stockholm - annars blir alla bokningar en timme fel.
2) Kommunens JSON-parser KAVER eskapade snedstreck i datumstrangar. encoding/json:s normala output (/Date(...)/) ger HTTP 500 'There was an error processing the request'. Losningen var en egen MarshalJSON pa TelerikDate som skriver \/Date(...)\/.
3) Anropet fungerar utan session/login - ingen auth mot det kommunala systemet (verifierat 2026-11-06).

Uppdatering 2026-11-06 (senare pa dagen): mappingen och sync-jobbet ar KLARA. Kopplingen mellan kommunresurs och MATCHi-bana ligger i tabellen court_mappings (kommunresurs 628-637, 751 -> MATCHi-banor for idrottshall/squash/coworking). Synkroniseringen ligger i src/go/internal/sync och kors av det kommunala systemet (Actor Smartbook) mot MATCHi varje dag kl 03:00 via cronjobbet matchi-sync.

Viktigast att veta om sync-jobbet:
- Det speglar 21 dagar framover, avrundar start nedat och slut uppat till hela/halva timmar.
- Det ar sjavaterstallande: en bokning som flyttats manuellt i MATCHi flyttas tillbaka.
- Identifiering sker via kommun_appointment_id + matchi_court_id, inte via MATCHi slot-id.
- Orphan-avbokning ar AVSTANGD som standard (SYNC_REMOVE_ORPHANS=false), eftersom en bokning kan doldas i det kommunala systemet.

## Relaterade

- [[notes/query-parametrar-i-actor-smartbook-endpointen-url-avkodas-till-mellanslag]]
- [[notes/gron-markering-av-tider-som-ar-bokade-i-matchi-skrivlasande-jamforelse]]
- [[notes/krypterade-secrets-i-klustret-ogiltiga-inloggningar-och-keycloak-brute-force]]
- [[notes/stk-matchi-synk-sjalvaterstallande-synk-och-var-database-url-satts]]
- [[notes/matchi-bokning-slot-id-andras-nar-bokningen-skapas]]
- [[notes/matchi-keycloak-session-code-maste-fangas-med-generell-teckenklass]]
