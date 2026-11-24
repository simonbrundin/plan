---
title: 16x1 HDMI KVM RS232/LAN communication protocol
type: source
added: 2026-11-13
origin: https://www.media-amazon.com/images/I/B1OMK10mvIL.pdf
immutable: true
---

# 16x1 HDMI KVM - RS232/LAN communication protocol

Source: TESmart "HKS801-EB23 & HKS1601-EB23 User Manual" (kapitel 10.3),
publicerad av TESmart. Samma protokoll används av en stor grupp OEM-byggen
(MLEEDA, EZNH, ANGEET, camgeet, gracetop) av 16-portars HDMI KVM-switchar.

## RS232

- Baud rate: 9600 bps (vissa OEM-listningar anger 115200 - testa båda)
- Stop bit: 1, data length: 8, parity: none
- Ingen flow control
- DB9 hona på KVM:en, medföljande DB9 -> 3-pinskruvterminal (GND/TX/RX)

### Kommandon (hex)

| Index | Kommando | Parameter | Betydelse |
|---|---|---|---|
| 1 | AA BB 03 01 XX EE | XX = 01..16 | Byt ingångskälla |
| 2 | AA BB 03 03 XX EE | XX = 0A/1E/00 | LED-display timeout 10s/30s/aldrig |
| 3 | AA BB 03 02 XX EE | XX = 00/01 | Mute/unmute buzzer |
| 4 | AA BB 03 81 XX EE | XX = 00/01 | Auto-detect av/på (vissa: AA BB 02 81 ...) |
| 5 | AA BB 03 10 00 EE | - | Läs aktiv port |
| 6 | AA BB 03 11 XX EE | XX = 00..0F | Svar: aktiv port, 0-baserad |

## LAN (TCP/IP)

- IP: 192.168.1.10 (fast, ingen DHCP)
- Port: 5000
- Gateway: 192.168.1.1, mask 255.255.255.0
- Samma hex-kommandon som RS232

## Viktigt

- RS232 och LAN styr bara ingångsvalet - ingen video/USB/data går över dem.
- Auto-scan måste stängas av, annars byter switchen port på egen hand.
