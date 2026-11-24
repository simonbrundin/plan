---
title: ROCK 4C+ och 4SE har ingen SPI-flash i fabrik - U-Boot ligger på microSD/eMMC
type: note
tags: [pxe, uboot, radxa, bug, worker, projekt:simon]
---

Radxa-wikin (wiki.radxa.com/Rockpi4/dev/spi-install) anger: 'Latest ROCK 4 SE, 4A+, 4B+, and 4C+ is unsupported, since there is no SPI Flash on the board'. Upstream-U-Boot-patchen 'rockchip: rock-pi-4-rk3399: Enable booting from SPI flash' sager samma: braden har footprint men inget chip (SOP8, 3.3V, W25Q64FV/GD25Q127CSIG/XT25F32B/XT25F128B kan lodas). Följd: repots pxe/UBOOT.md-claim att 4C+ bootar fran U-Boot i SPI-flash var for obekraft - U-Boot ligger i praktiken pa microSD eller eMMC. Darmed ar det INTE repots SPI-image (pxe/Bootloaders/rock4c-plus.bin) som bestammer bootordningen, utan U-Boot pa kortet - verifiera med 'printenv boot_targets' och 'sf probe' i U-Bootprompten, eller 'talosctl read /sys/class/mtd' pa noden (tomt = ingen SPI-flash). PXE-forst pa dessa noder gor ar med 'setenv boot_targets pxe dhcp mmc1 mmc0 nvme scsi usb spi; saveenv' i U-Bootprompten (env lagras i MMC), eller med iPXE pa microSD (mmc1 ligger forst i listan).
