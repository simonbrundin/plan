---
title: worker-8: korrupt XFS pa EPHEMERAL-partitionen gjorde att kubelet aldrig startade - fix med mkfs.xfs
type: note
tags: [worker-8, talos, xfs, felsokning, pxe, projekt:infrastructure]
---

Symtom (2026-11-23): worker-8 (10.10.10.28) var medlem i Talos-klustret (talosctl get members) och svarade pa Talos-API:t, men syntes inte i `kubectl get nodes`. Pa noden: service "cri" = Waiting i 3h med "Waiting for volumes /var/lib, /var/lib/cni, CRI, /var/run, /var/run/lock, /opt to be mounted" - kubelet startade aldrig.

Rotorsak: EPHEMERAL-partitionen (/dev/mmcblk1p6, XFS) var korrupt. Talos-loggen: "volume status EPHEMERAL phase waiting -> failed, error growing filesystem: error growing XFS: FSCONFIG_CMD_CREATE failed: structure needs cleaning". xfs_repair -n visade "bad primary superblock - bad magic number" (forsta 512 byte nollade) och "would reset bad agf/agi for ag 0" - dvs utbockad metadata, inte bara en smutsig logg. /var monterades aldrig, sa CRI/kubelet blockerade.

Fix (gick pa ~20 min, utan PXE och utan fysisk atkomst):
1. Hitta rätt admin-kubeconfig: `talosctl --talosconfig ~/.talos/config --nodes 10.10.10.11 kubeconfig /tmp/kc.yaml` (talos/talosconfig i repot har tomma certifikat -> "tls: failed to find any PEM data"; ~/.kube/config har flera kontexter, admin@cluster1-2..6 fungerar, admin@cluster1 har en gammal CA).
2. Skaffa xfsprogs pa noden: `talosctl -n <nod> debug docker.io/library/alpine:latest --args /bin/sh --args -c --args "<script>"`. Noden kan inte na alpinelinux.org normalt (busybox wget far IPv6 och timeout:ar) - men TCP over IPv4 fungerar. Losningen var `printf 'GET /alpine/v3.21/main/aarch64/<fil>.apk HTTP/1.1\r\nHost: dl-cdn.alpinelinux.org\r\n\r\n' | nc 151.101.66.132 80` och sedan strippa HTTP-headern med `tail -c +$((total - storlek + 1))` (busybox grep saknar -b, sa byte-offset via grep gick inte). Nod-paket: xfsprogs + libblkid + libuuid + userspace-rcu + inih + libeconf.
3. Kora xfs_repair -d (disable O_DIRECT - annars "Permission denied" pa O_DIRECT-oppna av blockdevicen i debug-containern). Resultat: FS var for stort skadad (agf/agi nollade) -> repair raddade inte allt.
4. Därfor: `mkfs.xfs -f -L EPHEMERAL /dev/mmcblk1p6` (samma fs-label som worker-7 har: LABEL="EPHEMERAL", UUID slumpad). /var-innehallet pa worker-8 gick forlorat - noden hade aldrig varit i klustret sa inget viktigt forsvann.
5. `talosctl -n 10.10.10.28 reboot` -> Talos monterar EPHEMERAL, cri+kubelet startar, och worker-8 blev Ready i klustret pa ~2 min.

Viktiga lardomar:
- Ett Talos-klustermedlem (talosctl get members) kan vara helt OK i Talos-lagret men saknas i Kubernetes - kolla ALLTID "talosctl services" (om "cri" star Waiting pa volumes) och inte bara "get members".
- "structure needs cleaning" pa en XFS betyder ofta att utbockad metadata, inte bara logg - da hjaller xfs_repair inte, man maste mkfs:a om (och da vet man att /var-innehallet gick forlorat).
- xfs_repair/xfs_info med O_DIRECT pa blockdevicer i en Talos debug-container ger "Permission denied" -> anvand -d.
- Busybox-verktyg pa Talos-noder: grep har inte -b, ash har $'...'-quoting? (anvand base64-skript for att undvika citattecken), /etc/hosts och /etc/resolv.conf ar read-only i debug-containern.
