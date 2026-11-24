---
title: Tre satt att PXE-boota ROCK 4C+ och varfor dhcp-metoden valdes
type: note
summary: Radxa:s egen PXE-wiki (wiki.radxa.com/Rockpi4/dev/u-boot/pxe) beskriver U-Boot-native PXE: TFTP med kernel+initrd+dtbs/ och pxelinux.cfg/...
tags: [pxe, uboot, rock4cplus, projekt:infrastructure]
created: 2026-10-23
updated: 2026-10-23
status: active
related: [entities/pxe-server, notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first, notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot, notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]
---

Radxa:s egen PXE-wiki (wiki.radxa.com/Rockpi4/dev/u-boot/pxe) beskriver U-Boot-native PXE: TFTP med kernel+initrd+dtbs/ och pxelinux.cfg/01-<mac> + pxelinux.cfg/extlinux.conf. U-Boot kor pxe get + pxe boot, och devicetreedir laddar och DTB:n.

Men: matchbox levererar file "boot.ipxe" i DHCP-offerten, och U-Boot anvander den filen som config - en iPXE-text kan inte koras av U-Boot. Dessutom ar Talos kernel-arm64 en zstd-komprimerad EFI-stub (zboot, MZ+PE), sa booti klarar inte den - den maste startas via bootefi. Darfor anvands dhcp-metoden: boot_script_dhcp=talos/bsc + ip_dyn=1, och scriptet tftpbootar kernel+initrd och bootar med bootefi.

Kernelparametrarna i bsc ska vara identiska med matchbox-profilen node-worker-1.json: initrd=initramfs-arm64.xz init_on_alloc=1 slab_nomerge pti=on console=tty0 printk.devkmsg=on talos.platform=metal talos.config=<url>. Ingen extern DTB behovs - Talos-karnan har boardens DTB inbakad (matchbox-receptet skickar ingen DTB och fungerar for 4SE/Pi4).

Viktigt: i bsc far initrd hamtas SIST, sa att env-variabeln filesize innehaller initrd:ns storlek nar bootefi anropas (U-Boot far annars fel initrd-storlek). Alla tre metoder dokumenterade i pxe/UBOOT.md.

## Relaterade

- [[entities/pxe-server]]
- [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]]
- [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]]
- [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]]
