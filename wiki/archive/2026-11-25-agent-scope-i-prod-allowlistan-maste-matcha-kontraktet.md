---
title: Agent-scope i prod: allowlistan måste matcha kontraktet
type: note
tags: [agent, api, scope, bug, projekt:plan]
---

Agentens egen API-nyckel har scope 'agent' i prod (hashen sätts in av environments/prod/secrets/plan-api-key.sql). Då gäller allowlistan i src/go/internal/middleware/agent_scope.go - i dev har nyckeln scope 'full', så avvikelser syns inte där.

Fynd 2026-11-25: POST /api/v1/goals/relations (anropet agenten gör i create_subgoal för att skapa undermål) var INTE tillåtet för agent-scope. Allowlisten behandlade varje segment efter /api/v1/goals/ som ett mål-id, så 'relations' föll på 'GET eller PATCH'. I prod skulle varje undermål ha fått 403 Forbidden.

Varför det inte syntes: dev-nyckeln har scope 'full', så samma anrop fungerade där. Felet fanns bara i prod-konfigurationen.

Fix + skydd (commit a1226a0): allowlisten tillåter nu POST på /api/v1/goals/relations. Kontraktet src/contracts/plan-agent.yaml markerar vilka anrop agenten gör (agent: true), och två tester håller ihop sidorna: agentens klientkällkod skannas så att varje anrop den gör är markerat, och Plan testar att varje markerat anrop är tillåtet för agent-scope.

Verifiering i dev: en API-nyckel med scope 'agent' gav 200/201 på samtliga 12 anrop agenten gör (inklusive goals/relations) och 403 på cancel och agent-run-lifecycle, som agent-scope ska neka.

Generell regel: dev-nyckeln har fel scope jämfört med prod. Varje ny agent-nyckel eller ny allowlistregel ska testas med scope 'agent', inte med sessionlogin.
