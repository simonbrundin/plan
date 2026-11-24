---
title: Målfaserna är gemensamma för alla mål i Plan
type: note
tags: [feature, goals, phases, agent, projekt:plan]
---

Plan använder samma fem faser för varje mål: define → plan → execute → verify → completed. Nya mål börjar i define; målets status_id måste skickas med både detalj- och prioriteringssvar. Fasbyten går via PATCH /api/v1/goals/:id/status så status_updates-historiken fortsätter gälla. Agent-flaggan är separat opt-in för agentarbete och agentens körningsstatus (queued/running/blocked/done/failed) är separat från fasen. completed sätter inte automatiskt finished. Detta återanvänder de befintliga statuses-raderna i stället för att införa parallell fasmodell.
