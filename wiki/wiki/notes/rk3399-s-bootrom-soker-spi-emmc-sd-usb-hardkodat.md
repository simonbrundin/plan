---
title: RK3399:s BootROM söker SPI → eMMC → SD → USB (hårdkodat)
type: note
summary: Verifierat mot U-Boot v2025.04 (arch/arm/mach-rockchip/spl.c: boot_devices[] = {EMMC: /mmc@fe330000, SPINOR: /spi@ff1d0000/flash@0, SD: /...
tags: [pxe, radxa, uboot, fakta, bug, projekt:simon]
created: 2026-10-22
updated: 2026-10-22
status: active
related: [entities/pxe-server, notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran, notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt, notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]
---

Verifierat mot U-Boot v2025.04 (arch/arm/mach-rockchip/spl.c: boot_devices[] = {EMMC: /mmc@fe330000, SPINOR: /spi@ff1d0000/flash@0, SD: /mmc@fe320000}) och Radxa-forumet ('the boot order is set in SOC's BootROM (SPI -> eMMC -> SD -> USB) and cannot be changed'). DT-alias i vår U-Boot-image: mmc0 = /mmc@fe330000 (sdhci = eMMC), mmc1 = /mmc@fe320000 (dwmmc = SD-kort).

Följd: en giltig U-Boot i eMMC vinner over microSD-kortet. Kortet anvands da bara som bootkalla (mmc1) och en patchad U-Boot-image pa kortet har DARMED effekt - noden bootar Talos fran NVMe som vanligt. På ROCK 4C+ är eMMC en bestyckningsbar kontakt som normalt är tom (Radxa: 'Removable eMMC - empty connector, not populated by default on 4c Plus'), så för 4C+ ska kortet vara bootloader-bärare - men det måste bekräftas per nod.

BUG hittad och fixat: pxe/flash.nu skrev U-Boot-bilder med 'dd bs=1M' (byte 0). Rockchip-bootROMen letar efter idbloader på sector 0x40 (byte 32768) och u-boot.itb på sector 0x4000 (byte 8 MB) - uppstroms doc/board/rockchip/rockchip.rst säger 'sudo dd if=u-boot-rockchip.bin of=/dev/sda seek=64'. Att skriva till byte 0 ger ett kort som INTE bootar. flash.nu skriver nu med seek=64 for U-Boot-bilder (och skippar partitioneringen for dem), och den nya pxe/flash-uboot.sh verifierar RK33-signaturen pa byte 32768 och FIT pa 8388608 fore och efter skrivning.

Verifieringsverktyg: pxe/Bootloaders/inspect-sd-card.sh laser (skriver inte) forsta 16 MB fran sector 64 och rapporterar RK33/FIT/U-Boot-version/boot_targets/sparad env (CONFIG_ENV_OFFSET=0x3F8000 fran kortets borjan - ar den raderad anvands den inbyggda defaulten, vilket ar det som galler for vart patchade kort).

Viktig regel: en sparad env vinner over den inbyggda defaulten. Vart patchade kort har env-arean raderad (0xFF), sa defaulten galler - men om nodens nuvarande kort har en sparad env maste den raderas (eraseenv) eller sattas manuellt (setenv boot_targets pxe dhcp mmc1 mmc0 nvme scsi usb spi; saveenv).

## Relaterade

- [[entities/pxe-server]]
- [[notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran]]
- [[notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt]]
- [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]]
