---
title: PXE-först-bootloader för ROCK 4C+ byggd som patchad U-Boot-image
type: note
tags: [pxe, uboot, boot, worker, projekt:simon]
---

Radxa ROCK 4C+ har ingen UEFI-meny - U-Boot i SPI-flash väljer bootkälla via boot_targets. Defaulten 'mmc1 mmc0 nvme scsi usb pxe dhcp spi' lägger PXE sist. Lösning: byte-patcha den inbyggda default-env (bara 30 byte skiljer) till 'pxe dhcp mmc1 mmc0 nvme scsi usb spi' - verktyget pxe/Bootloaders/patch-boot-targets.py, artefakter i pxe/Bootloaders/Radxa Rock 4C+/. Viktig gotcha: en sparad env i SPI-flash (CONFIG_ENV_OFFSET=0x3F8000) vinner över den inbyggda defaulten - källbildens env-sektor är 0xFF, så att flasha HELA bilden nollställer env och den nya defaulten gäller direkt; skriver man bara u-boot-delen (rkdeveloptool wl 0x4000) måste env reras via U-Bootprompten (eraseenv) eller sättas manuellt. SPI-layout i bilden: idbloader @0x800, u-boot proper @0x800000 (sector 0x4000), resten 0xFF. Flash: rkdeveloptool db <spinor-loader> && rkdeveloptool wl 0 <bild> && rkdeveloptool rd. För 4SE (ingen SPI-flash) och Pi 4 (UEFI) finns ingen binär - där är setenv/saveenv respektive efibootmgr rätt väg.
