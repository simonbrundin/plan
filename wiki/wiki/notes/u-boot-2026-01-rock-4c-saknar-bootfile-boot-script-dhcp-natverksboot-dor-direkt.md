---
title: U-Boot 2026.01 (ROCK 4C+) saknar bootfile/boot_script_dhcp - nätverksboot dör direkt
type: note
summary: Fynd via paketfångst (tcpdump pa PXE-varden) under omstart av worker-1: U-Boot gor DHCP men fragar INTE efter bootfilen - Parameter-Reque...
tags: [pxe, radxa, uboot, fakta, projekt:simon]
created: 2026-10-22
updated: 2026-10-22
status: active
related: [entities/pxe-server, notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran, notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat, notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]
---

Fynd via paketfångst (tcpdump pa PXE-varden) under omstart av worker-1: U-Boot gor DHCP men fragar INTE efter bootfilen - Parameter-Request listar bara Subnet-Mask, Default-Gateway, Domain-Name, DNS, Static-Route, NTP, MTU, Hostname. dnsmasq serverar korrekt BOOTP-faltet file "boot.ipxe" i OFFER, men U-Boot lagrar det inte -> ingen TFTP sker alls och noden faller tillbaka till mmc1 (kortets egen Talos).

Rotorsak: nodens inbyggda env (U-Boot 2026.01 pa kortet) innehaller BARA 11 variabler: pxefile_addr_r, kernel_addr_r, kernel_comp_addr_r, fdt_addr_r, fdtoverlay_addr_r, ramdisk_addr_r, kernel_comp_size, stdin, stdout, stderr, boot_targets. Varken 'bootfile' eller 'boot_script_dhcp' finns.

Varfor det spelar roll (U-Boot-kallan):
- 'pxe'-metoden (boot/bootmeth_pxe.c, u-boot,extlinux-pxe) -> extlinux_pxe_read_bootflow() anropar dhcp_run() + pxe_get(), och pxe_get() (cmd/pxe.c) laser env_get("bootfile"). Utan 'bootfile' avslutas metoden utan TFTP.
- 'dhcp'-metoden (BOOTFLOW_METHF_DHCP_ONLY) hanteras av SCRIPT-bootmetoden (boot/bootmeth_script.c, u-boot,script): script_check() avvisar bara PXE_ONLY, script_read_bootflow_net() laser env_get("boot_script_dhcp") och gor sedan dhcp_run(addr, fname, true) (lwIP-DHCP med bootfilnamn) eller tftpb_run(). Utan 'boot_script_dhcp' returnerar den -EINVAL och inget laddas.
- U-Boot 2026.01 har ingen klassisk NBP-PXE-metod kvar - bara extlinux-PXE (master: boot/bootmeth_pxe.c = extlinux, ingen bootmeth-dhcp.c, net/bootp.c saknar DHCP-bootmeth).

Konsekvens: en iPXE-textscript (tftp/boot.ipxe) kan aldrig anvandas av U-Boot - U-Boot kan inte tolka '#!ipxe'. Vagan ar antingen (a) U-Boot-prompten manuellt: dhcp + tftp kernel/initrd + bootefi  , (b) boot.scr.uimg (mkimage-script) + env-variabeln boot_script_dhcp, eller (c) iPXE pa kortets EFI-partition (da tar iPXE over nätverket och anvander matchbox-fljoden over HTTP som redan funkar for Pi-noderna).

## Relaterade

- [[entities/pxe-server]]
- [[notes/pa-rock-plattorna-finns-ingen-bootloader-i-hardvaran]]
- [[notes/rk3399-s-bootrom-soker-spi-emmc-sd-usb-hardkodat]]
- [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]]
