---
title: Korrigering: rock4c-plus.bin är ett microSD/eMMC-image, inte en SPI-image
type: note
tags: [pxe, uboot, radxa, korrigering, projekt:simon]
---

Fakta: pxe/Bootloaders/rock4c-plus.bin är INTE en SPI-image. Det är ett U-Boot-image i uppströmsformat 'u-boot-rockchip.bin' (idbloader @ 0x0, u-boot.itb @ 0x7f8000, data @ 0x800000), dvs. samma format som skrivs till microSD/eMMC. Verifierat genom att jmfara byte-for-byte mot inindev/uboot-rockchip v2025.04 (rk3399-rock-4c-plus.zip och rk3399-rock-4se.zip): samma idbloader-prefix, samma offset, samma env-rad. Den SPI-relaterade varianten 'u-boot-rockchip-spi.bin' har en helt annan layout (RK33 @ 0x1000) och ar inte det vi har.

Varfor det spelar roll: en SPI-image och ett microSD-image skrivs pa helt olika satt. microSD: 'sudo dd if=u-boot-rockchip.bin of=/dev/sdX seek=64 conv=fsync' (sector 64 = 32 KB, dR ar bootROM:en letar idbloader; u-boot.itb hamnar pa 8 MB = sector 0x4000). En SPI-image med rkdeveloptool. Tidigare wiki-noter och pxe/UBOOT.md beskrev repot som SPI-image - det var fel och ar rattat.

Metod for 'PXE first' utan chip: pxe/Bootloaders/patch-sd-card.sh patchar bootordningen pa microSD-kortet som redan sitter i noden - den laser U-Boot-blobben fran sector 64, patchar 'boot_targets' i minnet och skriver tillbaka de 49 byte som andrades (allt annat pa kortet lamnas ord). Testad: patch + revert ger en byte-exakt aterstallning av originalkortet. Verktyget fungerar pa valfri U-Boot-blob (t.ex. en nod med annan U-Boot-version).
