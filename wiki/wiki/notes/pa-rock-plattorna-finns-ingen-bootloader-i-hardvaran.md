---
title: På Rock-plattorna finns ingen bootloader i hårdvaran
type: note
summary: RK3399 SoC:en innehaller bara en maskinvaru-stub, BootROM, som soker efter 'idbloader' (TPL+SPL) pa bootmedium och laddar det.
tags: [pxe, radxa, uboot, fakta, projekt:simon]
created: 2026-10-22
updated: 2026-10-22
status: active
related: [entities/pxe-server, notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat, notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt, notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]
---

RK3399 SoC:en innehaller bara en maskinvaru-stub, BootROM, som soker efter 'idbloader' (TPL+SPL) pa bootmedium och laddar det. All U-Boot (SPL, u-boot proper, dtb) ligger pa microSD eller eMMC - aldrig i chipet. Pi 4 har en motsvarighet i Pi-EEPROM (VideoCore-firmware) och Mac mini har iBoot - men da ligger bootloadern pa flash/media, inte i SoC:en.

Falskfoljd for vardra noderna: 4C+/4SE saknar SPI-flash i fabrik, sa utan microSD-kort (eller U-Boot i eMMC) finns det ingen bootloader alls - noden startar INTE. Det gora att 'PXE via rescue-kort' ar den enda fungerande vagen for en trasig nod: kortet med U-Boot ar inte en valfri bekvamlighet utan en forutsattning for att noden over huvud taget ska starta.

Dokumentationsfix i pxe/README.md: tabellen 'Firmware och bootloader per nod' skrev 'Rockchip BootROM + U-Boot' i firmware-kolumnen, vilket blandade ihop lageren. Nu står BootROM (firmware, i maskinvaran) och U-Boot (bootloader, pa microSD/eMMC) i varsin kolumn. Inget i koden hardkodar U-Boot i SPI-flash - det var felet.

## Relaterade

- [[entities/pxe-server]]
- [[notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat]]
- [[notes/u-boot-2026-01-rock-4c-saknar-bootfile-boot-script-dhcp-natverksboot-dor-direkt]]
- [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]]
