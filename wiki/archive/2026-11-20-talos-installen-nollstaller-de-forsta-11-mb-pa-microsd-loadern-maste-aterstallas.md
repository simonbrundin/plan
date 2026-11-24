---
title: Talos-installen nollstaller de forsta 11 MB pa microSD - loadern maste aterstallas
type: note
tags: [talos, pxe, worker-3, bug, projekt:infrastructure]
---

2026-11-20, worker-3 (ROCK 4C+, microSD-kort, ingen eMMC, ingen SPI-flash):

## FAKTA
Talos-installern (UnattendedInstallConfig) lägger partitionerna med första
partitionen på byte 11 534 336 (= 11 MB, sektor 22528). Innan dess nollställer
den området FÖRE den första partitionen - alltså de första 11 MB. Där ligger
Rockchips bootloader:

    idbloader.img  @ sector 64    (byte 32768)
    u-boot.itb     @ sector 16384 (byte 8388608)

Följd: varje installation till microSD raderar ut bootloader, och brädet startar
inte alls efteråt (svart skärm, noll paket, bara PHY-länk + cirka 2 W). Det är
samma fel som slagit ut worker-1.

## KONSEKVENS FOR DESIGNEN
"PXE-first-bootloader pa microSD" och "Talos installerat pa microSD" ar
inkompatibla pa den har plattformen. Alternativ:
  1. PXE-boot + maintenance mode (ingen installation) - repeterbart, men ingen
     kubelet, alltsa ingen medlem i klustret.
  2. Installera till NVMe (M.2) i stallet - da nollstalls kortet inte och
     bootloadern overlever. boot_targets innehaller "nvme".
  3. Installera till microSD och ATERSTALLA loadern efterat
     (pxe/Bootloaders/restore-loader.sh) - fungerar men maste goras manuellt
     varje gang, och kraver att kortet tas ut ur bradet.

## VERIFIERAT ATT STATE/EPHEMERAL INTE SKREVS TILL DISK
Efter en avbruten installation sag Talos:
    META      /dev/mmcblk1p2  ready
    STATE     /dev/mmcblk1p3  ready
    EPHEMERAL /dev/mmcblk1p4  failed ("open /dev/mmcblk1p4: no such file or
                                     directory")
Men GPT:en pa kortet innehall BARA EFI + META - STATE och EPHEMERAL fanns bara
i karnans vy och skrevs aldrig till disk. Det ar darfor installationen fastnar
i fas 5/9 (mountEphemeralPartition) och omstarten sedan hanger i teardown-fasen
14/14.

Fix: skapa partitionerna sjav pa kortet (parted/sfdisk) med ratt PARTLABEL
("STATE", "EPHEMERAL") och typ 0FC63DAF - samma som META. Da ser karnan dem
redan vid boot, volymhanteraren formaterar dem, och fas 5/9 passeras.
