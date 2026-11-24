---
title: worker-1 svarar inte i nätet - felsökningsstatus
type: note
tags: [pxe, worker-1, rock4cplus, debugging, projekt:simon]
---

worker-1 (ROCK 4C+, MAC 0a:3f:39:25:2b:c8, IP 10.10.10.21, PoE-port 21) har under lång tid varit helt onåbar i nätet: ingen ARP, ingen DHCP, ingen TFTP, ingen synlig klient i UniFi, ingen bild på skärmen.

Vad som är UTSLUTET (genom verifiering):
- Fel eller föråldrad matchbox/dnsmasq-konfiguration - PXE-servern är verifierad och serverar korrekt nodconfig för worker-1.
- Fel U-Boot-image - repot innehåll en image med ROCK Pi 4A-DTB (se noten om mainline-bygget). Ersatt med en mainline-bygd 4C+-image, verifierad: RK33 @32768, FIT @8MB, DTB 'Radxa ROCK 4C+', boot_targets=pxe dhcp ..., boot_script_dhcp=talos/bsc.
- Fel i boot_targets/boot_script_dhcp - båda är nu satta rent i källkoden.

Vad som KVARSTÅR och måste testas fysiskt:
1. Strömlampa och PoE på port 21.
2. microSD-kortets kontakt/slot - boarden läser kortet alls tvekan.
3. eMMC-modul: RK3399:s BootROM söker SPI -> eMMC -> SD -> USB. Finns det en giltig loader i eMMC vinner den och kortet används bara som bootkalla (mmc1), aldrig som bootloader. 4C+ har eMMC som bestyckningsbar kontakt (tom i fabrik) - kontrollera worker-1.
4. Maskrom via USB-OTG: rkdeveloptool v1.3 byggd i ~/tools/rkdeveloptool, officiell loader rk3399_loader_v1.27.126.bin i /tmp/rkbin. Enligt Radxa: USB-A till USB-A-kabel i Ovre USB 3.0-porten, microSD-kortet BORT, håll Maskrom-knappen (4C+ har Reset/Maskrom/Recovery) medan strömmen slås på. lsusb ska visa 2207:330c.

 Verktyget: pxe/Bootloaders/verify-uboot-image.sh fångar fel board i 5 sekunder - kör det alltid före flash.

Obs: USB-läsaren i laptopen återenumereras varje sekund (kapaciteten rapporteras då fel, t.ex. 9.6 MB i st f för 29 GB) - partitioneringen på kortet misslyckades av den anledningen. Funktionellt är det kosmetiskt: SPL:en läser u-boot.itb pa fast offset 16384, och DTB:n ligger inbakad i FIT:en.
