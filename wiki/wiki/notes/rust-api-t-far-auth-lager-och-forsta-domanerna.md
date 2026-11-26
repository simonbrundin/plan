---
title: Rust-API:t får auth-lager och första domänerna
type: note
tags: [rust, migration, auth, paritet, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Plan-migreringen från Go till Rust har tagit ett steg till: auth-lagret och
de första domänerna (statusar, mål, taggar) är nu portade, och en
live-paritetsskript jämför de två tjänsterna.

**Varför auth-lagret kom först:** agentens API-nyckel ger tillgång till
Plan, och gränsen mellan "agenten får arbeta med mål" och "agenten får
radera och besluta" måste ligga i API:t - inte i en prompt. Därför är
agent-scopets allowlist en port av Go:s `agent_scope.go`, med samma
undantag: agenten får aldrig radera och får aldrig besluta om ett
godkännande.

**Sessionstoken är byte-kompatibla med Go.** Under migreringen kör båda
tjänsterna parallellt mot samma databas, så en cookie som skapats av den
ena måste fungera i den andra. Det uppnås genom att payloaden serialiseras
i alfabetisk ordning (Go marshalerar en map) och tidsstämplar skrivs utan
sekunder ("2006-01-02T15:04:05Z").

**Live-paritet istället för bara enhetstester.** `scripts/compare-api-parity.sh`
anropar samma vägar i båda tjänsterna och jämför svaren efter att
tidsstämplar normaliserats till epok. Det fångar skillnader som
enhetstester inte gör - till exempel att Go svarar `null` för platser och
taggar i ett uppdaterat mål, men `[]` i en lista.

Kvar att porta: prioriteringslistan, relationer, beroenden, återkommande
mål, platser, kalendrar, kalenderhändelser, agentdomänerna och
API-nycklarna.
