---
title: plan-db: CNPG-kluster fastnade i 'Setting up primary' på en Longhorn-volym utan replikor
type: note
tags: [cnpg, longhorn, incident, plan, projekt:simon]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [concepts/cloudnative-pg]
---

Plan-appens databas (CNPG-kluster plan-db i namespace plan) fastnade i 'Setting up primary' i 20 timmar. Symptom: Go API:t svarade 500 'Failed to create user session' med 'dial tcp 10.108.196.67:5432: connect: connection refused', och initdb-jobbet plan-db-1-initdb stod i Failed.

Rotorsak: Longhorn-volymen pvc-91f9c951 saknade helt replikor (replicas: {}, robustness unknown) och var detached - den kunde inte monteras. Dessutom vägrade CNPG skapa primärinstansen ('refusing to create the primary instance while the latest generated serial is not zero') eftersom serien på klusterresursen inte var noll.

Fix: radera det trasiga initdb-jobbet, radera den tomma PVC:n, radera CNPG-klustret (data fanns aldrig - initdb hade aldrig lyckats) och låt Flux återskapa det. Den nya volymen provisionerades på worker-9 och blev attached/healthy.

Lärdom: när en CNPG-kluster står kvar i 'Setting up primary' under lång tid ska man kontrollera Longhorn-volymens replicas och robustness, inte bara om podden startar. Ett kluster som serial>0 måste raderas och återskapas för att få en ny initdb.

## Relaterade

- [[concepts/cloudnative-pg]]

