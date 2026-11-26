---
title: Tilt dev-server ska köras som serve_cmd
type: note
summary: En långlivad Nuxt-server ska vara en Tilt serve_cmd-resurs, inte ett avslutande cmd-build.
tags: [arbetsflöde, dashboard, dev, projekt:dashboard, projekt:simon, tilt]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [notes/tilt-filbevakning-kan-do-starta-om-tilt-up]
---

2026-11-26 upptäcktes att Dashboardens Tiltfile startade långlivade nub run dev --port 40807 via local_resource cmd. Tilt behandlade servern som ett build-kommando som aldrig avslutades; en manuell trigger lämnade resursen i build och visade Build Failed: Command did not terminate eller build canceled.

Ändringen är serve_cmd tillsammans med serve_dir i stället för cmd och dir. Tilt laddade den nya Tiltfile och uiresource dev blev Ready=True. serve_cmd kräver serve_dir; Tilt visar felmeddelande om vanlig dir används.

## Relaterade

- [[notes/tilt-filbevakning-kan-do-starta-om-tilt-up]]

