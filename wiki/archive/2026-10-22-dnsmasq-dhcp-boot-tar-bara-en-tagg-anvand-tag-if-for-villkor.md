---
title: dnsmasq --dhcp-boot tar bara EN tagg - använd tag-if för villkor
tags: bug, pxe, dnsmasq, session
---

Fynd under arbete med pxe/dnsmasq.conf (2026-10-22). Verifierat empiriskt
med riktiga DHCP-förfrågningar mot en isolerad dnsmasq (pxe/test-dhcp-boot.sh).

Fakta:
- `dhcp-boot=[tag:<tag>,]<filename>[,<servername>[,<server addr>]]` - det är
  EN tagg per rad. Flera taggar med komma ("tag:!ipxe,!uefi-arm64,!bios") blir
  istället filnamn/servername-fält! En rad som såg ut att filtrera på tre
  villkor gav i praktiken bootfile="!uefi-arm64", sname="!bios".
- Negation fungerar i tagg-position ("tag:!ipxe" = klienten har INTE taggen
  ipxe).
- Villkorslogik ("A och inte B") måste byggas med tag-if:
  `tag-if=set:pxe-arm64,tag:uefi-arm64,tag:!ipxe` - varje "tag:<x>" är ett
  villkor och villkoren AND:as.
- Om flera dhcp-boot-rader matchar samma klient vinner den SISTA raden.
  Det är inte dokumenterat, så bättre att göra taggarna disjunkta med tag-if
  och slippa beroende av ordningen.
- HTTP-URL i bootfältet stöds ("dhcp-boot=tag:ipxe,http://host:8080/boot.ipxe")
  och är dnsmasqs egen dokumenterade lösning för iPXE-klienter.

Varför det spelar roll: PXE-kedjan behöver ge olika bootfiler beroende på
vilken klient som frågar - iPXE kan köra ett textskript, UEFI-firmware kan
inte det (får en EFI-binär via TFTP), legacy BIOS får pxelinux.0 och
U-Boot (ingen option 93) får boot.ipxe.
