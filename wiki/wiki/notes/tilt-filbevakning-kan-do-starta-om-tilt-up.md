---
title: Tilt-filbevakning kan dö - starta om tilt up
type: note
tags: [tilt, dev, infra, lesson, projekt:plan]
created: 2026-11-16
updated: 2026-11-26
status: active
related: [notes/repo-template-rust-api-kopplad-till-kubernetes-manifests-tilt, notes/tilt-dev-server-ska-koras-som-serve-cmd]
---

Symtom: en långvarig 'tilt up'-instans i environments/dev upphörde att bygga om bilden när källfiler ändrades (inga nya byggloggar, deploymenten behöll gammal image-tagg). Lösning: döda den gamla processen och starta om 'tilt up' - den nya instansen fångade ändringar direkt.

Varför det händer: Tilt watcherar byggcontexten (docker_build), men i den här installationen kunde watcher-förlusten inte återskapas utan omstart av hela Tilt-processen.

FÄLLA: Tilt's UI-port (10350) kan vara upptagen av en annan Tilt-instans (t.ex. för ett annat repo i samma kluster) - starta då med egen port: 'TILT_PORT=<fri port> tilt up --file <Tiltfile>'. Ett manuellt 'kubectl set image' mot en deployment som Tilt hanterar skrivs över av Tilt vid nästa reconcilering - bygg via Tilt istället.

## Relaterade

- [[notes/tilt-dev-server-ska-koras-som-serve-cmd]]
- [[notes/repo-template-rust-api-kopplad-till-kubernetes-manifests-tilt]]
