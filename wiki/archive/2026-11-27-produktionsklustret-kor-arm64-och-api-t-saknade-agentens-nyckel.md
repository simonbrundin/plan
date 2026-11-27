---
title: Produktionsklustret kör arm64 och API:t saknade agentens nyckel
type: note
tags: [bug, migration, rust, produktion, projekt:plan]
---

# Produktionsklustret kör arm64 och API:t saknade agentens nyckel

Två fel upptäcktes när Rust-imagerna skulle appliceras i produktion.

## Klustret kör arm64, men Rust-imagerna byggdes bara för amd64

Agent-imagen pekade på `sha-1fdc501`, som aldrig publicerades. Den
publicerade taggen `sha-6777b92` fanns, men kubelet svarade:

    no match for platform in manifest: not found

Beslutet "bygg bara amd64" byggde på att klustret skulle köra amd64.
Det gör det inte - alla noder i `admin@cluster1` är arm64 (Talos på
ARM). Go-imagen är byggd multi-arch och fungerar därför, vilket gjorde
felet lätt att missa.

Lösning: byggflödet bygger nu varje arkitektur på en runner av samma
arkitektur (`ubuntu-24.04-arm` för arm64) och sätter ihop
manifest-listan i ett separat jobb. Arm64 under QEMU tar timmar för en
Rust-kompilering, så emulering är inte ett alternativ.

## Plan-API:t saknade AGENT_API_KEY

Agentens job controller hämtar agentkörningar från Plan med
`PLAN_API_KEY`. Men API:t validerar API-nycklar mot env-variabeln
`AGENT_API_KEY`, och den sattes aldrig i produktion. Loggen visade
"Invalid API key" i varje reconciliation - i tjugofyra timmar.

API:t får nu agentens egen `PLAN_API_KEY` ur samma secret
(`plan-agent-api-key`). Middleware:t letar efter namnet
`AGENT_API_KEY`, men värdet är det agenten faktiskt skickar. Dev hade
redan rätt ordning - där är båda nycklarna `dev-agent-key`.

Viktigt: agentens scope-gräns lever i API:t, inte i agenten. En nyckel
med agent-scope får arbeta med mål och körningar men aldrig besluta om
ett godkännande eller radera något.

## Lärdom

Ett byggflöde som begränsar sig till en arkitektur måste kontrolleras
mot den arkitektur som faktiskt körs. "Klustret använder amd64" var
ett antagande, inte en kontroll - `kubectl get nodes
-o custom-columns=NAME:.metadata.name,ARCH:.status.nodeInfo.architecture`
avgör det på en sekund.
