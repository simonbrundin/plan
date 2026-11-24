---
title: Gotcha: 'PXE först' gör noden till en PXE-nod (maintenance mode)
type: note
summary: Nodernas PXE-configer ar maintenance-configer (ingen machine.install-sektion), sa en PXE-start landar i Talos maintenance mode och skrive...
tags: [pxe, boot, drift, worker, projekt:simon]
created: 2026-10-22
updated: 2026-11-17
status: active
related: [entities/pxe-server, notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image, notes/talos-image-cache-kraver-ephemeral-cirkulart-beroende-med-state-meta-pa-en-farsk-metal-nod, notes/talos-noden-kan-hamna-i-maintenance-mode-om-state-volymen-hittas-efter-config-fasen, notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1, notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]
---

Nodernas PXE-configer ar maintenance-configer (ingen machine.install-sektion), sa en PXE-start landar i Talos maintenance mode och skriver INGEN disk. Med boot_targets='pxe dhcp mmc1 ...' startar noden ALLTID via nätverket sa länge PXE-servern svarar - den startar aldrig sin egen installation.

Konsekvens: ett PXE-först-kort ska användas som RESCUE-kort, inte som permanent bootmedium. Byt tillbaka det vanliga kortet (eller 'setenv boot_targets mmc1 mmc0 nvme scsi usb pxe dhcp spi; saveenv') nar noden ska ga tillbaka till drift.

Nuans: 'bootflow scan -lb' faller vidare om PXE inte svarar (trasig server eller ingen server = noden bootar normalt), men en FUNGERANDE PXE-server ar det som stoppar noden fran sin egen installation. Det ar alltsa inte 'PXE g'ar saker' - det ar 'PXE alltid vinner' som ar problemet.

Anvandning: 49 byte skrivs pa kortet, sa det ar enkelt att patcha fram och tillbaka med pxe/Bootloaders/patch-sd-card.sh (--targets for att satta den ordinar ordningen).

## Relaterade

- [[notes/talos-noden-kan-hamna-i-maintenance-mode-om-state-volymen-hittas-efter-config-fasen]]
- [[notes/talos-image-cache-kraver-ephemeral-cirkulart-beroende-med-state-meta-pa-en-farsk-metal-nod]]
- [[entities/pxe-server]]
- [[notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image]]
- [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]]
- [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]]
