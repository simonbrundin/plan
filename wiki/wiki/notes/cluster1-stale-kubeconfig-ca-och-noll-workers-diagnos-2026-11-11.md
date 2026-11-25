---
title: cluster1: stale kubeconfig-CA och noll workers - diagnos 2026-11-11
type: note
tags: [felsokning, pki, worker, talos, kubeconfig, projekt:infrastructure]
created: 2026-11-13
updated: 2026-11-23
status: active
related: [notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device, notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs]
---

Diagnos 2026-11-11 (session: felsökning av 'workers ansluter inte').

## UPPDATERING 2026-11-23 - "noll workers" gäller INTE längre
- Klustret har nu 6 workers i Kubernetes: worker-3, worker-4, worker-5, worker-6, worker-7 och
  worker-8 (alla Ready). worker-1, worker-2 och worker-9 saknas fortfarande.
- worker-8 kom in via en helt annan(or) route an "recover-worker.sh": noden var medlem i Talos-lagret
  sedan tidigare men EPHEMERAL-partitionens XFS var korrupt, sa /var monterades aldrig och kubelet
  startade aldrig. Fix: mkfs.xfs pa /dev/mmcblk1p6 + reboot.
  Se notes/worker-8-korrupt-xfs-pa-ephemeral-partitionen-gjorde-att-kubelet-aldrig-startade-fix-med-mkfs-xfs.
- NYTT METODKRAV: "talosctl get members" visar bara Talos-lagret. En nod kan vara medlem dar och
  samtidigt saknas i Kubernetes - kontrollera "talosctl -n <nod> services" (om "cri" star Waiting
  pa volumes) och "kubectl get nodes".
- NYTT ADMIN-KUBECONFIG: repot talos/talosconfig har tomma certifikat (kan inte anvandas).
  ~/.talos/config fungerar for Talos-API:t, och en fungerande admin-kubeconfig far man med
  "talosctl --talosconfig ~/.talos/config --nodes 10.10.10.11 kubeconfig <fil>".
  I ~/.kube/config fanns 2026-11-25 kontexterna admin@cluster1-2 ... admin@cluster1-6 (samt -1-1);
  admin@cluster1 hade fortfarande en for gammal CA. UPPDATERAT 2026-11-25: alla dubbletter ar
  borttagna och admin@cluster1 ar nu den enda giltiga admin-kontexten (VIP 10.10.10.10 + aktuell
  CA). Se [[notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna]].
  Varfor: talosctl kubeconfig --merge lagger in nya kontexter varje gang - anvand
  "talosctl kubeconfig /tmp/kc.yaml --merge=false" istallet.

## 1. Två helt separata fel - blanda inte ihop dem

A) KLIENTEN (kubeconfig) - det som gav 'x509: certificate signed by unknown authority'.
   - ~/.kube/config kontexten admin@cluster1 pekar pa server 10.10.10.12:6443 med en CA fran 2026-10-14 (sha256 CA:5C:EB:...).
   - KLUSTRET anvander CA 2026-10-22 (sha256 7B:3F:A5:9A:4F:69:E7:20:28:1A:AF:79:C7:56:86:C4:56:7B:3F:F1:09:29:C0:E1:16:2F:66:C0:A1:09:5C:DA).
   - Apiservernas servercertifikat pa alla tre controlplanes ar utstalda 2026-11-11 (10:54 / 11:22 / 11:24) - dvs CA:n roterades om.
   - LOSNING: 'talosctl --nodes 10.10.10.11 kubeconfig' (skriver in i ~/.kube/config). Skapade kontexten admin@cluster1-4, som ar den nuvarande current-context och fungerar. (2026-11-25: de -N-kontexterna ar borttagna, admin@cluster1 anvands nu.)
   - VARFORS: talosctl kubeconfig mergar in i befintlig ~/.kube/config och byter namn pa krockande 'cluster1' -> 'cluster1-2/-3/-4'. Det var darfor jag fick tre extra kontexter; -2 och -3 ar nu borttagna.

B) KLUSTRET HAR NOLL WORKERS - det faktiska problemet.
   - kubectl get nodes: endast controlplane-1/2/3, alla Ready. Inga worker-noder.
   - Ingen av de 9 worker-IP:n (10.10.10.21-29) svarar pa Talos-API:t (port 50000). Worker-1/2 svarar inte ens pa ping. 23-27 svarar pa ping men saknar alla tjanster (22/80/443/50000/6443 stanga) - de kor INTE Talos.
   - PXE-varden 10.10.10.2 var NERE vid denna diagnos: ping fejar, 69/80/8080 stanga, nere sedan 2026-10-26 (se tidigare noter).
     UPPDATERING 2026-11-13: varden ar IGANG igen och serverkedaan ar helt verifierad fungerande - men noderna startar fortfarande inte via PXE. Se notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device. (Varden ar en Radxa-board, MAC 2c:cf:67:43:f3:b0, UniFi-port 8 - inte en Raspberry Pi.)
   - KONFIGEN AR INTE PROBLEMET: pxe/matchbox/assets/talos-config/nodes/worker-3.yaml har clusterID c+KQTosqRpwxg5j2FXMinAGKd17qB5l1Ebb4ViWNeTE=, clusterSecret och endpoint https://10.10.10.10:6443 - exakt samma som live-klustret - och acceptedCAs-matchar live-CA:n (7B:3F:A5...) bit for bit. Alla PXE-worker-configs i repot är alltsa uppdaterade mot nuvarande PKI.
   - FOLJD: nar PXE-varden kommer tillbaka kan 'pxe/recover-worker.sh <nod> --yes' anvandas for att installera om noderna; de ska da kunna joina.

## 2. Cluster1 ar i grunden tomt
   - Inga Flux-CRD:ar (no fluxinstance/helmrelease), inga namespaces utover default/kube-system/kube-node-lease/kube-public, inga storageclasses, inga appar.
   - etcd: 3 medlemmar (50b0f7c3297c1e4e cp-1, 83d29f633552cb78 cp-2, 199a9f61a6dc6e1b cp-3), leader cp-1, raft index 4.62M, db 57 MB - sundt.
   - Innebod: den nuvarande cluster1-generationen ar en ny/rebootstrappad generation (2026-10-22) som aldrig fatt tillbaka sina workers eller sin GitOps-installation.

## 3. Metodregler (viktiga)
   - JAMFOR aldrig machine.ca.crt (O=talos) mot Kubernetes-CA (O=kubernetes) - de ar separata hierarkier. Anvand openssl -fingerprint pa rätt fild.
   - 'talosctl get kubeconfig' finns INTE i v1.14 ('resource kubeconfig is not registered') - anvand 'talosctl kubeconfig'.
   - 'talosctl etcd status' med flera -n/-e fungerar; 'talosctl get members' visar INTE workers (bara controlplane-medlemmar).
   - UniFi-API:t pa 10.1.1.1: 'proxy/network/api/s/default/stat/device' fungerar, men 'stat/portup'/'stat/port' svarar api.err.NotFound - portstat/PoE-status kan inte lasas via det API:t.
     UPPDATERING 2026-11-13: PoE-styrning per port fungerar via PUT '/proxy/network/api/s/default/rest/device/{device_id}' med {"port_overrides":[{"port_idx":N,"poe_mode":"off"|"auto"}]} (HTTP 200). 'stat/device' (listan) innehaller dessutom hela port_table med speed/poe_good/last_connection. 'cmd/devmgr power-cycle' svarar rc=ok men gor INGENTING - falsk framgang.

## 4. Per-nod bootmetod - INTE alla workers bootar via PXE (viktig korrektion)

| Nod | Hårdvara | Bootväg | PXE nödvändigt? | Läge 2026-11-11 |
|---|---|---|---|---|
| controlplane-1..3 | RPi4 | egen NVMe via EDK2/boot.img pa SD | Nej (fungerar) | Ready |
| worker-1..3 | ROCK 4C+ | PXE-ONLY | Ja | .21/.22 svarar inte pa ping (dod hvara/PoE, se tidigare noter), .23 har IP men ingen Talos |
| worker-4..8 | ROCK 4SE | PXE-ONLY | Ja | .24-.27 har IP men ingen Talos, .28 svarar inte pa ping |
| worker-9 | Mac mini/MacBook | egen Talos-partition, MANUELLT val i iBoot (manual-boot: true) | Nej - PXE anvands inte i praktiken (x86-64-iPXE saknas i tftp/) | .29 svarar inte pa ping (avstangd/slagen) |

Varfor ROCK 1..8 ar PXE-only: Talos kan inte installeras till disk pa 4C+/4SE - U-Boot saknar EFI-tjansten SetVariableRT, sa installBootloader() misslyckas, META forblir tom och STATE går inte att montera. Noderna maste alltsa boota fran natet varje gang. Raspberry Pi 4 klarar det (EDK2 har SetVariableRT).

Konsekvens for recovery:
- worker-1..8: KRAVS att PXE-varden 10.10.10.2 (Raspberry Pi) ar igang. recover-worker.sh wipe:ar disken, applicerar install-configen och verifierar medlemskap.
- worker-9: PXE hjelper inte. Noden maste startas manuellt (iBoot -> Talos-partition), sedan taleser noden gamla PKI sa den maste wipe:as (talosctl reset --wipe-mode=system-disk) och installeras med nuvarande config, vilket kan goras over talosctl utan PXE.
- PXE-varden ar en permanent single point of failure for 8 av 9 workers - den maste vara igang hela tiden, inte bara vid recovery.

Verifiering 2026-11-11: full portscann av 10.10.10.0/24 (22, 67, 69, 80, 443, 50000, 6443, 8006, 8080) gav ENDAST 10.10.10.1 (gateway), 10.10.10.10-13 (controlplanes, 50000 OPEN). Ingen matchbox/dnsmasq nagonstans. 10.10.10.1:8080 ar en UniFi-tjanst (Java/Tomcat, /boot.ipxe -> 404), inte matchbox.

Alla tre konfigurationsfixar ar redan i de genererade nodconfigarna (ImageCacheConfig local.enabled=false, TimeSyncConfig useNTS=false + bootTimeout 5m, UnattendedInstallConfig med diskSelector pa FAKTISK disk + reboot=false i maintenance-configen) - se pxe/README.md "Tre konfigurationsfixer".

## Relaterade

- [[notes/workers-bootar-inte-via-pxe-trots-fungerande-pxe-server-unifi-poe-styrning-via-rest-device]]

