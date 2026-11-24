---
title: worker-4 (ROCK 4SE): Talos bootar lokal NVMe-installation, apid startar aldrig - PXE-kedjan redo, bootordningen kravs fysiskt
type: note
tags: [felsokning, talos, pxe, rock4se, worker-4, projekt:infrastructure]
---

Diagnos 2026-11-20 (worker-4, 10.10.10.24, ROCK 4SE, MAC a2:83:f9:db:93:13).

Fynd (alla matta fran LAN:en via PXE-varden 10.10.10.2, inte via SD-WAN-tunneln):
- Noden lever: link 1000 Mbit, PoE ~4,3 W, svarar pa ARP och ping.
- DHCP-lease med hostname "worker-4" => Talos karn och maskin-config fungerar.
- MEN: full portscann (1-65535) fran LAN:en ger ENDAST 10250.
  - 50000 (Talos-API) -> connection refused => apid startar aldrig.
  - 10250 (kubelet) -> TLS-handshake "internal error" => kubelet utan giltigt
    servercert (fa apiserverkontakt). JAMFOR: worker-3 (frisk) har 50000 + 10250
    med korrekta cert.
- 60 s tcpdump pa PXE-varden: noden skickar INGEN trafik alls (ingen
  apiserver-, DNS- eller NTP-trafik). Den ar helt stilla - bara DHCP-renewal
  var 6:e timme och ARP-svar.
- PoE-cykel (off 40 s -> auto) startar om noden: DHCP-par (bootloader utan
  hostname + OS med hostname) - men NOLL TFTP-forrfragningar och NOLL
  matchbox-requests. Den bootar sin LOKALA installation pa NVMe, inte natet.
- Samma symptom pa worker-5/6/7 (alla ROCK 4SE). worker-1: dod board (0
  rx-paket, Class 0, 1,85 W). worker-2/8: ingen strom. worker-9: manuell boot.

Varfor detta ar ratt: ROCK 4SE har ingen SPI-flash - U-Boot ligger pa
microSD och boot_targets har natet SIST (mmc1 mmc0 nvme scsi usb pxe dhcp
spi), sa noden bootar NVMe-Talos vars apid/config ar trasig/gammal (forsta
PKI-generationen). PXE-serverkedjan ar bevisad OK: matchbox svarar med
iPXE-script for worker-4:s MAC (rock4se-assets, config worker-4.yaml),
matchbox-assets och config 10.10.10.24.yaml (med AKTUELL cluster-CA som
matchar live-klustret bit for bit) ligger pa PXE-varden.

DET SOM FAKTISKT BEHOVS (fysiskt): fa noden att boota fran natet.
  setenv boot_targets pxe dhcp mmc1 mmc0 nvme scsi usb spi
  setenv boot_script_dhcp talos/bsc-4se   (om env har boot_script_dhcp)
  saveenv
  reset
Sa PXE-bootar den, matchbox ger Talos i maintenance mode med API pa 50000,
och sa gaar pxe/recover-worker.sh worker-4 --yes (wipe + install-config ->
EFI/META/STATE/EPHEMERAL pa NVMe -> noden gaar med i cluster1).

NYTT I REPO/PA PXE-VARDEN (2026-11-20):
- pxe/tftp/talos/bsc-4se.hush (ny kallscriptkalla) + byggt/deployat
  talos/bsc-4se: laddar 4SE-kernel/initramfs (talos/rock4se/) och DTB:n
  dtb/rockchip/rk3399-rock-4se.dtb. Tidigare fanns BARA 4C+-DTB i TFTP-roten,
  sa en 4SE som bootar via U-Boots dhcp-script-metod fick fel DTB.
- rk3399-rock-4se.dtb extraherad ur repoartefakten
  pxe/Bootloaders/Radxa Rock 4SE/u-boot-rockchip-pxe-first.bin (inindev
  U-Boot 2025.04, FIT-en innehaller DTB for "Radxa ROCK 4SE"), deployad till
  /home/simon/PXE/tftp/dtb/rockchip/.
- MATTNING pa PXE-varden: PoE-cykel maste vara >= 30 s (porten bryter
  strommen efter ~30 s, trots att API:t svarar "ok" direkt). 20 s racker
  inte - noden startar inte om.

Relaterade:
- notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device
- notes/rock-4c-4se-talos-kan-aldrig-installeras-till-disk-u-boot-saknar-efi-setvariablert
- notes/radxa-bradena-behover-bootmedium-men-talos-och-bootloader-maste-ligga-pa-skilda-medium
