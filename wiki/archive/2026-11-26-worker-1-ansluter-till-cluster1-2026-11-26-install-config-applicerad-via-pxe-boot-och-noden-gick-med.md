---
title: worker-1 ansluter till cluster1 (2026-11-26) - install-config applicerad via PXE-boot och noden gick med
type: note
tags: [pxe, worker-1, rock4cplus, cluster1, drift, projekt:infrastructure]
---

worker-1 (ROCK 4C+, 10.10.10.21, PoE-port 21) bootar numera PXE med matchbox och har gått med i cluster1 som Ready (Kubernetes) och som Longhorn-nod.

Varför noden inte gick med tidigare: noden bootade men stannade i maintenance-lage. API:t svarade pa 50000 men med ett certifikat signerat av en gammal CA (x509 unknown authority for live-klustret), sa noden kunde inte autentisera mot cluster1. Den gamla noten 'worker-1 svarar inte i natet' beskrev en hardvarufall (noll paket pa 1 Gbit/s) - det var en tidigare, separat incident. Boarden har sedan dessutom installerats pa eMMC (mmcblk1, serienummer 0x00002039) och bootar darifran (BOOT_IMAGE=/A/vmlinuz).

Vad som faktiskt gjorde noden anvbarbar:
1. ./pxe/generate-node-configs.py --only worker-1 - genererade nodens config mot den KORANDE controlplane-2 (10.10.10.12), sa CA/token matchar det nuvarande klustret. (Den gamla nodfilen hade foraldrad PKI.)
2. Noden svarade bara pa det OUTHALADE (maintenance-) API:t, sa configen applicerades med --insecure:
   talosctl apply-config --insecure -n 10.10.10.21 -f pxe/matchbox/assets/talos-config/nodes/worker-1-install.yaml -m auto
3. unattendedinstallstatus gick till 'installed' - partitionerna (META/STATE/EPHEMERAL pa mmcblk1) skapades. EPHEMERAL pa mmcblk1p6 = 29 GB.
4. Efter detta gick noden med i cluster1 direkt: medlemskap i talosctl get members, Ready i kubectl get nodes, Longhorn-nod registrerad, instance-manager poden startade (longhorn-image-pull tar ca 2.5 min, kyverno varnar om resource limits - det ar en farlig policy, inte ett fel).

Verifiering: talosctl reboot -n 10.10.10.21 - noden kom tillbaka, bootade om och blev Ready igen inom ~2 min. Ingen fysisk aterhamtning behovdes.

Viktig falla: worker-1:s UnattendedInstallConfig pekar pa disk.serial == '0x00002039' (eMMC), precis som worker-3. Inte '0xFFFFFFFF' som underhallsconfigen - sa installationen kor varje gang noden startar om. Det fungerar, men det ar inte samma uppstallning som worker-8 (maintenance-config).

Kvar att gora: worker's volymconfig (worker-1-volumes.yaml) kunde inte appliceras - 'the applied machine configuration doesn't contain v1alpha1 config'. Den maskinen har en USB-SSD (Intenso 128 GB, sda) och matcher INTE de symlinks som worker-1-volumes.yaml anger (Intenso_USB_3.0_device_... och en WDC WD30EZRZ som inte finns pa noden). Longhorn-konfigurationen for worker-1 anvander darfor bara default-disken /var/lib/longhorn/ precis som worker-3/5/6.
