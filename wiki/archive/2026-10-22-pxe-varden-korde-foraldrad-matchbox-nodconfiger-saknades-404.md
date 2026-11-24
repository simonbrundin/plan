---
title: PXE-värden körde föråldrad matchbox - nodconfiger saknades (404)
type: note
tags: [pxe, incident, matchbox, projekt:simon]
---

När worker-1 skulle startas via PXE visade det sig att PXE-värden (10.10.10.2) körde en gammal matchbox-datamängd: bara de generiska profilerna (default/pi4/pi5/rock4c-plus/rock4se/worker.json), inga node-worker-N.json, inga assets/talos-config/nodes/*.yaml. Följd: /ipxe?mac=0a:3f:39:25:2b:c8 returnerade den GENERISKA worker.yaml (maintenance-config utan hostname) och nodconfigen /assets/talos-config/nodes/worker-1.yaml gav HTTP 404 - en PXE-boot hade startat noden som anonym worker, inte som worker-1. Dessutom var matchbox-assetsen (kernel/initramfs) av Talos v1.13 medan repot och talosctl är v1.14.1, och dnsmasq.conf var en gammal per-nod-version av dhcp-boot-reglerna.

Fix: rsync av pxe/matchbox/ (assets, profiles, groups) + dnsmasq.conf till /home/simon/PXE, omstart av dnsmasq via docker restart dnsmasq-proxy (containernamn, inte tjänstenamnet - tjänsten heter 'dnsmasq' i compose). Verifierat: /ipxe?mac=... pekar nu pa nodes/worker-1.yaml, nodconfigen ger 200, kernel/initramfs (Talos v1.14.1) serveras.

Viktig lärdom: nodspecifika Talos-configer ar gitignorerade (innehaller kluster-PKI) och synkas INTE automatiskt - de maste manuellt kopplas till PXE-varden varje gang generatorn kor. Repo: pxe/.gitignore forklarar detta. Ett uppdaterat matchbox utan nodconfiger ar farligare an en gammal: den gamla gav en generisk config, den nya skulle gett 404.
