---
title: PXE-återställning: nodspecifika matchbox-profiler och maintenance-mode-flöde
type: note
summary: Varför: den generiska worker.yaml som matchbox serverade installerade Talos på /dev/sda utan hostname.
tags: [pxe, talos, infrastructure, worker-9, projekt:simon]
created: 2026-10-22
updated: 2026-11-17
status: active
related: [entities/pxe-server, notes/recovery-av-en-rock-worker-kraver-fysisk-rattning-esp-filen-pa-emmc-fangar-booten]
---

Varför: den generiska worker.yaml som matchbox serverade installerade Talos på /dev/sda utan hostname. På worker-1/2/3/5/6/7/8 är /dev/sda Longhorn-datadisken - en PXE-boot hade alltså kunnat skriva över Longhorn-data. En audit av infra-repots PXE-configer hittade dessutom att worker-9 var registrerad som Raspberry Pi 5/arm64 fast den är en Mac mini/amd64, att generic controlplane.yaml pekade på en diskSelector (model YongzhenWeiye, 128GB) som inte matchar de körande controlplanen, och att talos/generated/machine_config_worker-8.yaml bar worker-1:s serial (0x00002039) under worker-8-namnet.

Vad som gäller framöver:
1. Varje nod har en egen matchbox-profil (node-<nod>.json) och en MAC-grupp. En MAC-grupp kan inte själv byta talos.config-URL - matchbox renderar bara profilens Boot-objekt. Nodspecifika configer kräver alltså nodspecifika profiler.
2. PXE-boot levererar alltid en maintenance-config UTAN install-sektion. Installationen är en separat, manuell handling via <nod>-install.yaml. Generiska profiler pekar på generic-worker.yaml så en okänd MAC aldrig installerar något.
3. Tre filer per nod, åtskilda: <nod>.yaml (maintenance), <nod>-install.yaml (UnattendedInstallConfig med diskSelector.serial), <nod>-volumes.yaml (endast UserVolumeConfig). Volymconfigen får INTE appliceras förrän etikett, partitionstab och XFS-status är kontrollerade - Talos kan annars omformatera Longhorn-disken.
4. worker-9 är Mac mini/amd64 (UniFi: 'Mac Mini 64 GB', OUI CE LINK LIMITED), inte Pi 5. Den gamla pi5-1-gruppen togs bort från live-PXE-hosten eftersom den delade MAC med worker-9-gruppen.
5. Secrets: generiska och nodspecifika talos-configer, talos/generated/* och talos/talosconfig-new är inte längre versionsstyrda - de innehåller CA-nycklar, kubelet-token och privat nyckel. De genereras med pxe/generate-node-configs.py (talosctl gen secrets --from-controlplane-config) och kopplas till PXE-värden manuellt.
6. rpi_5 och rpi_generic ger byte-identiskt netboot-innehåll i Talos v1.14.1 (samma kernel, samma init/rootfs.sqsh). Skillnaden mellan Pi 4 och Pi 5 ligger i diskbilden, inte i filerna matchbox serverar - kernelhashen räcker alltså inte för att skilja dem åt.
7. Initramfs-filer med suffixet .xz är zstd-komprimerade, och de är inte bit-reproducerbara mellan nedladdningar (tidsstämpel/UUID). Verifiera innehållet efter uppackning, inte SHA-256.
8. Live-PXE-hosten har ingen Git och 98% diskutryttjande - distribution måste vara explicit filkopiering, inte git pull.

## Relaterade

- [[notes/recovery-av-en-rock-worker-kraver-fysisk-rattning-esp-filen-pa-emmc-fangar-booten]]
- [[entities/pxe-server]]
