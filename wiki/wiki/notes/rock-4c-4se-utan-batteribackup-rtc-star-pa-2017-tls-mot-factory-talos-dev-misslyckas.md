---
title: Rock 4C+/4SE utan batteribackup: RTC står på 2017, TLS mot factory.talos.dev misslyckas
type: note
tags: [pxe, talos, rtc, tls, ntp, image-factory, projekt:infrastructure]
created: 2026-10-24
updated: 2026-10-24
status: active
---

Symtom: image-pullen går aldrig igenom på ROCK-noderna:
  talos: retrying error: failed to pull image
    "factory.talos.dev/metal-installer/...:v1.14.1": failed to resolve reference:
    Head "https://factory.talos.dev/v2/...": tls: failed to verify certificate:
    x509: certificate has expired or is not yet valid:
    current time 2017-08-05T09:01:09Z is before 2026-10-10T06:12:41Z"
Talos startar med RTC-klockan pa 2017-08-05 (ingen batteribackup pa braden), sa
TLS-valideringen mot Image Factory misslyckas. Talos har en NTS-fallback som
gor en "time jump" efter ~30 s, men den kraver DNS-uppslag mot
time.cloudflare.com - och DNS mot router (10.10.10.1) gor timeout.

FIX (nu i varje nodconfig, genereras av pxe/generate-node-configs.py):
  apiVersion: v1alpha1
  kind: TimeSyncConfig
  bootTimeout: 5m
  ntp:
    servers: [162.159.200.1, 216.239.35.0]   # Cloudflare, Google
    useNTS: false
IP-baserad NTP utan NTS fungerar utan DNS och utan certifikatvalidering.
Obs: bootTimeout satter en grans - utan den vantar noden pa tidsynk fore
boot-sekvensen (och da configen inte hamtas alls).

VARNING: bootTimeout far INTE vara for kort. Nodens forsta NTP-forsok sker
innan natverket ar helt igang ("dial udp ...: connect: network is
unreachable"), sa tiden synkas tidigast ~30 s efter boot. Ga noden vidare med
fort fort fel tid far den ogiltiga kubelet-certifikat:
  kubelet client certificate does not match any accepted CAs, removing
  x509: current time 1970-01-01T00:00:15Z is before 2026-10-24T...
och noden hamnar som NotReady i klustret. 5m ger tid for synk utan att
tappen hanga.

Verifierat pa controlplane-3 (10.10.10.13):
  adjusting time (jump) by 498016h7m20s via 162.159.200.1, state TIME_OK
NTP-serverna svarar via UDP 123 ut mot internet - det "unreachable" i loggen
ar bara initfasen, inte en brandvagg.

Varfor detta mat: felen ser ut som natverksproblem men ar egentligen
klockproblem. Kolla tiden i loggen innan du skyller pa brandvagg.
