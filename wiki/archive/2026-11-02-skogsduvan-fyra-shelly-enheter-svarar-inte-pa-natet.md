---
title: Skogsduvan: fyra Shelly-enheter svarar inte på nätet
type: note
tags: [home-assistant, skogsduvan, hardware, session, projekt:simon]
---

Diagnos 2026-11-02: HA-loggen säger 'Config entry X for shelly integration not ready yet: Device communication error' for FYRA shelly-config entries, och samma fyra svarar inte pa ping fran add-on-containern (samma LAN som de som fungerar). Doda ar de som INTE svarar: '4372' Flakt 192.168.189.106, Luftvarmepump 192.168.189.202, Element 192.168.189.131, Varmvattenberedare 192.168.189.234. Fungerande: Golvvarme .229 och Handdukstork .137 (http 200 pa /status).

VARFOR det spelar roll: sensor.element_temperatur (Shelly Element) ar medlem i temperaturgruppen 'Hemmet', sa med dodheten blir binary_sensor 'Alla rum over 5 grader' unavailable (fail-closed) och aldrig 'on'. Dessutom finns de Shelly-enheterna kvar i .storage/core.config_entries och loggar en retry var 10-40 s i loggen.

MISTANKBAR TOLKNING (viktig): 'state saknas i recordern' betyder INTE att enheten ar dod. Recordern skriver bara en ny rad nar statet ANDRAS - en Shelly-temperatursensor som hela tiden lamnar samma varde (t.ex. 21.0) skriver ingen ny states-rad alls, bara statistics. Hall korrekt for att 'Aldrig rapporterat' egentligen betyder 'andrat aldrig'. Anvand 'unavailable' i state-tabellen for att se dodheten, inte 'saknar rad'.

Alla fyra ar natskopplade (230V) - sa det ar troligen Wi-Fi/AP-sida eller en gemensam stramkast/power-strip, inte tomma batterier. Behover fysisk tillsyn i stugan. Zha-batteribeacons (Utbyggnad, Tvastall) ar en separat fraga: de svarar inte pa ZHA sedan 2026-10-05 / 2026-11-01.
