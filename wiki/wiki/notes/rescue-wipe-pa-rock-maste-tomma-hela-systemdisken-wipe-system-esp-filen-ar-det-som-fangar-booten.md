---
title: Rescue-wipe pa ROCK maste tomma hela systemdisken (wipe=system) - ESP-filen ar det som fangar booten
type: note
tags: [pxe, worker, recovery, bug, projekt:infrastructure]
created: 2026-10-26
updated: 2026-10-26
status: active
---

Varfor: en begransad wipe (talos.experimental.wipe=system:EPHEMERAL,STATE) racker INTE att stoppa en ROCK-worker fran att hamna i diskboot, eftersom ESP-paketet (EFI/BOOT/BOOTAA64.efi) ligger kvar. Nar ESP finns pa systemdisken vinner eMMC (mmc0) over PXE i U-Boots boot_targets, sa noden bootar disken i stallet for natverket, startar systemd-boot utan LoaderEntryDefault och hanger utan DHCP och utan API. Symptomet ar
