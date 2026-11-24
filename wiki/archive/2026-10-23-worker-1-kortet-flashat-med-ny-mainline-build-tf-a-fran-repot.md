---
title: worker-1-kortet flashat med ny mainline-build (TF-A fran repot)
type: note
tags: [pxe, uboot, rock4cplus, mainline, projekt:infrastructure]
---

Kortet i worker-1 flashades 2026-10-23 med ny mainline-build (TF-A f37f388 + U-Boot 2026.10, bada byggda i repot via bind-mount for att kringga mellanslaget i 'Radxa Rock 4C+').

Verifierat med inspect-sd-card.sh:
  idbloader (RK33) pa byte 32768 (sector 0x40)
  u-boot.itb (FIT) pa byte 8388608 (sector 0x4000)
  U-Boot 2026.10
  boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi
  Ingen sparad env (arean raderad) - den inbyggda defaulten galler

Hashar: u-boot-rockchip-pxe-first.bin e8505e54..., idbloader.img 493a26e8..., u-boot.itb 6732fa11...

USB-SD-lasaren enumererar intermitter kortet med FEL storlek (9,4 MB istan for 29,1 GB), vilket gor att parted/sfdisk skriver en ofullstandig partitionstabell (bara 'loader1' hamnade). Detta ar harmlost for booten: BootROM:en laser idbloadern fraan de raa sector 64, och med boot_targets = 'pxe dhcp ...' skannas MMC-partitionerna forst om nätbooten misslyckas.

Status: worker-1 och worker-2 (PoE-port 21 och 22) ar INCOMPLETE i ARP och svarar inte pa ping, medan worker-3..9 ar friska. Ingen DHCP/ARP/TFTP-traffik syns overhuvudtaget fran worker-1 - boarden kommer inte ens till U-Boots nätverksstack. Beslutande test: samma kort i worker-3 (port 23) - bootar det dar, sa ar worker-1s maskinvara/port trasig.
