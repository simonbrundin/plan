---
title: PicoKVM (Luckfox) startar inte i labbet: Lite saknar eMMC + ingen PoE
type: note
tags: [hardware, unifi, picokvm, felsokning, projekt:infrastructure]
created: 2026-11-13
updated: 2026-11-17
status: active
related: [notes/u7-lite-94-2a-6f-7a-4b-33-adopterar-inte-saknar-management-ip-och-nattverkssegment-pa-switchen]
---

Köpt Waveshare/Luckfox PicoKVM Lite (Amazon ASIN B0H4G22JG8, 'Luckfox PicoKVM Lite', full-version-hus med 1.54-tums display men UTAN eMMC).

Varför den inte startar:
1) Ingen ström från switchen. Enheten har ingen PoE-mottagare - matning sker enbart via USB-C 5V (medföljande typ-C-förlängningskabel med extra matning, eller USB-C-laddare, eller USB-porten på datorn som ska styras). UniFi US48P500 ger 48V 802.3af men enheten är inte PD, så ingen ström (och ingen risk, porten är 802.3af-detekterande).
2) Ingen eMMC => systemet MÅSTE flashas till ett microSD-kort (>=8GB, Class 10) innan första boot. Utan kort: svart display, ingen gul system-LED. Detta är exakt felet en Amazon-recension beskriver.

Kortstorlek: imagen är exakt 1494220800 B (1425 MiB) och innehaller INGEN GPT/MBR i sector 0 - det är en rå Rockchip-SD-image (parameter-block + uboot + systempartitioner) som enhetens bootloader partitionerar om kortet vid första boot. Layout: env 32K, idblock 512K, uboot_a/b 512K vardera, misc 256K, security 1M, boot_a/b 32M vardera, system_a/b 640M vardera (= ~1347 MiB fasta), och userdata = ALL rest av kortet. Wiki kräver >=8 GB (Class 10). Större kort = större userdata (utrymme för virtuella media/ISO). FAT32 har 4 GB-gräns per fil; Sys0.1.7+ kan omformatera userdata till exFAT.

Färdig bild (1.5 GB): https://drive.google.com/file/d/1g3jijNyVMSIILhP226WXOpNQwQXK7OO3/view -> Luckfox-PicoKVM-Lite_MicroSD_Sys0.1.9_App0.1.5.img (mappen firmware/Lite i https://drive.google.com/drive/folders/1c_1EdF3fCocWYCSY2KlshBpRdEznSKzy). Flashas med balenaEtcher eller dd.

Nät: RJ45 är 10/100M. Switchportarna 1-48 (utom 8) ligger i Default-VLAN 10.1.1.0/24, så enheten får DHCP-lease i 10.1.1.x och syns i UniFi stat/sta. Port 8 är Datacenter-VLAN 10.10.10.0/24.

Signaler: gul LED blinkar = kernel bootad, grön LED = nätverk. Web-UI: http://<ip> (eller picokvm.local). SSH root/luckfox.

## Status 2026-11-13: kortet flashat

SD-kortet (32 GB, i datorns inbyggda Realtek-läsare /dev/mmcblk0, udev-id mmc-SH32G) innehöll INTE ett tomt kort utan en GPT med ESP-partition 'RPI4-IPXE' - alltså ett iPXE-bootkort för Raspberry Pi 4 (RPI_EFI.fd, start4.elf, efi/boot/bootaa64.efi, firmware/brcm, config.txt med armstub=RPI_EFI.fd). Det är samma setup som beskrivs i pxe/Bootloaders/Raspberry Pi 4/README.md metod 2. Kortet var alltså en resurs för en Pi4-nod, inte ett oanvänt kort. 100 MiB backup av GPT+ESP togs till /home/simon/backup-rpi4-ipxe-sdcard-2026-11-13.img innan flash.

Flash: dd if=picokvm-lite.img of=/dev/mmcblk0 bs=4M conv=fsync, sedan nollresten (seek=356 conv=notrunc). Verifierat: cmp -n 1494220800 /dev/mmcblk0 image => exit 0. FALLGROP: sha256 via 'dd bs=4M count=356' jämfört med hela filens sha256 ger FALSK mismatch - 1494220800 B är INTE jämnt delbar med 4 MiB (4 MiB*356 = 1493172224 B), så count=356 läser 1 MiB för lite. Använd cmp -n eller head -c för exakta jämförelser.

Kortets läsning är i övrigt stabil: 10 upprepade läsningar av samma 8 MiB-region ger identisk hash (utom första läsningen direkt efter skrivning, som gav en avvikande hash - skrivningen var inte helt utflushad). dmesg visar 'Card stuck being busy' vid insättning, inga CRC-/timeout-fel under flashningen.

## Relaterade

- [[notes/u7-lite-94-2a-6f-7a-4b-33-adopterar-inte-saknar-management-ip-och-nattverkssegment-pa-switchen]]

