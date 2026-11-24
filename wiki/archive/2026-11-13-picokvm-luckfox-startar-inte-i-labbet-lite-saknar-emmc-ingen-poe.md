---
title: PicoKVM (Luckfox) startar inte i labbet: Lite saknar eMMC + ingen PoE
type: note
tags: [hardware, unifi, picokvm, felsokning, projekt:infrastructure]
---

Köpt Waveshare/Luckfox PicoKVM Lite (Amazon ASIN B0H4G22JG8, 'Luckfox PicoKVM Lite', full-version-hus med 1.54-tums display men UTAN eMMC).

Varför den inte startar:
1) Ingen ström från switchen. Enheten har ingen PoE-mottagare - matning sker enbart via USB-C 5V (medföljande typ-C-förlängningskabel med extra matning, eller USB-C-laddare, eller USB-porten på datorn som ska styras). UniFi US48P500 ger 48V 802.3af men enheten är inte PD, så ingen ström (och ingen risk, porten är 802.3af-detekterande).
2) Ingen eMMC => systemet MÅSTE flashas till ett microSD-kort (>=8GB, Class 10) innan första boot. Utan kort: svart display, ingen gul system-LED. Detta är exakt felet en Amazon-recension beskriver.

Kortstorlek: imagen är exakt 1494220800 B (1425 MiB) och innehaller INGEN GPT/MBR i sector 0 - det är en rå Rockchip-SD-image (parameter-block + uboot + systempartitioner) som enhetens bootloader partitionerar om kortet vid första boot. Layout: env 32K, idblock 512K, uboot_a/b 512K vardera, misc 256K, security 1M, boot_a/b 32M vardera, system_a/b 640M vardera (= ~1347 MiB fasta), och userdata = ALL rest av kortet. Wiki kräver >=8 GB (Class 10). Större kort = större userdata (utrymme för virtuella media/ISO). FAT32 har 4 GB-gräns per fil; Sys0.1.7+ kan omformatera userdata till exFAT.

Färdig bild (1.5 GB): https://drive.google.com/file/d/1g3jijNyVMSIILhP226WXOpNQwQXK7OO3/view -> Luckfox-PicoKVM-Lite_MicroSD_Sys0.1.9_App0.1.5.img (mappen firmware/Lite i https://drive.google.com/drive/folders/1c_1EdF3fCocWYCSY2KlshBpRdEznSKzy). Flashas med balenaEtcher eller dd.

Nät: RJ45 är 10/100M. Switchportarna 1-48 (utom 8) ligger i Default-VLAN 10.1.1.0/24, så enheten får DHCP-lease i 10.1.1.x och syns i UniFi stat/sta. Port 8 är Datacenter-VLAN 10.10.10.0/24.

Signaler: gul LED blinkar = kernel bootad, grön LED = nätverk. Web-UI: http://<ip> (eller picokvm.local). SSH root/luckfox.
