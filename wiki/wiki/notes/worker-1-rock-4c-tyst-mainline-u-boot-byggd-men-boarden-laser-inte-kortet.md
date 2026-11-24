---
title: worker-1 (ROCK 4C+) tyst: mainline-U-Boot byggd, men boarden läser inte kortet
type: note
summary: worker-1 (ROCK 4C+, 10.10.10.21) startar inte: ingen skarmbild, ingen ARP/DHCP/TFTP, ingen klient i UniFi.
tags: [pxe, uboot, rock4cplus, felsokning, projekt:infrastructure]
created: 2026-10-23
updated: 2026-10-23
status: active
related: [entities/pxe-server, notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first, notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot, notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]
---

worker-1 (ROCK 4C+, 10.10.10.21) startar inte: ingen skarmbild, ingen ARP/DHCP/TFTP, ingen klient i UniFi. worker-2 (ocksa 4C+) ar nere, worker-3..9 (4C+/4SE) ar uppe och leasear - sa hardvtyp, nat och PXE-uppsattning fungerar.

Byggde mainline-U-Boot 2026.10 enligt docs.u-boot.org/board/rockchip/rockchip.html: TF-A BL31 (PLAT=rk3399) + rock-4c-plus-rk3399_defconfig, flashat med dd seek=64 enligt guiden. Verifierade RK33/FIT-signaturer, inbakad DTB (Radxa ROCK 4C+), boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi. Boarden forblir tyst.

Tre gotchas i mainline som gjorde flodet tyst:
1) boot/bootmeth_script.c väljer tftpb_run (statisk IP) om CONFIG_CMD_TFTPBOOT ar enabled och ip_dyn ar 0/unset. Maste satta ip_dyn=1 i env for att dhcp_run (DHCP+TFTP) anvands.
2) TFTP-kommandot heter tftpboot (inte tftp): tftpboot <addr> <host>:<fil>.
3) bootefi tar initrd-storleken fra env-variabeln filesize (satt av sista tftpboot).

Verifierade mot guiden: DDR-parametrarna (rockchip,sdram-params) finns i SPL:ns DTB i idbloadern, sa DRAM-initen finns.

Felsokningsmetod som rekommenderas: aterstall en KAND-GOD loader pa kortet (radxa U-Boot 2026.01 + DTB i ESP) och testa. Delt resultat = boarden laser inte kortet/har ingen strom (hardware/maskrom). Ljust resultat = min loaderimage ar trasig. Anvande det for att isolera bildfel fran hardvarufel.

Miljoprat: sudo -n fungerar i toppnivens bash men INTE i subshell/script (PAM-session) - undvik pipes efter sudo-kommandon. USB-SD-lasaren flippar mellan full storlek och 9,6 MB, sa partitionering failar; anvand wipefs + sfdisk (atomisk) istallet for parted.

## Relaterade

- [[entities/pxe-server]]
- [[notes/u-boot-for-rock-4c-built-from-mainline-with-pxe-first]]
- [[notes/worker-1-kortet-flashat-med-ny-mainline-build-tf-a-fran-repot]]
- [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]]
