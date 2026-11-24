---
title: PXE-kedjan: dhcp-boot pekar på textskript, matchbox.efi är UEFI-vägen
type: note
tags: [pxe, boot, dhcp, matchbox, projekt:infrastructure]
---

Fynd under dokumentationsarbete i pxe/README.md (2026-10-22).

Fakta:
- pxe/dnsmasq.conf ger varje nod dhcp-boot=tag:<nod>,boot.ipxe. boot.ipxe är ett iPXE-textskript (tftp/boot.ipxe), inte en EFI-binär.
- En UEFI-firmware kan inte köra ett textskript över TFTP. Filen fungerar därför bara för klienter som redan kör iPXE (t.ex. iPXE på SD-kort) eller som bootar via U-Boot bootmeth_pxe.
- pxe/tftp/matchbox.efi är iPXE 1.21.1 byggt för ARM64 UEFI (PE32+), med inbäddat skript 'chain http://10.10.10.2:8080/boot.ipxe'. Det är den fil som UEFI-firmware kan köra.
- matchbox serverar /boot.ipxe inbyggt (matchbox egen endpoint), så kedjan matchbox.efi -> /boot.ipxe -> /ipxe fungerar utan extra filer.

Varför det spelar roll: utan matchbox.efi i tftp/ saknas vägen för kall UEFI-boot. Filerna togs aldrig bort efter diskussion i sessionen.

Öppen fråga: ska dnsmasq skilja på klienttyp (option 60/93 client-arch) så att icke-iPXE-klienter får matchbox.efi och iPXE-klienter får HTTP-URL till boot.ipxe, enligt matchbox upstream-mönster?
