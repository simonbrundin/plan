---
title: Gotcha: 'PXE först' gör noden till en PXE-nod (maintenance mode)
type: note
tags: [pxe, boot, drift, worker, projekt:simon]
---

Nodernas PXE-configer ar maintenance-configer (ingen machine.install-sektion), sa en PXE-start landar i Talos maintenance mode och skriver INGEN disk. Med boot_targets='pxe dhcp mmc1 ...' startar noden ALLTID via nätverket sa länge PXE-servern svarar - den startar aldrig sin egen installation.

Konsekvens: ett PXE-först-kort ska användas som RESCUE-kort, inte som permanent bootmedium. Byt tillbaka det vanliga kortet (eller 'setenv boot_targets mmc1 mmc0 nvme scsi usb pxe dhcp spi; saveenv') nar noden ska ga tillbaka till drift.

Nuans: 'bootflow scan -lb' faller vidare om PXE inte svarar (trasig server eller ingen server = noden bootar normalt), men en FUNGERANDE PXE-server ar det som stoppar noden fran sin egen installation. Det ar alltsa inte 'PXE g'ar saker' - det ar 'PXE alltid vinner' som ar problemet.

Anvandning: 49 byte skrivs pa kortet, sa det ar enkelt att patcha fram och tillbaka med pxe/Bootloaders/patch-sd-card.sh (--targets for att satta den ordinar ordningen).
