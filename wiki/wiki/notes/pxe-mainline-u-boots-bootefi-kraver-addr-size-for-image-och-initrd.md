---
title: PXE: mainline-U-Boots bootefi kräver addr:size for image och initrd
type: note
summary: Symtom: U-Boot skriver 'Booting /talos\bsc' followed by 'Not a PE-COFF file'.
tags: [pxe, bug, uboot, bootefi, projekt:infrastructure]
created: 2026-10-23
updated: 2026-10-23
status: active
related: [entities/pxe-server, notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot, notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist, notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]
---

Symtom: U-Boot skriver 'Booting /talos\bsc' followed by 'Not a PE-COFF file'. Noden laddar kernel+initrd via TFTP men startar inte Talos.

Rotorsak 1: bootefi-signaturen i mainline (cmd/bootefi.c) ar
  bootefi <image-adress>[:<storlek>] [<initrd-adress>:<initrd-lengd>] [<fdt-adress>]
Utan ':' efter image-adressen kräver U-Boot att imagen redan laddats av EFI-loadern
(efi_get_image_parameters) - annars 'No UEFI binary known at 0x2000000'.
Utan ':' efter initrd-adressen tolkas initrden SOM FDT-adress (fdt_addr = argv[2]),
sa EFI-stuben far initramfsen som device tree.

Rotorsak 2: 'Booting /talos\bsc' + 'Not a PE-COFF file' kommer fran boot/bootmeth_efi.c.
EFI-bootmetoden tar DHCP-filen (boot_script_dhcp = talos/bsc) och forsoker starta den
som PE/COFF - garanterat misslyckande, och den skriver over kernel_addr_r med
scriptet. Denna metod bor inte ligga i boot-floden for Rocky-bord.

Fix i bsc: spara storlekarna fran 'filesize' (satt av tftpboot) i egna env-variabler
och anropa:
  tftpboot ${kernel_addr_r} 10.10.10.2:talos/rock4c-plus/kernel-arm64
  setenv kernel_size ${filesize}
  tftpboot ${ramdisk_addr_r} 10.10.10.2:talos/rock4c-plus/initramfs-arm64.xz
  setenv initrd_size ${filesize}
  bootefi ${kernel_addr_r}:${kernel_size} ${ramdisk_addr_r}:${initrd_size}

Obs: 'filesize' innehaller bara den senaste filen - kernelstorleken maste sparas
direkt efter kernelhamtningen. Med initrd pa colon-form ar fdt_arg 3, sa med
argc 3 blir fdt_addr 0 -> EFI_FDT_USE_INTERNAL (U-Boots interna DTB).

Kallor: cmd/bootefi.c (do_bootefi), boot/bootmeth_efi.c, lib/efi_loader/efi_bootbin.c.

## Relaterade

- [[entities/pxe-server]]
- [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]]
- [[notes/worker-1-rock-4c-nar-pxe-via-u-boot-boot-targets-men-pxe-ligger-sist]]
- [[notes/worker-1-rock-4c-tyst-mainline-u-boot-byggd-men-boarden-laser-inte-kortet]]
