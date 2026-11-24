---
title: PXE: boot_script_dhcp kräver mkimage-uImage, inte ren hush-text
type: note
summary: Symtom: worker-1 (Rock 4C+) hämtar talos/bsc via TFTP men laddar aldrig kernel-arm64/initramfs-arm64.xz.
tags: [pxe, bug, uboot, rock4cplus, projekt:infrastructure]
created: 2026-10-23
updated: 2026-10-23
status: active
related: [entities/pxe-server, notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first, notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot, notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]
---

Symtom: worker-1 (Rock 4C+) hämtar talos/bsc via TFTP men laddar aldrig kernel-arm64/initramfs-arm64.xz. Booten misslyckas och noden startar om (ny DHCP-DISCOVER).

Rotorsak: mainline-U-Boots bootmeth_script kör det hämtade scriptet via cmd_source_script() -> image_locate_script(), som kräver att filen är ett legacy mkimage-uImage (ih_magic 0x27051956) eller en FIT. En ren hush-textfil ger 'Wrong image format for source command' och scriptet körs aldrig.

Fix: mkimage -T script -C none -n talos-pxe -d bsc.hush bsc  (talos/bsc pa PXE-servern blir uImage; kallatexten sparas som bsc.hush).

Repro: pxe/Bootloaders/build-bsc.sh [--deploy] bygger uImage-filen ur pxe/tftp/talos/bsc.hush och kan deploya den till PXE-servern. Den genererade filen och Talos-bootartefakterna (kernel/initramfs ~106 MB) gitignoras - kalltexten versionshall. CONFIG_LEGACY_IMAGE_FORMAT=y i rock-4c-plus-rk3399_defconfig, sa legacy-formatet fungerar.

Verifiering: dnsmasq-logg ska visa 'sent /var/lib/tftpboot/talos/rock4c-plus/kernel-arm64' och initramfs-arm64.xz efter bsc.

Kallan: u-boot/boot/bootmeth_script.c, common/command.c (cmd_source_script), boot/image-board.c (image_locate_script).

Verifierat 2026-10-23 (efter fix): worker-1 startade om och loggen visar
  sent talos/bsc -> sent talos/rock4c-plus/kernel-arm64 -> sent talos/rock4c-plus/initramfs-arm64.xz
Matchbox serverade /assets/talos-config/nodes/worker-1.yaml (Talos hamtade configen)
och Talos API-porten 50000 svarar - noden star i maintenance mode.
worker-1 ar medvetet INTE med i klustret (bara controlplane-1..3 ar medlemmar):
PXE-configen ar en maintenance-config utan install-sektion, sa installationen
gor manuellt med talosctl apply-config --config worker-1-install.yaml -n 10.10.10.21.

Notera: en andra bootcykel efter initramfs hamtade bsc + dtb men inte kernel
(U-Boot fortsatte till nasta boot_target). Det gick inte att avgora varfor utan
seriekonsol. Har ingen fungerande foljdfald: noden star stabil i maintenance mode.

## Relaterade

- [[entities/pxe-server]]
- [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]]
- [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]]
- [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]]
