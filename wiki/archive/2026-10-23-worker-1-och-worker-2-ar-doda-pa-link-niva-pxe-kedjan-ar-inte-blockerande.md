---
title: worker-1 och worker-2 ar doda pa link-niva - PXE-kedjan ar inte blockerande
type: note
tags: [pxe, hardware, felsokning, rock4cplus, projekt:infrastructure]
---

Status 2026-10-23: worker-1 (10.10.10.21, PoE-port 21) och worker-2 (10.10.10.22, PoE-port 22) ar INCOMPLETE i ARP-tabell och svarar inte pa ping. worker-3 till worker-9 (10.10.10.23-29) ar alla uppe med korrekta MAC-adresser.

worker-1 har testats med mainline-U-Boot 2026.10 (self-contained DTB, env: boot_targets=pxe dhcp ..., boot_script_dhcp=talos/bsc, ip_dyn=1) och med den kanda-goda Radxa U-Boot 2026.01. Bade givser NOLL nätverkstrafik - ingen DHCP, ingen ARP, ingen TFTP. Anvandaren bekraftade att PXE aldrig fungerat for boarden; det som fungerade tidigare var Talos bootat fraan kortet.

Samma nattat for worker-1 och worker-2 (tvA angransande PoE-portar 21 och 22) talar for en gemensam orsak: PoE/port/stram for de tvA, inte bootloader eller PXE-konfiguration.

Beslutande test: samma microSD-kort i worker-3 (PoE-port 23, frisk) - om boarden bootar dar ar worker-1s maskinvara trasig; om den inte gor det ar kortet eller imagen.

Alla PXE-delar ar verifierade serverside: TFTP fungerar (talos/bsc 1609 byte, kernel-arm64 19 MB, initramfs-arm64.xz 87 MB), DHCP-mappning for worker-1:s MAC ar korrekt, och mainline-U-Boots pxe-metod fallerar snabbt (ingen pxelinux.cfg) sa att dhcp-metoden tar over. Nartet ar alltsa inte blockerande - bara maskinvaran.
