---
title: Kalenderhändelser återställda från Neon WAL-historik
type: note
tags: [bug, incident, kalender, neon, projekt:plan]
---

2026-11-20: Neon dev-branchen hade 15 rader i public.calendar_events vid 13:42 UTC men 0 från 13:43 UTC. Go API-loggarna visade inga DELETE-anrop för kalenderhändelser; Tilt-klienten var dessutom otillgänglig trots kvarvarande process. Händelserna lästes read-only från Neon time-travel och återställdes efter användarens uttryckliga godkännande med ursprungliga id:n, parent_event_id, kalenderkopplingar, tider, recurrence_rule, status och tidsstämplar. Därefter verifierades 15 rader i Neon och GET /api/proxy/calendar-events?scope=all returnerade 15. Orsaken till borttagningen är fortfarande okänd och måste utredas innan fler kalenderändringar.
