---
title: HDMI KVM-switch 16 port (labbet)
type: entity
tags: [kvm, hardware, labb, projekt:infrastructure, projekt:simon]
sources: [raw/2026-11-13-16x1-hdmi-kvm-rs232-lan-communication-protocol.md]
created: 2026-11-13
updated: 2026-11-15
status: active
related: [notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm, summaries/16x1-hdmi-kvm-rs232-lan-protokoll]
---

En 16-portars HDMI KVM-switch (MLEEDA, Amazon ASIN B0GCD38QN9) som ska styras från PXE-Pi:n (Raspberry Pi 5, 10.10.10.2, Ubuntu 25.04).

Egenskaper: 16 HDMI- och USB-ingångar, EDID-emulering per port, USB 3.0-hubb, SD/TF, audio, RS232 och (möjligen) RJ45 för TCP-styrning.

Varför RS232 valdes: RS232-porten är garanterad att finnas, medan LAN-delen i paketet är okänd. Kräver USB-RS232-adapter (FTDI FT232RL) eftersom Pi:ns GPIO-UART är 3.3V TTL, inte RS232. Alternativväg om adaptern inte ska köpas: LAN/TCP mot 192.168.1.10:5000, eller USB HID-gadget som skickar hotkeys.

Implementation: infrastructure/kvm/ - kvmctl.py (CLI), kvm-api.py (HTTP-API på :8099), systemd-unit, install.sh, probe.sh och ett Home Assistant-exempel. Testad mot en emulerad KVM (virtuell pty och TCP-socket) på Pi:n; inget hårdvarutest ännu.

## Uppmätt RS232-pinout (2026-11-15)

Porten är en **3-pins skruvterminal** (inte DB9 - manualen är missvisande). Skruvarna: **vänster = KVM:s TX** (idle **−8,82 V**), **mitten = GND**, **höger = KVM:s RX** (0 V). Porten är alltså **riktig RS232**, inte TTL - MAX3232 krävs mellan Pi:s 3,3 V GPIO och KVM:ns terminal.

Relaterat: [[summaries/16x1-hdmi-kvm-rs232-lan-protokoll]], [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]].

## Relaterade

- [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]]
- [[summaries/16x1-hdmi-kvm-rs232-lan-protokoll]]
