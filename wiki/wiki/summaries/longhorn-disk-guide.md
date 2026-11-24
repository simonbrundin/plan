---
title: Sammanfattning: Diskar i Longhorn
type: summary
summary: `docs/add-disk-to-longhorn.md` — Komplett guide med arkitektur, steg-för-steg, och felsökning.
tags: [longhorn, storage, talos, kubernetes]
sources: docs/add-disk-to-longhorn.md
created: 2026-06-02
updated: 2026-10-24
---

# Sammanfattning: Lägga till diskar i Longhorn via Talos

## Källa

`docs/add-disk-to-longhorn.md` — Komplett guide med arkitektur, steg-för-steg, och felsökning.

## Huvudpunkter

### Två-lager arkitektur

| Lager | Ansvar |
|-------|--------|
| **Talos** | Formatera, partitionera, montera på `/var/mnt/<namn>` |
| **Longhorn** | Storage management, replicas, snapshots |

Talos är det underliggande lagret — det måste finnas en giltig sökväg innan Longhorn kan använda disken.

### Diskkonfiguration i Talos

```
talos/patches/workers/<NODNAMN>/
├── patch.yaml                    # machine.type, install.diskSelector
└── disks/
    └── ssd-1.yaml               # UserVolumeConfig
```

**UserVolumeConfig kräver:**
```yaml
apiVersion: v1alpha1
kind: UserVolumeConfig
name: ssd-1
volumeType: disk                    # OBLIGATORISKT
provisioning:
  diskSelector:
    match: "'/dev/disk/by-id/usb-...' in disk.symlinks"
```

### Diskkonfiguration i Longhorn

```yaml
apiVersion: longhorn.io/v1beta2
kind: Node
metadata:
  name: <NODNAMN>
  namespace: longhorn-system
spec:
  allowScheduling: true
  disks:
    ssd-1:
      path: /var/mnt/ssd-1
      allowScheduling: true
      evictionRequested: false
```

### Flöde: Fysisk disk → Longhorn volym

```
1. Disk kopplas in → Talos upptäcker via USB
2. UserVolumeConfig matcher via symlinks
3. Talos formaterar och monterar på /var/mnt/ssd-1
4. Longhorn Node CRD pekar ut sökvägen
5. Longhorn kan använda disken för replicas
```

## Kritiska lärdomar

| Problem | Lösning |
|---------|---------|
| Fel disk monterad (mmcblk1p6 vs USB) | Använd disk-by-id, wipea och rebootstrap |
| Filesystem mismatch (GPT vs XFS) | `talosctl reset --wipe-mode=user-disks` |
| UUID mismatch efter Talos-reinstall | Ta bort/lägga till disken i Longhorn |
| Longhorn synk-lock | Vänta 30-60 sekunder, använd manuell kubectl patch |

### Verifieringskommandon

```bash
# Talos volume status (UUID, filesystem, phase)
talosctl -n <NODE_IP> get volumestatus u-ssd-1 -o yaml

# Blkid för filesystem-info
talosctl -n <NODE_IP> exec -t 10 -- blkid | grep sda

# Longhorn disk-status
kubectl get node.longhorn.io/<NOD> -n longhorn-system -o jsonpath='{.status.diskStatus}' | jq .
```

## Relaterade

- [[entities/talos-linux]]
- [[concepts/flux-operator]]
- [[concepts/longhorn]]
