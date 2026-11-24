---
title: worker-2 utan strom - port 22 up=false poe_power=0
type: note
tags: [pxe, bug, hardware, unifi, projekt:simon]
---

När worker-2 inte gick att PXE-boota visade sig att noden aldrig nade nätverket.
UniFi API (/proxy/network/api/s/default/stat/device, port_table) visade
port 22 up=false, speed=0, poe_power=0.00 W. Ingen DHCP i dnsmasq-loggarna
och ingen ping.

Metod: läs port_status via stat/device for switchen (US48P500) och se
vilka portar som har klienter via stat/sta. Ett par PoE-cykler (off -> auto
pa port_idx via PUT /rest/device/<id> med port_overrides) hjalte inte -
dod maste arsaken ligga i strumforsorjningen eller kabeln.

Viktig lardom: felsok PXE-floden BOST verifiera tre lager var for sig:
1. har noden strom och nattverkslan (UniFi port-status),
2. går DHCP och TFTP (dnsmasq-logg: DHCPACK + 'sent <fil>'),
3. får noden rätt config och kernel (U-Boot-konsol, 'Booting ...',
   'Not a PE-COFF file' = fel argumentformat till bootefi).

Utan steg 1 kan manfelsoka fel i PXE-konfigurationen i timmar.
