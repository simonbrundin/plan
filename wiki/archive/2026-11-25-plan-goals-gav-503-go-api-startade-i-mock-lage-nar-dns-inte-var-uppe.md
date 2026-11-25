---
title: plan /goals gav 503 - go-api startade i mock-läge när DNS inte var uppe
type: note
tags: [bug, incident, plan, kubernetes, projekt:simon]
---

Symtom: /goals i plan visade "Fel vid laddning av mål: [GET] \"/api/proxy/goals\": 503 Service Unavailable".

Felsökning (kind-dev):
- go-api-loggen: "Warning: Could not connect to database: unable to ping database: hostname resolving error: lookup ep-shy-brook-...neon.tech on 10.96.0.10:53: read: connection refused" och "Running in mock mode without database access".
- Samtliga DB-endpoints (goals, calendars, places) svarade 503 via checkDb() -> "Database not connected. Set DATABASE_URL environment variable."
- Kortslutning: inget i Go-koden eller i Nuxt-proxyn - 503 kommer från go-api, och frontendens ofetch-formatering gör att bara status-texten syns.

Rotorsak: värden (och därmed kind-dev-klustret) hade startat om ~09:53. go-api startade samtidigt som CoreDNS inte ännu svarade, så DNS-fel gjorde att Connect() misslyckades. go-api fortsätter i mock-läge (dbConnected=false) utan att försöka igen - det fanns ingen reconnect/retry.

Fix: DNS fungerade igen -> kubectl --context kind-dev rollout restart deploy/go-api -n plan. Loggen blev "Connected to database" + "Seed data verified".

Lärdom:
1. go-api ska inte fastna i mock-läge - antingen retry:a Connect med backoff vid start, eller koppla upp en DB-status i /health och låta liveness-probe restarta poden.
2. Frontendens felmeddelande borde visa API-svarets body ("Database not connected...") istallet for bara "503 Service Unavailable" - nuvarande feltext ar svagare for diagnos.
