---
title: worker-2 och worker-9 saknas fortfarande som cluster1-workers - diagnos 2026-11-26
type: note
tags: [worker-2, worker-9, pxe, felsokning, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Status 2026-11-26 efter att worker-1 repaired.

worker-2 (10.10.10.22, ROCK 4C+): svarar pa Talos-API:t men med ett certifikat fran en gammal CA (x509 unknown authority utan --insecure). Med --insecure: META-partitionen (mmcblk1p4) ar ready men STATE-partitionen saknas helt, och /dev/sda saknas (ingen USB-SSD ansluten). Noden star alltsa i maintenance med en halvinstallerad eMMC och ingen Longhorn-disk. Patchen talos/patches/workers/worker-2/patch.yaml anger disk.serial 0x00000d42, vilket stammer med nodens mmcblk1. Mangden att gora: noden behover en full ominstallation via install-config (samma spar som worker-1) och en USB-SSD ansluts for Longhorn.

worker-9 (10.10.10.29, MacMini): svarar inte alls - 'no route to host' pa 50000, ingen ARP, ingen ping. Behover fysisk kontroll (strom/PoE/nat) innan nagonting kan goras programmatiskt.

Viktigt: den serverade PXE-configen per IP (t.ex. /assets/talos-config/nodes/10.10.10.21.yaml) har fortfarande UnattendedInstallConfig med disk.serial == 0xFFFFFFFF, alltsa ingen installation - en ren PXE-start av worker-1 skulle inte installera nagon systemdisk. Den installerade noden bootar nu fran sin eMMC (/A/vmlinuz), sa detta paverkar bara en ny PXE-start (t.ex. med rescue-kort).
