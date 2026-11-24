---
title: Twingate-ressurserna forsvann pa arbetsstationen 2026-11-04 - hemnatet (10.1.1.1) och datacenter (10.10.10.0/24) oppnarbara
type: note
tags: [incident, twingate, unifi, projekt:infrastructure]
---

Symtom: fran arbetsstationen (omarchy) nar inte 10.1.1.1 (UniFi-controller),
192.168.1.1, 10.10.10.0/24 (pxe/datacenter) eller mooni (10.1.1.50/150).
Twingate raporterar "online" men `twingate resources` ger "No resources
currently available".

Rotorsak: loggen visar att resources_total hoppade fran 4 till 0 vid
2026-11-04 12:21:22, precis efter en naterverksbyte ("Switching bypass
interface from wlp0s20f3", "failed to get SD: aborting due to network change").
Efter det fick enheten en ny naterverksidentitet (naretverk 248115 -> 187582),
och den nya SD:n innehaller policies=[] och resources=[]. Enheten heter nu
"Unknown device" i portalen. Ingen auth-fraga - alla TCP-anrop till
10.1.1.1/10.10.10.x fylls pa timeout (tunneln finns men ingen connector
svarar).

Konsekvens: ingen UniFi-API (AP/port/PoE-status), ingen talosctl, ingen
pxe-loggar fran arbetsstationen. Stugan (78.69.91.181, ssh skogsduvan) och
internet fungerar, sa felsokningen far ske dar.

Fix: logga in i Twingate-portalen (simonbrundin.twingate.com) och koppla
denna enhet ("Unknown device") till resurserna igen - antingen via
device profile eller ACL:policyn for 10.1.1.0/24, 10.10.10.0/24 och
mooni/10.1.1.50/150. Resurserna finns i naterverket men enheten har ingen
policy.
