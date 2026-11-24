---
title: Radxa-bradena behover bootmedium - men Talos och bootloader maste ligga pa skilda medium
type: note
tags: [hardsvaru, radxa, worker-3, projekt:infrastructure]
---

2026-11-20, slutsats om hårdvaran for worker-1..8 (Radxa ROCK 4C+/4SE):

Varje brade maste ha ett bootmedium (microSD eller eMMC-modul) - det finns ingen
annan plats for bootloadern. ROCK 4C+/4SE har ingen SPI-flash i fabrik (bara en
SOP8-footprint) och eMMC-sockeln ar tom i vara braden. BootROM:ns ordning ar
SPI -> eMMC -> SD -> USB, sa utan kort (och utan eMMC) hamnar boarden i USB
download-lage (maskrom) och startar aldrig sjalv.

Det som gick sjalv idag var INTE kortet utan placeringen av Talos: installern
nollstaller de forsta 11 MB, och det ar precis dar bootloadern ligger.

Inventering i repot (talos/patches/workers/*/patch.yaml, systemdiskens serial):
  worker-1  0x00002039   microSD   (4C+)
  worker-2  0x00000d42   microSD   (4C+)
  worker-3  0x00000f9b   microSD   (4C+) - verifierat = /dev/mmcblk1
  worker-4  0x501003f2   NVMe      (4SE)
  worker-5  0x00000fb4   NVMe/eMMC (4SE)
  worker-6  0x5db80239   NVMe      (4SE)
  worker-7  0x185a42d8   NVMe      (4SE)
  worker-8  0x48500643   NVMe      (4SE)
  worker-9  DataTraveler 3.0 (Mac mini + USB-stick)

Alltsa: 4SE-bradena har redan den rena designen (Talos pa NVMe via M.2,
U-Boot pa microSD). Det ar 4C+-bradena (worker-1..3) som har Talos pa microSD
- och det ar de som varje gang tappar bootloadern vid installation.

Tre satt att losa det:
  A. eMMC-modul (16-64 GB) i sockeln -> eMMC = bootloader, microSD = Talos.
     BootROM når eMMC fore SD, och installern ror bara kortet.
  B. M.2 2230 NVMe -> NVMe = Talos, microSD = bootloader (samma som 4SE).
     KONTROLLERA HARDVARUREVISION: ROCK 4C+ V1.2 har 2230-slot pa toppen,
     tidigare revisioner har bara footprint utan kontakt.
  C. Loda SPI-flash (SOP8, t.ex. W25Q64FV) -> bootloadern i SPI, kortet kan
     anvandas fritt. Flashning sker via rkdeveloptool over USB-C i maskrom-lage.

Rekommendation: B om revisionen har M.2-slot, annars A. C ar billigast men
kraver lodarbete.
