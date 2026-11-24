---
title: PXE-bootloop: initramfs startar om for att boota installerad Talos pa eMMC
type: note
tags: [pxe, talos, worker-2, felsokning, projekt:infrastructure]
---

Symtom: worker-2 (10.10.10.22) startade om i en loop (~90 s) trots att hela
PXE-kedjan fungerade: DHCPACK -> talos/bsc -> kernel-arm64 -> initramfs ->
dtb -> bootefi -> initramfs hamtade configen via HTTP (talos.config).

Rotorsak: noden har en Talos-installation pa eMMC. Initramfs upptacker
installationen och startar om for att boota den (Talos startar aldrig sig
sjalv pa en befintlig installation). Eftersom U-Boots boot_targets har
"pxe" forst bootas PXE igen -> oandlig loop. Talos-dokumentationen sager
uttryckligen: "If there is already a Talos installation on the disk, the
machine will boot into that installation when booting from network. The
boot order should prefer disk over network."

Bekraftelse utan konsol: kernelparamet talos.halt_if_installed=1 gor att
Talos stannar i stallet och skriver ut ett meddelande ("already installed")
pa console=tty0 i stallet for att boota den installationen. Med parametern
upphor de nya bsc-hamtningarna - noden stallde. Det bevisade att en
installation finns pa disken.

Felsokningsmetoder som INTE fungerade:
- netconsole=10.10.10.2:6666 i bootargs: Talos-karnan skickar ingen
  kernellogg over nätet (ingen data pa UDP-lysnaren).
- talosctl -e insecure mot noden: "produced zero addresses" - nodens
  certifikat matchar inte klientens CA.

Viktigt fynd: nodens installerade Talos hade en GAMMAL PKI (CA
90:19:72:0F) medan klustret nu anvander B8:D2:AF:CE. En nod med en
utandande CA kan inte gora trog mot API-servern. Orsaken var att
worker-2-install.yaml inte hade regenererats - generate-node-configs.py
bygger install-configen ur den KORANDE controlplanet, sa den maste
regenereras nar klustrets PKI roterar. Nu har bade worker-2.yaml och
worker-2-install.yaml B8:D2:AF:CE.

Saker att beakta:
- U-Boots PXE-first-ordning och en installerad Talos pa eMMC ar i direkt
  konflikt. Anvand PXE for att installera om, inte for att daglig drift.
- En maintenance-config som innehaller KubeClusterConfig/KubeletConfig gor
  att noden GOR MED I KLUSTRET nar den bootar stabilt - inte "maintenance
  mode" i Talos-betydelse. Om noden inte ska anslutas ska configen strippas
  pa cluster-delen.
