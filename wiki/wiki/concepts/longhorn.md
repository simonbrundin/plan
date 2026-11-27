---
title: Longhorn
type: concept
summary: Longhorn är ett distribuerat block-lagringssystem för Kubernetes.
tags: [longhorn, storage, csi, persistent-storage, distributed-storage]
sources: [infrastructure-flux/components/longhorn-system/, docs/add-disk-to-longhorn.md]
created: 2026-06-02
updated: 2026-11-27
related: [notes/flux-ager-coredns-configmap-och-longhorn-node-cr-ownership-reglerna-som-gor-det-saker, notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning, notes/worker-1-s-ssd-1-har-200-gib-schemalagt-pa-119-gib-och-de-tva-volymerna-ar-levande-media]
---

# Longhorn

## Definition

Longhorn är ett distribuerat block-lagringssystem för Kubernetes. Det använder container storage interface (CSI) för att tillhandahålla persistent storage med funktioner som replikering, snapshots och backups.

## Nyckelpunkter

### Installation via Flux

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: longhorn
  namespace: longhorn-system
spec:
  chart:
    spec:
      chart: longhorn
      version: 1.12.1
      sourceRef:
        kind: HelmRepository
        name: longhorn
  values:
    defaultSettings:
      defaultReplicaCount: 3
    preUpgradeChecker:
      jobEnabled: false
```

### HelmRepository

Longhorn-charten finns i eget repository:
```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: longhorn
  namespace: flux-system
spec:
  interval: 1h
  url: https://charts.longhorn.io
```

### StorageClass

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: longhorn
provisioner: driver.longhorn.io
allowVolumeExpansion: true
reclaimPolicy: Delete
volumeBindingMode: Immediate
parameters:
  numberOfReplicas: "3"
  staleReplicaTimeout: "2880"
  fromBackup: ""
  fsType: ext4
```

### Två-lager arkitektur

| Lager | Ansvar | Konfigurationsfil |
|-------|--------|-------------------|
| **Talos** | Formatera, montera diskar | `talos/patches/workers/<nod>/disks/*.yaml` |
| **Longhorn** | Storage management | `manifests/.../nodes.yaml` |

```
Talos UserVolumeConfig → /var/mnt/ssd-1 → Longhorn Node CRD → Replicas
```

### Konfiguration per nod

```yaml
apiVersion: longhorn.io/v1beta2
kind: Node
metadata:
  name: worker-9
  namespace: longhorn-system
spec:
  allowScheduling: true
  disks:
    ssd-1:
      path: /var/mnt/ssd-1
      allowScheduling: true
      evictionRequested: false
```

### Replikeringsstrategi

| Antal replicas | Tolerans |
|---------------|----------|
| 1 | 1 nod-fel |
| 2 | 2 nod-fel |
| 3 | 2 nod-fel, 1 disk-fel |

**Default:** 3 replicas per volym

## Användning

### Skapa PersistentVolumeClaim

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-data
spec:
  storageClassName: longhorn
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

### Backup till S3

Longhorn stödjer backup till S3-kompatibla lagringssystem:

```yaml
apiVersion: longhorn.io/v1beta2
kind: BackupTarget
metadata:
  name: s3-backup
  namespace: longhorn-system
spec:
  type: s3
  s3:
    endpoint: "s3.amazonaws.com"
    bucketName: "longhorn-backups"
    # ...
```

## Felsökning

### Kommandon

```bash
# Lista noder och deras diskar
kubectl get node.longhorn.io -n longhorn-system -o wide

# Lista volymer
kubectl get volumes.longhorn.io -n longhorn-system

# Lista replicas
kubectl get replicas.longhorn.io -n longhorn-system -o wide

# Visa node-status med disk-info
kubectl get node.longhorn.io/<nod> -n longhorn-system -o yaml | jq '.status.diskStatus'

# Ta bort trasig replica
kubectl delete replicas.longhorn.io/<replica> -n longhorn-system
```

### Vanliga problem

| Problem | Orsak | Lösning |
|---------|-------|---------|
| Volume stuck i detaching | Nod nere eller nätverksproblem | Vänta eller force delete pod |
| `default-disk-*` auto-skapad | Ingen explicit disk-konfiguration | Definiera diskar i Node CRD |
| Fel disk används | Disk-namnskonflikt | Uppdatera `longhorn-disk.cfg` |
| UUID mismatch | Talos reinstall | Ta bort och lägga till disken i Longhorn |

### Diskhanterings-checklist

- [ ] Talos UserVolumeConfig pekar på rätt disk-by-id
- [ ] UserVolumeConfig applicerad och reboot genomförd
- [ ] Talos volume status visar `phase: ready` och `filesystem: xfs`
- [ ] blkid visar TYPE="xfs" och PARTLABEL="u-<namn>"
- [ ] longhorn-disk.cfg uppdaterad med korrekt UUID
- [ ] Disken definierad i Longhorn nodes.yaml

## Kopplingar

- [[entities/talos-linux]]
- [[entities/kubernetes]]
- [[summaries/longhorn-disk-guide]]

## Relaterade

- [[notes/worker-1-s-ssd-1-har-200-gib-schemalagt-pa-119-gib-och-de-tva-volymerna-ar-levande-media]]
- [[notes/flux-ager-coredns-configmap-och-longhorn-node-cr-ownership-reglerna-som-gor-det-saker]]
- [[notes/usb-disken-acer-wd-320gb-status-och-longhorn-bedomning]]
