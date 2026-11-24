---
title: Mac mini crashar varje gang man monterar den externa USB-stickan (msdosfs) - 2026-11-24
type: note
tags: [macmini, usb, crash, incident, felsokning, projekt:simon]
---

## Mönster (två oberoende crash-tillfällen samma dag)
1. ~17:25: jag monterade USB:ens ESP rw med mount_msdos (/private/tmp/m2) och den interna
   EFI-partitionen ro (/tmp/m0) -> direkt efter: "Connection timed out during banner exchange",
   sedan maskinen helt nere (ARP FAILED) - maskinen dog ALLTSÅ medans den monteras.
2. ~19:01: jag monterade ESP:en via sudo diskutil mount (mountades i /Volumes/EFI), skrev
   loader-entry + UKI, och provade bless -> "Connection timed out during banner exchange"
   och maskinen dog igen (ping fail, ARP FAILED).

Bada ganger dör maskinen inom ~1 min efter att en extern FAT32-volym monteras.
Misstanke: msdosfs-kexten + extern USB -> kernel panic/hang. Alternativt: sjuka USB-minnet
(62 GB DataTraveler) eller hårdvarstproblem.

## Lärdoms
- UNDVIK att montera den externa FAT32-volymen pa Mac minin frAn macOS.
- Rawoff skrivning (sudo dd if=X of=/dev/disk2s3) ar betryggare - ingen msdosfs inblandad.
- Saker som gick bra innan kraschen: UKI v1.14.1 (116 MB) kopierades till
  ESP:en som EFI/Linux/Talos-v1.14.1.efi, loader/entries/talos.conf skapades
  (default talos), gamla v1.12.x-UKI:er dolda som .bak, ESP:en avmonterad.
  Detta ska gå att verifiera nar maskinen ar igang igen (samt checksumma mot
  /tmp/talos-v1.14.1-metal-amd64-uki.efi).

## bless/NVRAM fungerar INTE pa denna Mac (viktigt)
- "sudo bless --folder /Volumes/EFI --setBoot --label Talos" -> fel 0xe00002e2
  (galler even FAT32-ESP monterad i /Volumes).
- "sudo bless --folder /Volumes/EFI/efi --setBoot" (med efi/boot.efi) testades inte
  innan kraschen.
- "sudo nvram efi-boot-device=..." -> "Error setting variable: (iokit/common) not permitted"
  = variabeln ar SIP-skyddad och kan inte skrivas frAn macOS.
- Konsekvens: automatisk boot frAn USB utan tangenttryck ar sannolikt INTE mojlig
  via bless/nvram pa denna maskin. Naga alternativ:
  a) Installera Talos pa den INTERNA disken (macOS skrivs over) -> da bootar maskinen
     Talos automatiskt (intern disk = forsta valet i firmware), USB:en behovs inte.
  b) Acceptera Option-tangenten vid varje start (valj "EFI"/Kingston i startmenyn).
  c) Testa om valet i startmenyn (Startup Manager) faktiskt persisterar - efter
     Talos-booten 2026-11-24 var efi-boot-device fortfarande macOS Preboot, vilket
     tyder pa att extern-volymsval INTE persisterar.
