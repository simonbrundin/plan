---
title: Talos: noden kan hamna i maintenance mode om STATE-volymen hittas efter config-fasen
type: note
tags: [talos, incident, controlplane, boot, projekt:infrastructure]
created: 2026-10-25
updated: 2026-10-25
status: active
---

Symtom: controlplane-3 gick in i maintenance mode direkt efter en reboot, trots att noden redan hadde en komplett config persistad pa STATE och hade bott korrekt flera ganger.

Root cause: en race i Talos v1.14.1 mellan config-acquisition och disk-volymdetektering. Loggen visar exakt detta:

  00:00:07.407 volume status STATE phase 'waiting -> missing'
  00:00:07.412 phase config (8/8): 1 tasks(s)
  00:00:07.414 downloading config (platform=metal)
  00:00:07.415 entering maintenance service
  00:00:08.048 volume status STATE phase 'missing -> ready' location /dev/sda5

AcquireController (internal/app/machined/pkg/controllers/config/acquire.go) börjar i stateDisk. Om VolumeStatus for STATE da fortfarande har fasen 'missing' (volymen har inte hunnit lokaliseras pa disken) tolkar den det som 'det finns ingen installerad Talos', gaer vidare till embedded/cmdline/platform och hamnar i maintenance. Nar disken sedan lokaliseras ar det for sent - state maskinen har redan lamnat stateDisk.

Varfor det drabbar just den har noden: systemdisken ar en USB-SSD pa Raspberry Pi 4 (Intenso, sda). Disk-enumereringen ar langsammare an pa inbyggd flash, sa fastern 'missing -> ready' infaller efter config-fasen. En nod som bootar fran inbyggd eMMC (controlplane-2) har samma risk men har i praktiken inte traffat den.

Fix i drift: applicera nodens nodspecifika config via maintenance-API:
  talosctl apply-config -f pxe/matchbox/assets/talos-config/nodes/controlplane-3.yaml -e <ip> --insecure -n <ip> -m auto
Detta persisteras till STATE ('machine configuration persisted to STATE') och nar booten sedan lyckas hitta STATE tidigt far noden configen darifran: 'machine config loaded successfully sources: [state]'.

Lardomar:
- 'entering maintenance service' i loggen betyder INTE att configen ar trasig - det kan vara enbart en race med volymdetekteringen. Kolla alltid om det finns en config pa STATE.
- En nod utan talos.config i /proc/cmdline (bootar fran egen disk) har ingen backup-kalla om STATE-racet intrafar. En nod som bootar via PXE med talos.config=URL klarar alltid ut.
- Kolla '/proc/cmdline' for att avgora vilken config-vag noden anvander. controlplane-1 bootar via PXE med talos.config, controlplane-2 och controlplane-3 bootar fran egen installerad disk och far helt pa STATE.
- Matchbox-profilen for controlplane-3 pekar nu pa /assets/talos-config/nodes/controlplane-3.yaml (nodspecifik) i stallet for den generiska controlplane.yaml som har HostnameConfig.auto.
