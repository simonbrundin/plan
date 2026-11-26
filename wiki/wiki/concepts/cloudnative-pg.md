---
title: CloudNativePG (CNPG)
type: concept
summary: CloudNativePG är en Kubernetes-native PostgreSQL-operator som tillhandahåller automatiserad hantering av PostgreSQL-databaser.
tags: [postgresql, database, postgres, cloudnativepg, cnpg]
sources: [infrastructure-flux/components/cloudnative-pg/]
created: 2026-06-02
updated: 2026-11-26
related: [concepts/zitadel, notes/plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor]
---

# CloudNativePG (CNPG)

## Definition

CloudNativePG är en Kubernetes-native PostgreSQL-operator som tillhandahåller automatiserad hantering av PostgreSQL-databaser. Den ersätter traditionella metoder för att köra databaser i Kubernetes genom att erbjuda high availability, backups och automatisk failover.

## Nyckelpunkter

### Installation via Flux

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: cloudnative-pg
  namespace: cloudnative-pg
spec:
  chart:
    spec:
      chart: cloudnative-pg
      version: 0.26.0
      sourceRef:
        kind: HelmRepository
        name: cloudnative-pg
  install:
    timeout: 15m
  upgrade:
    timeout: 15m
```

### HelmRepository

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: cloudnative-pg
  namespace: flux-system
spec:
  interval: 1h
  url: https://cloudnative-pg.io/charts/
```

### Resurser

| Fil | Typ | Beskrivning |
|-----|-----|-------------|
| namespace.yaml | Namespace | cloudnative-pg namespace |
| source.yaml | HelmRepository | CNPG chart repo |
| release.yaml | HelmRelease | Helm release |
| mutating-webhook.yaml | MutatingWebhookConfiguration | Pod webhook |
| validating-webhook.yaml | ValidatingWebhookConfiguration | Validation webhook |
| patch-crd-poolers.yaml | Kustomize patch | Anpassning av Pooler CRDs |

### Cluster-resurs

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: my-db
  namespace: databases
spec:
  instances: 3
  primaryUpdateStrategy: unsupervised
  storage:
    size: 10Gi
    storageClass: longhorn
  backups:
    retentionPolicy: "7d"
    barmanObjectStore:
      destinationPath: s3://my-bucket/backups
      s3Credentials:
        accessKeyId:
          name: s3-creds
          key: access-key
        secretAccessKey:
          name: s3-creds
          key: secret-key
```

### High Availability

CloudNativePG använder en Patroni-baserad arkitektur:

| Komponent | Ansvar |
|-----------|--------|
| **Primary** | Skriver och läser |
| **Standby replicas** | Läs-scaling, failover-kandidater |
| **Patroni** | Automatisk failover vid primary-fel |
| **Etcd/raft** | Konsensus |

### Automatic Failover

```
1. Primary blir otillgänglig
       ↓
2. Patroni på standby-replicas märker förlust av primary
       ↓
3. En standby promoted till new primary
       ↓
4. Andra standbys pekar om till new primary
       ↓
5. Applikationer reconnectar via Cluster DNS
```

### Backup-strategi

Barman Object Store för backup:

```yaml
spec:
  backups:
    retentionPolicy: "7d"  # Behåll 7 dagars backup
    barmanObjectStore:
      destinationPath: s3://bucket/backups
      walStorageClass: longhorn
```

### Connection Pooling med PgBouncer

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Pooler
metadata:
  name: my-db-pooler
  namespace: databases
spec:
  cluster:
    name: my-db
  instances: 2
  poolMode: transaction
  pgbouncer:
    poolSize: 20
```

### Kommandon för felsökning

```bash
# Lista databaskluster
kubectl get cluster -A

# Visa kluster-status
kubectl get cluster <name> -n <namespace> -o yaml | jq '.status'

# Lista pods i kluster
kubectl get pods -n <namespace> -l cnpg.io/cluster=<name>

# Visa WAL-replication status
kubectl exec -n <namespace> <primary-pod> -- psql -c "SELECT * FROM pg_stat_replication;"

# Manuell failover
kubectl cnp failover -n <namespace> <cluster-name>

# Restore från backup
kubectl cnp backup <cluster-name> -n <namespace>
```

### Vanliga problem

| Problem | Orsak | Lösning |
|---------|-------|---------|
| Pod stuck i init | S3-credentials fel | Verifiera secret och bucket-åtkomst |
| Replication lag | Nätverksproblem eller tung belastning | Kolla `pg_stat_replication` |
| `FencingActive` | Failsafe aktiverad | Kontrollera pod-status |
| Backup misslyckad | BUCKET/path saknas | Verifiera S3-konfiguration |

## Kopplingar

- [[entities/kubernetes]]
- [[concepts/longhorn]]
- [[concepts/zitadel]] (använder CNPG för `zitadel-db`)

## Relaterade

- [[notes/plan-db-cnpg-kluster-fastnade-i-setting-up-primary-pa-en-longhorn-volym-utan-replikor]]

