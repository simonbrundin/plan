---
title: HA Krokviken: robotdammsugare startar/stoppar pa person-entiteter
type: note
tags: [home-assistant, automation, dammsugare, tuya, session, projekt:simon]
created: 2026-11-25
updated: 2026-11-25
status: active
related: [entities/home-assistant]
---

Automationer skapade via HA:s config-API: POST /api/config/automation/config/<id> med JSON-body (alias/triggers/conditions/actions). Skapade 'Dammsugare - ga tillbaka till station nar nagon kommer hem' (id vacuum_return_when_someone_home): trigger person.simon_brundin/person.emma_wallin -> 'home', condition NOT (vacuum.tesvor_s6 in docked/charging), action vacuum.return_to_base pa vacuum.tesvor_s6.

Viktigt: det fanns REDAN en automation 'Dammsug nar vi lammnar huset' (id 1765542273027) som startar dammsugaren nar BADE personerna lammat zone.home (zone-trigger + 'not' condition pa bada zonerna, action vacuum.start utan target). Min egen start-automation var en dubblett och togs bort igen (DELETE /api/config/automation/config/<id>).

Fallgropar: (1) Python-dict med dubbla nycklar tappade 'not' - conditionen blev omvand (skicka {'condition':'not','conditions':[{...}]} istallet). (2) '/api/config/automation/config/<id>' skriver YAML i /config/automations.yaml och laddar om automationsdomainen - samma id uppdaterar. (3) vacuum.tesvor_s6 (Tuya cloud) har state 'unknown' sedan 2026-11-24 17:13 - entiteten uppdateras inte, sa conditionen 'not docked/charging' alltid passerar (det vill saga: return_to_base forsokas varje gang nagon kommer hem). Integrationen bor undersokas.

## Relaterade

- [[entities/home-assistant]]

