---
title: UniFi-gatewayen har ingen publik IPv6 (CGNAT via Starlink)
type: note
tags: [unifi, ipv6, cgNAT, starlink, wan, projekt:infrastructure]
created: 2026-11-15
updated: 2026-11-17
status: active
related: [notes/ipv6-ingress-till-envoy-designlage-2026-11-15]
---

## Slutsats (2026-11-15)

Starlink **ger** IPv6 på alla planer, men bara /64 på WAN via SLAAC + /56 via DHCPv6-PD.
UniFi + Starlink i bypass-läge fungerar **bara** med UniFi:s *Single Network*-läge.

## Root cause (varför det inte fungerade från början)

WAN-nätet "Internet 1" hade `wan_type_v6: dhcpv6`. Starlink levererar ingen DHCPv6 IA_NA
(stateful adress) - bara SLAAC (/64 i RA) och IA_PD (/56). UniFi med `dhcpv6` frågar efter
IA_NA och tog därför bara `fe80::` på WAN. LAN-näten hade dessutom
`ipv6_interface_type: none`, så ingen delegerad prefix applicerades.

## UniFi-begränsningar som upptäcktes (viktigt)

- LAN **Prefix Delegation kräver att WAN sättts till DHCPv6**
  (`api.err.PdRequiresAssignedDhcpv6Wan` annars).
- Med `dhcpv6` på WAN fick vi **ingen** global adress (Starlink ger ingen IA_NA), och efter
  90 s väntan kom **ingen** IA_PD - Starlink verkar inte leverera /56-PD till oss.
- Lösningen som fungerar: WAN `wan_type_v6: slaac` +
  `ipv6_wan_delegation_type: single_network` + `single_network_lan: <LAN-id>`.
  Då sätter controllern LAN-nätets `ipv6_interface_type` till `single_network` och delar
  WAN:s /64 till klienterna via SLAAC.
- ** Bara EN LAN kan ha IPv6 med single_network.** Satt `single_network_lan` till Default
  (10.1.1.0/24) idag; alternativt Datacenter (10.10.10.0/24) - testat och verifierat.
- `ipv6_interface_type: static` accepteras också men blir **statiskt**: om Starlink byter
  prefix (/64) så går LAN-sidan sönder. single_network hanterar det automatiskt.

## Slutlig konfiguration (verifierad)

- Internet 1: `wan_type_v6: slaac`, `ipv6_wan_delegation_type: single_network`,
  `single_network_lan: ec2a295b-5a27-4c3d-9148-f52a0b925ebc` (Default),
  `wan_dhcpv6_pd_size: 56`, `ipv6_setting_preference: manual`
- Default: `ipv6_interface_type: single_network` | IoT & Datacenter: `none`

Verifierat via `stat/device`:
- WAN-interface: `2a0d:3341:a000:6841:eea:14ff:fe56:90e9` (global) + link-local
- Gateway LAN-interface: `2a0d:3341:a000:6841:cea:14ff:fe56:90e6/64`
- Switch 10.1.1.2: `2a0d:3341:a000:6841:1ae8:29ff:fe2e:f093/64`
- AP 10.1.1.11: `2a0d:3341:a000:6841:962a:6fff:fe7a:4b33/64`

**End-to-end-test** (Talos-nod 10.10.10.10, IPv6 aktiverat på Datacenter under testet):
`ping -6 2606:4700:4700::1111` = 0 % förlust och `https://api64.ipify.org` svarade med
nodens egen GUA `2a0d:3341:a000:6841:e65f:1ff:fea0:81f7`. IPv4 oförändrat
(212.105.145.11, CGNAT WAN-IP 100.67.39.22/18).

## Övrig kontext

- IPv4 är CGNAT (Starlink 100.64.0.0/10) -> ingen publik IPv4. Publik IPv4 kräver
  Local/Global Priority-plan.
- Gatewayens WAN-GUA finns kvar oavsett vilken LAN som har single_network.
- Inkommande IPv6 blockeras av UniFis zonbrandvägg (established/related) - öppna portar
  kräver explicita firewall-regler.
- Arbetsstationen (omarchy) sitter på en **annan site**: Netgear R7800 (192.168.1.1) +
  Starlink, och når 10.1.1.0/24 via Tailscale/overlay (`sdwan0`). Netgars IPv6 är inte
  aktiverat - slå på "Internet IPv6" i Netgars gränssnitt för IPv6 där också.

## Relaterade

- [[notes/ipv6-ingress-till-envoy-designlage-2026-11-15]]

