---
title: PXE-kedjan for worker-1 verifierad serverside; pxe-metoden failar snabbt utan pxelinux.cfg
type: note
tags: [pxe, uboot, rock4cplus, mainline, projekt:infrastructure]
---

Verifiering av hela kedjan serverside (2026-10-23):

1. TFTP fungerar: curl tftp://127.0.0.1/talos/bsc ga 1609 byte (= filens storlek).
2. TFTP-roten (/var/lib/tftpboot, mountad fran /home/simon/PXE/tftp) innehaller talos/bsc, talos/rock4c-plus/kernel-arm64 och initramfs-arm64.xz.
3. dnsmasq: dhcp-host for worker-1 MAC 0a:3f:39:25:2b:c8 -> tag rock4c-plus-1 -> 10.10.10.21. En U-Boot (ingen iPXE-userclass) matchar tag:pxe-fallback och far DHCP-option 'file' = boot.ipxe.

Viktiga U-Boot-funktionar (mainline 2026.10, kontrollerat i kallkoden):
- pxe_get() anvander env_get("bootfile"), INTE DHCP:s file-option. Vår env saknar bootfile, sa pxe-metaden sokor pxelinux.cfg/01-<mac>, pxelinux.cfg/<ip>, pxelinux.cfg/default* - ingen av dem finns i TFTP-roten, sa pxe_get returnerar -ENOENT och bootflowen hoppas over (snabbt, dnsmasq svarar 'file not found' direkt).
- Nasta method i boot_targets ar dhcp, som laser boot_script_dhcp=talos/bsc och kor filen som hush-script (cmd_source_script, utan mkimage-header).
- bootmeth_script.c väljer tftpb_run() istallet for dhcp_run() om ip_dyn=0 - darfor måste ip_dyn=1 ligga i env (mainline-patchen satter det).
- Env som verifierats i imagen: kernel_addr_r=0x02000000, ramdisk_addr_r=0x12180000, ip_dyn=1, boot_script_dhcp=talos/bsc, boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi, bootcmd=bootflow scan -lb.

Talos kernel-arm64 ar 'Linux kernel ARM64 EFI executable zstd compressed zboot Image' (MZ+PE) - alltsa en zstd-komprimerad EFI-stub. Den maste startas med bootefi (inte booti), och har boardens DTB inbakad - matchbox-receptet skickar ingen DTB och fungerar for 4SE/Pi4, sa ingen DTB behovs over TFTP. I bsc far initrd hamtas SIST sa att filesize innehaller initrd:ns storlek nar bootefi anropas.
