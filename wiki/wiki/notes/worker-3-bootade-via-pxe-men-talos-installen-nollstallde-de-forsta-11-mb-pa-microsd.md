---
title: worker-3 bootade via PXE - men Talos-installen nollstallde de forsta 11 MB pa microSD
type: note
tags: [pxe, talos, worker-3, bug, projekt:infrastructure]
created: 2026-11-20
updated: 2026-11-20
status: active
---

2026-11-20: Forsta gangen lyckades en ROCK 4C+-nod boota via PXE hela vagen.
Kedjan verifierad i pcap pa PXE-varden (10.10.10.2) + dnsmasq + matchbox:

    13:50:10  TFTP  talos/bsc
    13:50:10  TFTP  talos/rock4c-plus/kernel-arm64
    13:50:15  TFTP  talos/rock4c-plus/initramfs-arm64.xz  (87 MB)
    13:50:37  TFTP  dtb/rockchip/rk3399-rock-4c-plus.dtb
    13:50:55  matchbox GET /assets/talos-config/nodes/10.10.10.23.yaml
    Talos v1.14.1 svarar pa API:t, hostname worker-3

## Varfor noden inte bootade PXE forut
Kortet i worker-3 hade U-Boot 2026.01 med
`boot_targets=mmc1 mmc0 nvme scsi usb pxe dhcp spi` - natet SIST. U-Boot hittade
den lokala Talos-installationen (EFI/BOOT/BOOTAA64.efi) pa kortet och bootade
den i stallet for nattet. Losningen: skriv om kortet med PXE-forst-U-Boot
(build-pxe-first-card.sh) -> boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi
+ boot_script_dhcp=talos/bsc.

## FALLAN: Talos-installen nollstaller de forsta 11 MB
Efter att install-configen applicerades pa microSD-kortet (mmcblk1) skapades
META (p2) och STATE (p3) och EPHEMERAL (p4) - men partitionstabellen grundar
sig pa byte 11534336 (= 11 MB). Installern nollstaller de forsta 11 MB, och
darmed forsvinner:
  - idbloader.img @ sector 64 (BootROMs laddarplats)
  - u-boot.itb     @ sector 16384 (0x4000)
Följd: boarden startar ingen loader alls - svart skärm, NOLL paket pa natet
(ingen DHCP), bara PHY-lank och ~2 W.

Fix: pxe/Bootloaders/restore-loader.sh - skriver tillbaka idbloader.img @
sector 64 och u-boot.itb @ sector 16384 UTAN att rora partitionerna, sa att
STATE/META/EPHEMERAL overlever.

## Andra lardomar fran samma test
- ROCK 4C+-bradena i detta natever har INGEN eMMC: `mmc0: Failed to initialize
  a non-removable card`. All Talos-installation ligger pa microSD-kortet, och
  install-configens diskSelector `disk.serial == "0x00000f9b"` matchar just
  microSD-kortet (Talos laser kortets serial via /sys/.../mmc1:0001/serial).
- Talos-installern fastnade i fas 5/9 (mountEphemeralPartition) for att karnan
  inte ser den nya partitionstabellen utan omstart -> "open /dev/mmcblk1p4: no
  such file or directory", och omstarten hangde sedan i teardown-fasen 14/14
  (unmountPromotable/unmountSystem gick inte klart). Losning: hard power-cycle
  via UniFi PoE, da ser karnan den nya partitionstabellen.
- U-Boot provar "pxe" (bootmeth_pxe) forst och letar efter
  talos/pxelinux.cfg/<mac> - det finns inte, sa det faller tillbaka till
  "dhcp" (boot_script_dhcp=talos/bsc) och det ar den vagen som fungerar.
- U-Boots PXE-metod behover ingen extra server: dnsmasq skickar DHCP option 54
  (server identifier), sa U-Boot vet att TFTP:a fran 10.10.10.2.
