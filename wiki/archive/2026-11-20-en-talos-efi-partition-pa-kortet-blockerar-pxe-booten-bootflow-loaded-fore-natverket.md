---
title: En Talos-EFI-partition pa kortet blockerar PXE-booten (bootflow loaded fore natverket)
type: note
tags: [pxe, u-boot, worker-3, projekt:infrastructure]
---

2026-11-20, worker-3 (ROCK 4C+): varför kortet med Talos-EFI-partition inte går PXE

## REGELN
U-Boot kör EN gång och väljer sedan startkälla via bootflow:

    bootcmd=bootflow scan -lb
    boot_targets=mmc1 mmc0 nvme scsi usb pxe dhcp spi   (Radxa:s env)

"pxe" och "dhcp" är INTE bootmedia - de är metoder på nätverksenheten, och det
är deras POSITION i listan som avgör. En bootflow som nådde "loaded"-läget är
giltig - en installation som startar en kärna som sedan kraschar är FÖR FAR
giltig, och då når PXE aldrig.

## SYMPTOMET
worker-3 bootade via PXE kl. 13:50 med ett "rent" PXE-first-kort
(loader1/loader2/trust/boot/rootfs). När vi senare lade på Talos-partitionerna
(EFI/META/STATE/EPHEMERAL) slutade brädet att PXE-boota helt: länk upp,
1000 Mbit, ~2 W, men noll paket - ingen DHCP, ingen TFTP.

Förklaringen som stämmer med allt: med en Talos-EFI-partition på kortet hittar
U-Boot en GILTIG bootflow i "mmc1" (EFI-posten) och försöker starta den. Den
startar inte (Talos-EFI-stubben kan inte få igång kärnan utan rätt config) ->
U-Boot går inte vidare till "pxe"/"dhcp" -> tyst boot.

Radxa:s eget U-Boot (2026.01) hämtar dessutom DTB:n från EFI-partitionen.
Saknas den blir booten tyst (ingen bild, inget nätverk, ingen DHCP) - samma
symtom, ännu en anledning till varför felet är svårt att diagnosticera över
nätet.

## VERIFIERAT
- Kortet med Talos-layout: boarden skickar NOLL paket (även efter verifierad
  PoE-cykel: up=False, 0,00 W i 30 s).
- Kortet återbyggt exakt som det "rena" 13:50-kortet (loader1@64,
  loader2@16384, trust@24576, boot@32768 med DTB, rootfs@262144): se om det
  bootar igen - det isolerar variabeln.
- Loader-byten är nu verifierade byte-for-byte mot källfilerna:
  idbloader sha256 493a26e8..., u-boot.itb sha256 6732fa11... (kortet == fil).

## KORREKTION AV TIDIGARE SLUTSATS
"worker-1 visar svart skärm" bevisade ingenting: (a) vi tittade på fel
KVM-port (den deployade mappningen var fel), (b) mainline-U-Boot har konsol på
UART, inte HDMI, och Talos-kärnan skriver inte till HDMI - så en svart skärm är
det NORMALA läget även när boardet bootar fint.
