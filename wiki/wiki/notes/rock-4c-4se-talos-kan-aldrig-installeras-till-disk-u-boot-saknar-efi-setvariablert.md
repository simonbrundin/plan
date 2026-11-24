---
title: ROCK 4C+/4SE: Talos kan aldrig installeras till disk (U-Boot saknar EFI SetVariableRT)
type: note
tags: [pxe, talos, rock4cplus, efivarfs, systemd-boot, felsokning, projekt:infrastructure]
created: 2026-10-24
updated: 2026-10-25
status: active
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

LARNING (uppdaterad 2026-10-24, se nedan): en installation pa dessa braden
lampar noden i DISKBOOT (inte "ingen installation"). Men partitionerna MÅSTE
finns for att noden ska kunna gora med i cluster1 - sa installationen maste
goras, och noden maste samtidigt ha PXE forst i U-Boot.

VIKTIG KORREKTION (2026-10-24):
Tidigare sagdes att man inte alls ska applicera worker-N-install.yaml pa
ROCK. Det var fel - det gav en reboot-loop (se noten om
UnattendedInstallConfig). RATT arbetsordning:
  1. PXE forst i U-Boot (rescue-kort eller "setenv boot_targets pxe dhcp ...").
  2. Wipe systemdisken sa den trasiga STATE forsvinner.
  3. Applicera worker-N-install.yaml -> installeraren skapar
     EFI/META/STATE/EPHEMERAL och fallerar i installBootloader() (efivarfs).
     Det ar VATAT: reboot: false i configen, noden startar inte om.
  4. Applicera tillbaka worker-N.yaml (maintenance) sa installeraren inte kor
     om varje boot.
  5. Behåll PXE-forst-laget. Da fungerar STATE/EPHEMERAL och noden gar med i
     cluster1, samtidigt som den alltid bootar fran natet.

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
3. Noden bootar PXE och hamtar configen fran matchbox. Utan STATE kan den
   INTE ga med i cluster1 (kubelet och containern behover EPHEMERAL), sa
   install-configen maste appliceras for att skapa partitionerna - se
   "VIKTIG KORREKTION" ovan.

Noderna ska alltsa vara PXE-only (alltid boota fran natet, aldrig fran disk),
men de ska ha partitionerna pa eMMC. Det hander sig om exakt det anvandaren
kraver for worker-1 ("alltid PXE, aldrig fran disk").

AUTOMATISERAD ATERSTALLNING (2026-10-25):
pxe/recover-worker.sh i infra-repot gorer stegen 2-5 automatiskt nar
API-fonsfret oppnar: node-uppslag i talos/nodes.yaml, power-cycle via UniFi
PoE, hostname-verifiering, wipe av systemdisken, apply av install-config och
maintenance-config, samt verifiering av volymer och medlemskap i cluster1.
Det som INTE kan automatiseras är steg 1 - flashningen av rescue-kortet -
eftersom ESP-filen pa eMMC gor noden osbar over natverket.

Relaterade noter:
- talos-v1-14-utan-unattendedinstallconfig-oinstallerad-metal-nod-bootar-om-i-en-loop
- rock-4c-4se-utan-batteribackup-rtc-star-pa-2017-tls-mot-factory-talos-dev-misslyckas
- talos-image-cache-kraver-ephemeral-cirkulart-beroende-med-state-meta-pa-en-farsk-metal-nod
- recovery-av-en-rock-worker-kraver-fysisk-rattning-esp-filen-pa-emmc-fangar-booten
