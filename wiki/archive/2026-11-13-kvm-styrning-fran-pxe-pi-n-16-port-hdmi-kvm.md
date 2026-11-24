---
title: KVM-styrning från PXE-Pi:n (16-port HDMI KVM)
type: note
tags: [kvm, rs232, pxe, hardware, home-assistant, projekt:simon]
---

Köpt 16-portars HDMI KVM-switch (MLEEDA, Amazon ASIN B0GCD38QN9) - samma OEM-familj som TESmart/EZNH. Målet: kunna byta port på KVM:en från Raspberry Pi 5 som är PXE-server (10.10.10.2, Ubuntu 25.04, pyserial redan installerat).

Varför RS232 och inte bara nätverket: enheten har bara LAN om den har RJ45 ('Network'-delen i paketet är okänd - RS232 är garanterat). RS232 kräver en USB-RS232-adapter (FTDI FT232RL, undvik PL2303) eftersom Pi:ns GPIO-UART är 3.3V TTL och inte RS232.

Protokollet är hex-frames: byt till port N = AA BB 03 01 NN EE, läs aktiv port = AA BB 03 10 00 EE (svar AA BB 03 11 XX EE, 0-baserat). Baud: 9600 enligt TESmart-manualen, 115200 enligt MLEEDA - testa båda. LAN-vägen (192.168.1.10:5000) använder samma frames.

Implementationen ligger i infrastructure/kvm/ (kvmctl.py CLI + kvm-api.py HTTP-API på :8099 + systemd-unit + HA-exempel). Testad på Pi:n mot en emulerad KVM över virtuell pty och TCP - inget hårdvarutest gjort ännu eftersom adaptern inte är köpt.

Gotchas att komma ihåg: auto-scan måste stängas av annars byter KVM:en port själv; TX/RX kan behöva bytas (DB9-hona har ofta korsade signaler); API:t lyssnar utan auth på 0.0.0.0 - brandvägg om nätverket breddas.
