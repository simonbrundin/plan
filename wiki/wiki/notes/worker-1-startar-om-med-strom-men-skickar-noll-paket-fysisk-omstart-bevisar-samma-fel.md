---
title: worker-1 startar om med stram men skickar noll paket - fysisk omstart bevisar samma fel
type: note
summary: Status 2026-10-27: en fysisk omstart av worker-1 och en kontrollerad PoE-cykel med 1 Hz-strommalning visar att boarden far strom, forhandlar PHY-link pa 1 Gbit/s men aldrig skickar ett enda eget paket - DHCP-rader saknas sedan 2026-10-23...
tags: [pxe, hardware, felsokning, rock4cplus, unifi, projekt:infrastructure]
created: 2026-10-27
updated: 2026-10-27
status: active
related: [entities/pxe-server, notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande, notes/worker-1-svarar-inte-ens-med-en-full-poe-cykel-noll-paket-i-bade-riktningar]
---

Status 2026-10-27: bade en **fysisk omstart** av worker-1 och en
**kontrollerad PoE-cykel med 1 Hz-strommalning** ger samma resultat: boarden
far strom och forhandlar PHY-link pa 1 Gbit/s, men skickar **noll** egna
paket. Felet ar alltsa inte PXE-kedjan och inte en gang som far en stale
lease - noden har inte fragat om DHCP sedan **2026-10-23 12:13 UTC**.

## Kalla fakta fran bestandiga loggar

Loggarna pa PXE-varden roterar, sa historiken lases ur de arkiverade
filerna i `/home/simon/PXE/logs/` (startade 2026-10-27 23:16):

- `dnsmasq.log` tacker **2026-10-19 09:18** -> nu.
- worker-1 (`0a:3f:39:25:2b:c8`) har **1168 DHCP-rader**, men den sista ar
  `2026-10-23T12:13:44Z DHCPACK ... worker-1`.
- **Noll** DHCP-rader med worker-1:s MAC efter 2026-10-25.
- `matchbox.log`: sista riktiga iPXE-fraega for worker-1:s MAC ar
  `2026-10-22T16:40:48Z`. (Raden `2026-10-27T14:00:33Z` ar en manuell
  `curl`-test fran laptopen, inte noden.)
- Andra workers forsatter handla lease normalt: `6a:4e:e8:63:ad:5c`
  (worker-3), `66:e1:e4:58:64:1b` (worker-7), `ee:26:6d:d4:b3:ed` (worker-5),
  `a2:83:f9:db:93:13` (worker-4) och `e4:5f:01:a0:80:83` (controlplane-1).
  Ett tcpdump i VLAN 2 (controlplane-1, iface `enabcm6e4ei0`) visar deras
  DHCP Discover/Offer/Request/ACK i klartext - **men nagon gang fran
  worker-1**.

## Stromkurvan under en kontrollerad PoE-cykel (1 Hz)

Port 21, `poe_mode=off` -> 70 s -> `poe_mode=auto`:

```
22:50:26  pwr=1.93  up=true   spd=1000  rx=0  mac=0  ld=7
22:50:27  pwr=0.00  up=false  spd=0     rx=0  mac=0  ld=8   <- porten ner
22:51:58  pwr=0.00  up=false  spd=0     rx=0  mac=0  ld=8
22:51:59  pwr=1.85  up=true   spd=1000  rx=0  mac=0  ld=8   <- tillbaka
```

Tre saker galler samtidigt:

1. **Strommen forsvinner helt** - porten matte verkligen kopplas ner.
2. **Strommen kommer tillbaka pa 1.85 W**, aldrig over 2 W.
3. **Noll rx-paket och noll MAC-inlarning** hela tiden, fore och efter.

Jmför portstrom for noderna som faktiskt bootar:

| Port | Nod | pwr | rx_packets | conn |
|------|-----|-----|-----------|------|
| 21 | worker-1 | 1.77-1.93 W | 0 | false |
| 22 | worker-2 | 2.08 W | 0 | false |
| 23 | worker-3 | 4.33 W | 8.7 M | true |
| 24 | worker-4 | 4.31 W | 8.5 M | true |
| 25 | worker-5 | 4.24 W | 8.8 M | true |
| 26 | worker-6 | 3.72 W | 116 k | true |
| 27 | worker-7 | 3.68 W | 9.9 M | true |
| 28 | worker-8 | 3.56 W | 116 k | true |
| 11 | controlplane-1 | 4.87 W | 9.5 M | true |

~1.8 W ar **hjelften** av vad en korrekt bootande ROCK drar. Boarden far
strom till PHY:n och handshakar pa 1 Gbit/s, men SoC:en gor inget arbete
alls. Ett U-Boot som korar `boot_targets=pxe dhcp ...` skulle dra DHCP inom
sekunder - det gor det inte.

## Varfor detta inte ar ett DHCP- eller PXE-problem

- dnsmasq och matchbox ar **samma containrar som serverar de fungerande
  noderna**. De fungerar.
- Switchens broadcast når port 21 i samma omfattning som port 23, sa VLAN
  2-ar flodet intet hinder.
- Gatewayn fragar aktivt ARP for `10.10.10.21` - noden svarar aldrig.
- Nodens MAC ar inte i UniFis klienttabell overhuvudtaget (`stat/sta` saknar
  port 21 och 22 helt), medan port 23-28 och 11-13 finns med korrekt
  `uptime`.

## Slutsats

worker-1 och worker-2 ar **doda i SoC-nivan**. Boarden bootar inte till
U-Boot-prompt, eller U-Boot bootar men nagon blockerar
`pxe dhcp`-steget (t.ex. `boot_script_dhcp` saknas, eller `saveenv` har
sparat en miljo vars `boot_targets` borjar med `mmc`). Inget programvariskt
nätverksverktyg kan felsöka detta - det kravs **fysisk U-Boot-konsol over
serie**.

Bekraftigat medvetet: `u-boot-rockchip-pxe-first.bin` i repot har ratt
`boot_targets=pxe dhcp mmc1 mmc0 nvme scsi usb spi` och
`boot_script_dhcp=talos/bsc`, medan **upstream**-bloben
`rock4c-plus.bin` har `mmc1 mmc0 nvme scsi usb pxe dhcp spi` - dvs
**eMMC fore PXE**. Om just worker-1:s kort har upstream-bloben eller en
sparad env med `mmc` forst kommer boarden att forsoka boota fran disk och
aldrig na PXE.

## Relaterade

- [[entities/pxe-server]]
- [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]]
- [[notes/worker-1-svarar-inte-ens-med-en-full-poe-cykel-noll-paket-i-bade-riktningar]]
