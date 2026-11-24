---
title: 16x1 HDMI KVM - RS232/LAN-protokoll
type: summary
tags: [kvm, rs232, protokoll, projekt:simon]
sources: [raw/2026-11-13-16x1-hdmi-kvm-rs232-lan-communication-protocol.md]
created: 2026-11-13
updated: 2026-11-13
status: active
related: [entities/hdmi-kvm-switch-16-port-labbet, notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]
---

Källa: TESmart-manualen för HKS801/HKS1601 (kapitel 10.3), som är referensmanualen för en stor grupp OEM-byggen av 16-portars HDMI KVM-switchar (MLEEDA, EZNH, ANGEET, camgeet, gracetop).

Huvudpunkter:
1. RS232: 9600 8N1, ingen flow control, DB9 hona på KVM:en plus medföljande DB9 till 3-pinskruvterminal (GND/TX/RX). Vissa OEM-listningar anger 115200 - testa båda.
2. Kommandona är hex-frames: byt port = AA BB 03 01 NN EE (NN = 01..16), läs aktiv port = AA BB 03 10 00 EE, svar = AA BB 03 11 XX EE (0-baserad port).
3. Ytterligare kommandon: timeout AA BB 03 03 XX EE, buzzer AA BB 03 02 XX EE, auto-detect AA BB 03 81 XX EE (vissa byggen använder 02 i stället för 03 i prefixet).
4. LAN: fast IP 192.168.1.10, TCP-port 5000, samma frames. Ingen DHCP - kräver statisk IP på sidan.
5. RS232/LAN styr bara ingångsvalet - ingen video, USB eller data går över dem.
6. Auto-scan måste stängas av annars byter switchen port på egen hand och det ser ut som att styrningen är trasig.

Relaterat: [[entities/hdmi-kvm-switch-16-port-labbet]], [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]].

## Relaterade

- [[notes/kvm-styrning-fran-pxe-pi-n-16-port-hdmi-kvm]]
- [[entities/hdmi-kvm-switch-16-port-labbet]]
