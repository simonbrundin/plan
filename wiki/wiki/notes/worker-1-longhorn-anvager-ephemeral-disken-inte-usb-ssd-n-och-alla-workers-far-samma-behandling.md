---
title: worker-1 Longhorn anvager EPHEMERAL-disken, inte USB-SSD:n - och alla workers far samma behandling
type: note
tags: [longhorn, worker-1, pxe, talos, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [concepts/longhorn]
---

Kartlagt 2026-11-26 efter att worker-1 gick med i cluster1.

Fynd: Longhorn-noderna i cluster1 använder ALLA default-disken /var/lib/longhorn/, som ligger på EPHEMERAL-partitionen (xfs /dev/mmcblk1p6 på eMMC), inte på USB-SSD:na. worker-1 har 26,9 GiB max (EPHEMERAL 29 GB), worker-4 57 GiB, worker-6 56 GiB. Ingen nod har en dedikerad Longhorn-disk aktiv trots att varje worker har en UserVolumeConfig-patch (ssd-1) i talos/patches/workers/<nod>/disks/.

Varför: worker-3:s u-ssd-1-volume har phase=failed, errorMessage 'filesystem type mismatch: gpt != xfs' - disken bär redan en XFS-partition med gammal Longhorn-data (longhorn-disk.cfg, diskName disk-7, 621 MB) och Talos kan inte provisionera om den utan wipe. worker-1:s ssd-1-patch pekar pa by-id 'usb-Intenso_USB_3.0_device_80200000000000000CBA-0:0' som inte finns pa noden (faktisk by-id: ata-Intenso_SSD_Sata_III_2024040210153, USB-bryggan: usb-Shenzhen_YongzhenWeiye_DD564198338A1-0:0 - samma bryggserial som pa worker-3, sallan by-id pa USB-bryggor ar inte unikt). worker-1:s ssd-2-patch pekar pa en WDC WD30EZRZ som inte finns alls.

Konsekvens: Longhorn-replikor ligger pa EPHEMERAL som rensas vid varje ominstallation av noden. Tre volymer (immich-library 100 Gi, arr-stack-media 100 Gi, arr-stack-downloads 50 Gi) star faulted med 'disks are unavailable; insufficient storage' - for att bygga om dem behovs mer fri lagring, och worker-1:s nya 21 GiB raknar inte ensam for 100 Gi x 3 replikor.

Beslut som anser vara anvandaren: (1) ska worker-1 fa en dedikerad Longhorn-SSD (korrigera by-id i patchen, wipea USB-SSD:n, patcha in UserVolumeConfig i full maskinconfig och peka Longhorn-disken pa /var/mnt/ssd-1) eller ligga kvar pa EPHEMERAL som resten av klustret; (2) ska bytet goras pa alla workers samtidigt eller bara pa worker-1; (3) ska de tre faulted-volymerna repareras - det kostar data.

## Relaterade

- [[concepts/longhorn]]

