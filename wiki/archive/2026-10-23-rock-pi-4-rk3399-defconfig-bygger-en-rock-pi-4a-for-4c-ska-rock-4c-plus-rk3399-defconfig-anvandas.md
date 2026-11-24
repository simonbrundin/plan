---
title: rock-pi-4-rk3399_defconfig bygger en ROCK Pi 4A - for 4C+ ska rock-4c-plus-rk3399_defconfig anvandas
type: note
tags: [ubygg, uboot, rock4cplus, mainline, felle, projekt:infrastructure]
---

Kontrollerat pa PXE-varden (Ubuntu 25.04) 2026-10-23: mainline-defconfigen rock-pi-4-rk3399_defconfig bygger en bild med DTB for 'Radxa ROCK Pi 4A' (rock-pi-4a.dtb) - INTE for ROCK 4C+. Det ar samma fel som det gamla rock4c-plus.bin i repot hade (där filen hette 4C+ men bar Pi 4A-DTB).

Ratt defconfig for 4C+ ar rock-4c-plus-rk3399_defconfig - den ger DTB 'Radxa ROCK 4C+' och ar det som repoartefakten u-boot-rockchip-pxe-first.bin etc. byggs fran.

Att en bild bar fel DTB gor att boarden inte startar alls: tyst skarm, dott natverk, ingen DHCP - eftersom Ethernet-PHY, regulatorer och skarm ligger i DTB:n.

Kortet i lasaren pa PXE-varden (/dev/sdb) hade just den felen: fdtfile=rockchip/rk3399-rock-pi-4a.dtb med U-Boot 2025.04. Det kan aldrig starta en 4C+ korrekt. Nu flashat med rock-4c-plus-rk3399_defconfig + patchen pxe-first-boot.patch:
  RK33 524b3333, FIT d00dfeed, U-Boot 2026.10, DTB 'Radxa ROCK 4C+'
  boot_script_dhcp=talos/bsc, ip_dyn=1, boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi

Miljo: PXE-varden har INGET /dev/sdc - kortet i lasaren heter /dev/sdb och /dev/sda ar systemdisken (root). Ett 'dd of=/dev/sdc' skulle alltsa antingen misslyckas eller, om en enhet dyker upp, riskera systemdisken. Verifiera alltid enheten innan flash.

Ubuntu 25.04 saknar paketet liblz4-tool (verktygen ligger i paketet lz4). Byggkrysskompilatorerna arm-none-eabi-gcc 14.2.1 och aarch64-linux-gnu-gcc 14.2.0 fungerade for TF-A och U-Boot 2026.10.

OBS: PXE-vardens rootfs var 100% fullt (117G, 0 byte fritt) - apt-get install misslyckades med 'No space left on device'. Atgarderades med apt-get clean + journalctl --vacuum-size. Efter frigjort utrymme (2,4G) gick klona+bygget att genomfora.
