---
title: plan-db: CNPG-kluster fastnade i 'Setting up primary' på en Longhorn-volym utan replikor
type: note
tags: [cnpg, longhorn, incident, plan, projekt:simon]
created: 2026-11-26
updated: 2026-11-27
status: active
related: [concepts/cloudnative-pg, notes/cnpg-initdb-agarollen-skapas-fore-postinitsql, notes/cnpg-managed-roles-secretet-saknar-losenord-om-username-avviker-fran-rollnamnet, notes/sops-kryptera-bara-stringdata-i-kubernetes-secrets-ellers-kraschar-flux-builden]
---

Plan-appens databas (CNPG-kluster plan-db i namespace plan) fastnade i 'Setting up primary' i 20 timmar. Symptom: Go API:t svarade 500 'Failed to create user session' med 'dial tcp 10.108.196.67:5432: connect: connection refused', och initdb-jobbet plan-db-1-initdb stod i Failed.

Symptomet kunde se ut som ett Longhorn-problem, men rotorsaken var initdb: Longhorn-volymen pvc-91f9c951 saknade helt replikor (replicas: {}, robustness unknown) och var detached - den kunde inte monteras. Dessutom vägrade CNPG skapa primärinstansen ('refusing to create the primary instance while the latest generated serial is not zero') eftersom serien på klusterresursen inte var noll. Men även efter att volymen ersatts av en frisk misslyckades initdb, av en annan orsak: `postInitSQL` refererade till rollen `plan` innan den existerade ('role "plan" does not exist'), se [[notes/cnpg-initdb-agarollen-skapas-fore-postinitsql]]. Och när initdb väl gick igen saknade rollen lösenord på grund av ett sekret med fel `username`, se [[notes/cnpg-managed-roles-secretet-saknar-losenord-om-username-avviker-fran-rollnamnet]].

Fix: radera det trasiga initdb-jobbet, radera den tomma PVC:n, radera CNPG-klustret (data fanns aldrig - initdb hade aldrig lyckats) och låt Flux återskapa det. Den nya volymen provisionerades på worker-9 och blev attached/healthy. Därefter sattes `bootstrap.initdb.owner: plan`, sekretets `username` ändrades till `plan` och det krypterade sekretet fick `encrypted_regex: ^(data|stringData)$` - annars panikar Flux-builden, se [[notes/sops-kryptera-bara-stringdata-i-kubernetes-secrets-ellers-kraschar-flux-builden]].

Slutläge (2026-11-27): `plan-db` är healthy med tre instanser (`plan-db-2` är primary), AtlasSchema är Applied (22 tabeller, ägda av `plan`), Go API:t loggar "Connected to database" och "Seed data verified" och ett autentiserat anrop mot https://plan.simonbrundin.com/api/v1/agent-runs svarar 200 med session-cookie och 401 utan.

Lärdom: när en CNPG-kluster står kvar i 'Setting up primary' under lång tid ska man kontrollera Longhorn-volymens replicas och robustness, inte bara om podden startar. Ett kluster som serial>0 måste raderas och återskapas för att få en ny initdb.

## Relaterade

- [[concepts/cloudnative-pg]]
- [[notes/cnpg-initdb-agarollen-skapas-fore-postinitsql]]
- [[notes/cnpg-managed-roles-secretet-saknar-losenord-om-username-avviker-fran-rollnamnet]]
- [[notes/sops-kryptera-bara-stringdata-i-kubernetes-secrets-ellers-kraschar-flux-builden]]

