---
title: UniFi-gatewayen har ingen publik IPv6 (CGNAT via Starlink)
type: note
tags: [unifi, ipv6, cgNAT, starlink, wan, projekt:infrastructure]
---

Kontrollerad via UniFi-API (10.1.1.1, stat/device + rest/networkconf) 2026-11-XX: UDM 'Krokviken' har ENDAST link-local IPv6 pa WAN/LAN (fe80::cea:14ff:fe56:90e6/64) - ingen global/publik IPv6. WAN: Internet 1 har wan_type_v6=dhcpv6 + ipv6_wan_delegation_type=pd konfigurerat, men upstream (Starlink-router pa 100.64.0.1, MAC 26:12:ac:1a:80:01) delegerar ingen prefix. IPv4 är CGNAT: gatewayns WAN-IP är 100.67.39.22/18 och klienter syns ut som 31.209.42.41. Slutsats: ingen publik IPv4 eller IPv6 mot internet; IPv6-tjänster kräver extern tunnel (Tailscale/Tw ingate) eller byte av ISP.
