---
title: Synk hallbokningar: kommunens Actor Smartbook -> matchi
type: note
tags: [stk, actorsmartbook, bokning, integration, projekt:stk]
---

Nytt projekt i stk: synka hallbokningar från Sundsvalls kommuns bokningssystem (Actor Smartbook, https://sundsvall.actorsmartbook.se) till klubbens eget bokningssystem matchi.

Första steget (2026-11-06): Go-paketet src/go/internal/actorsmartbook med endpointen GET /actorsmartbook/appointments (parametrar from/to, max 31 dagar). Klienten använder resurfiltret 628,629,630,631,632,751,633,634,635,636,637 (Baldershallen 1-5 + squashbanor 1-5).

Varför det här är viktigt att veta:
1) Kommunens API returnerar tider som /Date(1793642400000)/ där UTC-komponenterna utgor LOKAL VAGGKLOCKA (TimeZoneOffset: 0). Ett exempel: /Date(1793642400000)/ motsvarar 18:00 i Sundsvall, inte 18:00 UTC. Klienten tolkar därfor UTC-komponenterna som tid i Europe/Stockholm - annars blir alla bokningar en timme fel.
2) Kommunens JSON-parser KAVER eskapade snedstreck i datumstrangar. encoding/json:s normala output (/Date(...)/) ger HTTP 500 'There was an error processing the request'. Losningen var en egen MarshalJSON pa TelerikDate som skriver \/Date(...)\/.
3) Anropet fungerar utan session/login - ingen auth mot det kommunala systemet (verifierat 2026-11-06).

Nasta steg: mappa bokningar till matchi-reper (hall + tidsintervall + bokande organisation), och da en diff/sync-jobb.
