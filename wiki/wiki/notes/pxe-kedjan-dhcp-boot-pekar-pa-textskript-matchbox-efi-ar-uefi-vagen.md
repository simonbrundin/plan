---
title: PXE-kedjan: dhcp-boot pekar på textskript, matchbox.efi är UEFI-vägen
type: note
summary: Fynd under dokumentationsarbete i pxe/README.md (2026-10-22).
tags: [pxe, boot, dhcp, matchbox, projekt:infrastructure]
created: 2026-10-22
updated: 2026-10-22
status: active
related: [entities/pxe-server, notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image, notes/pxe-varden-korde-foraldrad-matchbox-nodconfiger-saknades-404, notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]
---

Fynd under dokumentationsarbete i pxe/README.md (2026-10-22).

Fakta:
- pxe/dnsmasq.conf ger varje nod dhcp-boot=tag:<nod>,boot.ipxe. boot.ipxe är ett iPXE-textskript (tftp/boot.ipxe), inte en EFI-binär.
- En UEFI-firmware kan inte köra ett textskript över TFTP. Filen fungerar därför bara för klienter som redan kör iPXE (t.ex. iPXE på SD-kort) eller som bootar via U-Boot bootmeth_pxe.
- pxe/tftp/matchbox.efi är iPXE 1.21.1 byggt för ARM64 UEFI (PE32+), med inbäddat skript 'chain http://10.10.10.2:8080/boot.ipxe'. Det är den fil som UEFI-firmware kan köra.
- matchbox serverar /boot.ipxe inbyggt (matchbox egen endpoint), så kedjan matchbox.efi -> /boot.ipxe -> /ipxe fungerar utan extra filer.

Varför det spelar roll: utan matchbox.efi i tftp/ saknas vägen för kall UEFI-boot. Filerna togs aldrig bort efter diskussion i sessionen.

Öppen fråga: ska dnsmasq skilja på klienttyp (option 60/93 client-arch) så att icke-iPXE-klienter får matchbox.efi och iPXE-klienter får HTTP-URL till boot.ipxe, enligt matchbox upstream-mönster?

## Lösning (2026-10-22)

pxe/dnsmasq.conf väljer nu bootfil per klienttyp: iPXE får HTTP-URL till
matchbox /boot.ipxe, UEFI ARM64 (option 93 = 11/18) får matchbox.efi via
TFTP, legacy BIOS (option 93 = 0) får pxelinux.0 och allt annat (t.ex.
U-Boot) får boot.ipxe. Se pxe/README.md och pxe/test-dhcp-boot.sh, som
verifierar beteendet med riktiga DHCP-förfrågningar mot en isolerad dnsmasq.

## Relaterade

- [[entities/pxe-server]]
- [[notes/pxe-forst-bootloader-for-rock-4c-byggd-som-patchad-u-boot-image]]
- [[notes/pxe-varden-korde-foraldrad-matchbox-nodconfiger-saknades-404]]
- [[notes/u-boot-kors-en-gang-boot-targets-styr-steg-2-inte-steg-1]]
