---
title: USB-SSD:orna har kvar giltig Longhorn-identitet ssd-1 med gamla replikor - de är bara inte registrerade i Longhorn
type: note
tags: [longhorn, ssd, worker, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [concepts/longhorn, notes/worker-1-longhorn-anvager-ephemeral-disken-inte-usb-ssd-n-och-alla-workers-far-samma-behandling]
---

Kartlagt 2026-11-26. Undersokning av alla workers USB-SSD (mountad read-only i debug-container) visar att de INTE ar kassa: de bar giltig Longhorn-diskidentitet och gamla replikor.

worker-1 sda1: longhorn-disk.cfg diskName=ssd-1 diskUUID=35c8b97c-097b-4b3a-87a9-97c5b5aac26e, replicas 2,3 GB, 10 PVC-kataloger.
worker-3 sda1: diskName=disk-7 (annat namn), diskUUID=b342bf79-93b9-456e-9d3d-a0455b5dbf5f, replicas 621 MB, 5 PVC-kataloger.
worker-4 sda1: diskName=ssd-1 diskUUID=9dc77a43-0359-42ae-914c-b391fde93511, replicas 10,5 GB, 15 kataloger.
worker-5 sda1: diskName=ssd-1, replicas 195 MB, 2 kataloger.
worker-6 sda1: diskName=ssd-1, replicas 11,0 GB, 17 kataloger.
worker-7 sda (hel disk, xfs utan partition): diskName=ssd-1 diskUUID=9a7e5553-fe8d-4ad1-864a-483d5fc2f801, replicas 750 MB, 1 katalog.
worker-8 sda: diskName=ssd-1, 0 replikor.

Viktigt: PVC-namnen i replikorna (t.ex. pvc-07928872, pvc-52955682, pvc-ca56b7ed) finns INTE i nuvarande cluster1-volymlista - de ar replikor fraan en tidigare klusterinkarnation. Datan ar alltsa doda, inte aktiva volymer.

Varfor det inte syns: Longhorn Node CR for worker-1..7 innehaller ENDAST auto-skapad default-disk-b30600000000 med path /var/lib/longhorn/ (pa EPHEMERAL). ssd-1-posten finns kvar i old-files/argo-manifests/.../longhorn-system/overlays/prod/nodes.yaml - en argo-fil som inte deployas. Endast worker-3 har installerad UserVolumeConfig, och den failar med 'filesystem type mismatch: gpt != xfs' (disken har GPT-partitionstabell med xfs i sda1, medan volumeType: disk vill ha xfs pa hela disken utan partition).

Kapacitet idag: total Longhorn-max 222 GiB (worker-1 26,9 / worker-3 26,9 / worker-4 57,0 / worker-5 26,9 / worker-6 56,0 / worker-7 28,5), tillgangligt 121,7 GiB, schemalagt 172,9 GiB - dvs worker-3 (138%), worker-5 (121%), worker-4 (90%) och worker-6 (72%) ar overallokerade, vilket forklarar 'insufficient storage'.

Korrigering av tidigare not: EPHEMERAL rensas INTE vid reboot eller upgrade (Talos wipear inte EPHEMERAL sedan 1.8). Den rensas vid talosctl reset - alltsa vid ominstallation, vilket ar exakt vad PXE-floden med install-config gor.

## Relaterade

- [[notes/worker-1-longhorn-anvager-ephemeral-disken-inte-usb-ssd-n-och-alla-workers-far-samma-behandling]]
- [[concepts/longhorn]]
