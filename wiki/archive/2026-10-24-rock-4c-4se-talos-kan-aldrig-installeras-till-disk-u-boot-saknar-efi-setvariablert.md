---
title: ROCK 4C+/4SE: Talos kan aldrig installeras till disk (U-Boot saknar EFI SetVariableRT)
type: note
tags: [pxe, talos, rock4cplus, efivarfs, systemd-boot, felsokning, projekt:infrastructure]
---

Status 2026-10-24: worker-1 (10.10.10.21) och worker-2 (10.10.10.22) gick med i
en installationsloop och fastnade sedan i diskboot.

ROT ORSAK (hardsvara, inte Talos):
Brasets U-Boot ("Das U-Boot") implementerar inte EFI-tjänsten SetVariableRT.
Karnan kan darmed inte mounta efivarfs rw:
  kernel: "Firmware does not support SetVariableRT. Can not remount with rw"
  talos:  "failed to install bootloader: failed to create efivarfs reader/writer: invalid argument"
Talos v1.14.1 installerar pa arm64 alltid med systemd-boot, och
installBootloader() avbryter installen i det steget. Da skrivs aldrig META
(cmd/installer/pkg/install/install.go: installBootloader -> handleMeta), och
STATE blir trasig sa narsta boot failar med
"failed to mount STATE: openfs failed: ... FSCONFIG_CMD_CREATE failed".
Raspberry Pi 4 (EDK2) klarar detta - darfor fungerar controlplane-installationen.

KONSEKVENS 2 (den som lurar en):
Installationen skriver EFI/BOOT/BOOTAA64.efi + EFI/Linux/Talos-*.efi till ESP
INNAN EFI-variablerna skrivs. Nar ESP finns pa eMMC tar U-Boot:s
"boot_targets=mmc1 mmc0 nvme scsi usb pxe dhcp spi" eMMC (mmc0) fore pxe, sa
noden bootar disken i stallet for nattverket. systemd-boot startar utan
LoaderEntryDefault (variabeln skrevs aldrig) och noden hanger DAR - utan DHCP,
utan API. Symptom: switch-porten har link men dnsmasq visar ingen DHCP och
talosctl ger "no route to host". En kall omstart hjalper inte.

LARNING: en installation pa dessa braden lampar alltsa noden. Gor INTE
talosctl apply-config -f worker-N-install.yaml pa ROCK 4C+/4SE.

ATERSTALLNING (kräver fysisk insats):
1. Flasha PXE-forst-kortet pa nodens microSD
   (dd if="pxe/Bootloaders/Radxa Rock 4C+/u-boot-rockchip-pxe-first.bin"
   of=/dev/sdX seek=64), eller "setenv boot_targets pxe dhcp mmc1 mmc0 ..." i
   U-Boots prompt. Da PXE-bootas noden igen.
2. Tom systemdisken (API:t svarar nar initramfs startat, men noden starter om
   i loop - fanga fonsfret):
   talosctl reset --wipe-mode=system-disk --reboot --graceful=false -n <ip>
   Alternativ utan API: kernelparametern talos.experimental.wipe=system i
   pxe/tftp/talos/bsc.hush - men den maste tas bort igen, annars wipear noden
   sig sjalv i en loop.
3. Noden bootar PXE, hamtar configen fran matchbox och starter tjansterna utan
   nagon diskinstallation - den gaar med i cluster1.

Noderna ska alltsa vara PXE-only. Det hander sig om exakt det anvandaren
kraver for worker-1 ("alltid PXE, aldrig fran disk").
