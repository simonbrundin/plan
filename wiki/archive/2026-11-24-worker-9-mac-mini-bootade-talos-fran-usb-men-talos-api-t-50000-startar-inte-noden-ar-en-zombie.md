---
title: worker-9 (Mac mini) bootade Talos fran USB men Talos-API:t (50000) startar inte - noden ar en zombie
type: note
tags: [worker-9, macmini, talos, usb, incident, felsokning, projekt:simon]
---

## Observation 2026-11-24 ca 18:00-18:45 (efter att anvandaren startat Mac minin med Talos)
Noden 10.10.10.29 (MAC a0:ce:c8:d9:0d:02, UniFi 'MinisomhorSIMON') ar UP:
- ICMP svarar (frAn worker-8 via samma L2, ARP REACHABLE).
- Reverse DNS (gateway 10.10.10.1) -> "worker-9" (= DHCP-hostname, alltsa Talos).
- Global IPv6: 2a0d:3341:a000:6841:a2ce:c8ff:fed9:0d02 (EUI-64, Talos-stil).
- MDB/ICMPv6: svarar inte pa IPv6-ping; inga mdns-svar (=> inte macOS).
- ENDAST port 10250 ar oppen (HTTPS = kubelet). Cert: CN=system:node:worker-9,
  issuer "O=kubernetes", SAN DNS:worker-9 + IP:10.10.10.29, notBefore **Mar 9 2026**
  (= installationen pa USB-sticken). Port 22 och 50000 (Talos-API) ar STANGDA.
- Full port-sweep (python asyncio, 1-65535) frAn worker-8: bara 10250.
- Kubelet avvisar bade anonym (401) och aktuell admin-kubeconfig-cert (401) -> noden
  litar pa GAMMAL cluster-CA (mars-generationen).
- worker-9 finns varken i `talosctl get members` eller i `kubectl get nodes`.

## Slutsats
Noden startade Talos frAn USB-sticken (Talos v1.12.5, gammal maskinconfig i STATE-partitionen)
men **Talos-API:t (apid) startar inte** - bara kubelet lever. Det gAr INGEN TALOS-v1.12.x
som startar kubelet utan apid normalt; maskinen ar alltsa i ett trasigt/"halvdod" tillstand.
Darmed kan vi inte applicera ny config via talosctl, och den kan inte joina klustret
(gammal PKI -> apiservern avvisar kubelet-certen).

## Varfor inte "bara reboot"
Talos-API:t har inte kommit upp under 3 minuters kontinuerlig port-monitoring.
Mogelig forklarening: gammal Talos v1.12.5 + maskinconfig frAn mars 2026.

## Plan
1. Fysiskt: strömcykla Mac minin (eller starta med Option -> "Macintosh HD" for att komma
   in i macOS). Har man macOS igang kan jag via SSH gora allt:
   - lagga Talos v1.14.1-UKI pa USB:ens ESP (EFI/Linux/Talos-v1.14.1.efi) + systemd-boot-entry,
   - `sudo bless --folder <ESP> --setBoot --label Talos` (automatisk USB-boot, ingen Option),
   - reboota -> Talos v1.14.1 med (forhoppningsvis) fungerande apid.
2. Nar apid svarar: `talosctl apply-config --file pxe/matchbox/assets/talos-config/nodes/worker-9.yaml --mode=boot`
   med det klientcert som matcher nodens generation (talosconfig-new, CA B58322..., feb 2026).
   Skript: /tmp/worker9-apply.sh (provar flera talosconfig + applicerar + rebootar + verifierar).
3. --mode=boot (inte install) sa att configen skrivs till STATE-partitionen pa samma USB
   och att den interna macOS-disken INTE roras. Install pa USB:en medan man bootar
   frAn den samma USB ar riskabelt.
4. Om apid fortfarande inte startar med v1.14.1: nasta steg ar en fysisk konsol
   (Mac mini har ingen videosignal -> macOS-boot + wipe av STATE + boot-mode-apply).

## Fakta om USB:en (for referens)
ESP (disk2s1) = systemd-boot (EFI/boot/BOOTX64.efi) + EFI/Linux/Talos-v1.12.4.efi & v1.12.5.efi,
loader/loader.conf (timeout 10). Partitioner: p3 = XFS "STATE" (100 MB), p4 = XFS "EPHEMERAL" (55,6 GB).
Klustret kor v1.14.1 -> v1.12.x pa sticken ar for gammalt for att joina.
