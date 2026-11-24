---
title: Controlplane-recovery via PXE är inte konfigurerad
type: note
summary: Matchbox-profilerna node-controlplane-1..3 pekar på /assets/talos-config/nodes/controlplane-N.yaml, men den filen genereras inte: pxe/gen...
tags: [pxe, talos, bug, controlplane, projekt:simon]
created: 2026-10-22
updated: 2026-11-17
status: active
related: [entities/pxe-server, notes/talos-etcd-en-gammal-ephemeral-kopia-pa-controlplanet-ger-en-trasig-raft-log]
---

Matchbox-profilerna node-controlplane-1..3 pekar på /assets/talos-config/nodes/controlplane-N.yaml, men den filen genereras inte: pxe/generate-node-configs.py hoppar över noder utan patch-katalog under talos/patches/workers/, och controlplane-noderna har i stället talos/patches/controlplanes/controlplane-N.yaml. Nodspecifika configs i matchbox/assets/talos-config/nodes/ finns bara for worker-1..9. Foljd: en PXE-boot av en controlplane-node tar idag 404 pa talos.config. Inte saker att fixa utan att install-disken for varje controlplane ar kartlagd - da kan man inte lita pa UnattendedInstallConfig.

## Relaterade

- [[notes/talos-etcd-en-gammal-ephemeral-kopia-pa-controlplanet-ger-en-trasig-raft-log]]
- [[entities/pxe-server]]
