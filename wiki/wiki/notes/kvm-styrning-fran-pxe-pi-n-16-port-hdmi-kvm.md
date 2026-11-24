---
title: KVM-styrning från PXE-Pi:n (16-port HDMI KVM)
type: note
tags: [kvm, rs232, pxe, hardware, home-assistant, projekt:simon]
created: 2026-11-13
updated: 2026-11-20
status: active
related: [entities/hdmi-kvm-switch-16-port-labbet, notes/picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe, notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard, summaries/16x1-hdmi-kvm-rs232-lan-protokoll]
---

Köpt 16-portars HDMI KVM-switch (MLEEDA, Amazon ASIN B0GCD38QN9) - samma OEM-familj som TESmart/EZNH. Målet: kunna byta port på KVM:en från Raspberry Pi 5 som är PXE-server (10.10.10.2, Ubuntu 25.04, pyserial redan installerat).

Varför RS232 och inte bara nätverket: enheten har bara LAN om den har RJ45 ('Network'-delen i paketet är okänd - RS232 är garanterat). RS232 kräver en USB-RS232-adapter (FTDI FT232RL, undvik PL2303) eftersom Pi:ns GPIO-UART är 3.3V TTL och inte RS232.

Protokollet är hex-frames: byt till port N = AA BB 03 01 NN EE, läs aktiv port = AA BB 03 10 00 EE (svar AA BB 03 11 XX EE, 0-baserat). Baud: 9600 enligt TESmart-manualen, 115200 enligt MLEEDA - testa båda. LAN-vägen (192.168.1.10:5000) använder samma frames.

Implementationen ligger i infrastructure/kvm/ (kvmctl.py CLI + kvm-api.py HTTP-API på :8099 + systemd-unit + HA-exempel). Testad på Pi:n mot en emulerad KVM över virtuell pty och TCP - inget hårdvarutest gjort ännu eftersom adaptern inte är köpt.

Gotchas att komma ihåg: auto-scan måste stängas av annars byter KVM:en port själv; API:t lyssnar utan auth på 0.0.0.0 - brandvägg om nätverket breddas.

## Mätning av RS232-porten (2026-11-15)

Multimeter (200 Ohm, kontinuitet) mot 12 V-adapterns ytterhylsa (= jord) gav **GND i mitten** av 3-pin-terminalen. Med KVM:en påslagen (DC 20 V, svart prober mot mitten) märktes **-8,82 V** mot den vänstra ytterskruven och **0 V** mot den högra.

KVM:ns RS232-port är en **3-pins skruvterminal** (INTE DB9 - manualen är missvisande). Tolkning: porten är **riktig RS232** (negativ idle), inte TTL. Vänster = KVM:s **TX**, mitten = **GND**, höger = KVM:s **RX**. En 3,3 V TTL-utgång (Pi:s GPIO eller FTDI-kort) kan alltså INTE kopplas in direkt - MAX3232 (eller USB-RS232) krävs.

Följd:implementationen använder Pi 5:s UART1 (GPIO0 TXD / GPIO1 RXD, `dtoverlay=uart1` -> `/dev/ttyAMA1`). GPIO14/15 undviks eftersom Pi 5:s firmware/debug skriver bootloggar där. `kvmctl.py` + `kvm-api.service` är installerade på Pi:n, config i `/etc/kvm-switch/config.json` (root:dialout, 640).

Varningar om MAX3232-moduler: **VCC ska vara 3,3 V** om TTL-sidan går till Pi:s GPIO - modulen ger TTL-utgång = VCC-nivå, och 5 V in i GPIO förstör pinnen. Header-märkningen (TXD/RXD) varierar mellan moduler - kontrollera silkscreen, och byt vid behov.

## Relaterade

- [[notes/simon-kvm-cli-kommando-for-att-byta-hdmi-kvm-port-fran-pxe-vard]]
- [[notes/picokvm-luckfox-startar-inte-i-labbet-lite-saknar-emmc-ingen-poe]]
- [[summaries/16x1-hdmi-kvm-rs232-lan-protokoll]]
- [[entities/hdmi-kvm-switch-16-port-labbet]]
