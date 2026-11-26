---
title: Målkompletteringar och auth i Rust - de sista API-vägarna
type: note
tags: [rust, migration, paritet, projekt:plan]
---

De sista 76 Go-vägarna är nu portade till Rust: auth-endpoints (login, callback, dev-login, session, logout) och målens kompletteringar (POST/DELETE /api/v1/goals/{id}/completions).

Fyra saker som var tvungna för att pariteten skulle hålla:

1. Go's authfel svarar {"error": ...} medan övriga API-fel svarar {"message": ...}. Rust har båda formaten nu - authlagret och felhanteringen är separata.

2. Upprepning kan sättas både vid skapande och vid uppdatering av ett mål. Serien börjar på recurrence_start_date (annars idag), och de tidigare slutförandena rensas när regeln byts - de hörde till den gamla regeln. Detta saknades helt i Rust och gav 501.

3. RRULE-felen måste skrivas med samma ordval som rrule-go ('undefined frequency: X', 'INTERVAL måste vara minst 1'), annars skiljer sig meddelandet mellan tjänsterna.

4. Varje route i tabellen måste ha en handler. Route-tabellens test kontrollerar det genom att läsa källan - annars svarar den 501. Testet fångar en endpoint som glöms bort, i stället för att den tyst ser ut att fungera.

Live-pariteten mot Go (samma databas) ger samma svar för hela flödet: skapa återkommande mål, slutföra förekomst, slutföra igen, ångra - och recurrence_next, recurrence_count och finished stämmer över.

Nästa steg: agenten - API-handlers, providers, kärna och worker.
