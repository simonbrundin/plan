---
title: U7 Lite (94:2a:6f:7a:4b:33) adopterar inte - saknar management-IP och nattverkssegment pa switchen
type: note
tags: [unifi, felsokning, incident, projekt:infrastructure]
---

Symtom (2026-11-11): accesspunkten U7 Lite (model UAPA693, MAC
94:2a:6f:7a:4b:33) syns inte i UniFi UI. Den sander ut "Krokviken WIFI"
(botha radior: 2.4 GHz 94:2A:6F:7A:4B:34, 5 GHz :35) och har stram, men
ar inte adopterad.

Historiskt (2026-10-27): AP:n var friskt - state=1, IP 10.1.1.11, switch
"US 48 PoE 500W" (10.1.1.2, _id 69416af3f21c6127cc813e1f) port 47,
1 Gbit/s, PoE 8.26 W, RX/TX-trafik, health wlan=ok. Uplink: switch port
48 -> UDRULT "Krokviken" 10.1.1.1 (_id 676b3969f5c488156c0d3ca9).
AP:n har _id 67efd07251d55d62b4865020.

Bevis fra arbetsstationen (efter associering till "Krokviken WIFI", PSK
Noomi2020):
- 4-way handshake OK, men DHCP ger INGEN lease (45 s timeout).
- En gang fick klienten 192.168.1.183/24 - UniFis default/provisioning-
  segment, inte hemnatet 10.1.1.0/24.
Tolkning: AP:n har config (SSID:er sander), men ingen fungerande
management-IP -> den naar inte controllern -> ar inte adopterad och
visas inte i UI. LED vitt blinkande = provisioning, sedan morkt.

Latenta landmine: 2026-10-26 skickade ett agentskript en PUT
rest/device med ENDAST port_overrides[{port_idx:21}] - UniFi ersatter
hela arrayen, vilket nollstaller overrides for andra portar (t.ex. 47
där AP:n satt). Kontrollera port 47s native network/VLAN + PoE i UI.

For att fa live-data maste AP:n fa controller-access: antingen fixa
Twingate-devicens resurspolicy (arbetsstationen har 0 resurser sedan
2026-11-04) eller kora /tmp/unifi-diag.sh pa en maskin i hemnatet.
