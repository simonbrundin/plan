---
title: Mac mini (worker-9) har en bootbar Talos-USB - och dog under inspektion 2026-11-24
type: note
tags: [worker-9, macmini, talos, usb, incident, projekt:simon]
---

## Diskar på Mac mini (10.10.10.29), inspekterade via ssh 2026-11-24 17:2x
- /dev/disk0 (intern, 251 GB): ENDAST macOS 15.7.2 (APFS). EFI-partitionen (disk0s1) innehåller
  ENDAST Apple-filer (EFI/APPLE/...) - ingen Talos bootloader där.
- /dev/disk2 (USB, Kingston DataTraveler 3.0, 62 GB): DETTA ÄR TALOS-DISKEN.
  GPT: p1 = EFI (FAT32, 2 GB, C12A7328), p2 = 1 MB Linux-fs (Talos "STATE"-metadata?),
  p3 = 100 MB XFS med label STATE, p4 = 55,6 GB XFS med label EPHEMERAL.
  ESP-innehåll: EFI/boot/BOOTX64.efi (systemd-boot), EFI/Linux/Talos-v1.12.4.efi (100 MB),
  EFI/Linux/Talos-v1.12.5.efi (107 MB), loader/loader.conf ("timeout 10").
  => Komplett Talos-installation pA en 64 GB-stick; bootbar via systemd-boot.
- /dev/disk3 (USB, WD My Passport 2 TB): EN partition XFS UTAN label, ingen ESP -> inte en
  Talos-installation (troligen en ren datadisk).

## Incident: maskinen dog mitt under inspektionen
Efter att jag monterat ESP:en rw (/private/tmp/m2) och den interna EFI-partitionen read-only
(/tmp/m0) med mount_msdos svarade ssh med "Connection timed out during banner exchange"
och d sedan var 10.10.10.29 helt nere (pingar ARP FAILED, ingen IPv4, ingen IPv6-svar).
ARP-svep fran worker-8 (10.10.10.28) hittade bara en STALE IPv6-neighbor-post
(fe80::1c3e:4433:cb5c:f9e5 / a0:ce:c8:d9:0d:02) - maskinen svarade inte pa ICMPv6 heller.
Slutsats: maskinen är AV eller hangg - den behöver FYSISK power cycle. Ingen PoE pa
UniFi-port 29 sa det kan inte goras remote.

## Plan naar maskinen ar tillbaka
1. Kopiera in Talos v1.14.1 UKI (matchar klustret, v1.14.1) pa USB:ens ESP som
   EFI/Linux/Talos-v1.14.1.efi + systemd-boot-entry (default).
2. bootordning: "sudo bless --folder <ESP-mount> --setBoot --label Talos" (Intel-Mac:
   galler forsta starten) eller valj "EFI"/Talos i startmenyn (Option-tangenten, persistant).
3. USB:ens STATE-partition har GAMMAL PKI (node talar gammal generation) -> nar noden
   bootat: "talosctl -n 10.10.10.29 --insecure ..." for att fa tag i den och applicera
   nuvarande worker-9-config (pxe/matchbox/assets/talos-config/nodes/worker-9.yaml).
   Overvaetta att "talosctl reset --wipe-mode=system-disk" INTE anvands - systemdisken
   ar USB:en den bootar fran, och den interna disken innehaller macOS.
4. Klusterversion: controlplanes ar v1.14.1 (worker-5 ar v1.14.6) -> USB:ens v1.12.x
   ar for gammal for attjoina, detta ar anledningen till steg 1.

Skript klart: /tmp/macmini-talos-boot.sh (kopierar UKI, skriver ESP-entry, blessar).
