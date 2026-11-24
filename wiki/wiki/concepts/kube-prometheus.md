---
title: Kube Prometheus Stack
type: concept
summary: Kube Prometheus Stack är en komplett observabilitetsstack för Kubernetes baserad på Prometheus, Alertmanager och Grafana.
tags: [prometheus, monitoring, alerting, grafana, observability]
sources: [infrastructure-flux/components/kube-prometheus-stack/]
created: 2026-06-02
updated: 2026-10-24
---

# Kube Prometheus Stack

## Definition

Kube Prometheus Stack är en komplett observabilitetsstack för Kubernetes baserad på Prometheus, Alertmanager och Grafana. Den samlar in metrics, genererar alerts och visualiserar data via dashboards.

## Nyckelpunkter

### Installation via Flux

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: kube-prometheus-stack
  namespace: kube-prometheus-stack
spec:
  chart:
    spec:
      chart: kube-prometheus-stack
      version: "61.0.0"
      sourceRef:
        kind: HelmRepository
        name: prometheus-community
  install:
    createNamespace: true
    timeout: 15m
  upgrade:
    timeout: 15m
```

### HelmRepository

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: prometheus-community
  namespace: flux-system
spec:
  interval: 1h
  url: https://prometheus-community.github.io/helm-charts
```

### Komponenter

| Komponent | Namespace | Beskrivning |
|-----------|-----------|-------------|
| prometheus | kube-prometheus-stack | Prometheus server |
| alertmanager | kube-prometheus-stack | Alert management |
| grafana | kube-prometheus-stack | Dashboards & visualisering |
| prometheus-operator | kube-prometheus-stack | CRD-baserad konfiguration |
| node-exporter | kube-prometheus-stack | Nod-metrics |
| kube-state-metrics | kube-prometheus-stack | Kubernetes state-metrics |

### CRD-baserad konfiguration

Istället för att konfigurera Prometheus direkt används Kubernetes Custom Resources:

```yaml
# Prometheus configuration
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: k8s
  namespace: kube-prometheus-stack
spec:
  serviceAccountName: prometheus-k8s
  serviceMonitorSelector:
    matchLabels:
      team: k8s

# ServiceMonitor för att scrape:a tjänster
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-app
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: my-app
  endpoints:
    - port: metrics
```

### HTTP Routes

Gateway API-rutter för åtkomst:

```yaml
# Grafana åtkomst via Envoy Gateway
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: grafana
  namespace: kube-prometheus-stack
spec:
  parentRefs:
    - name: envoy
  hostnames:
    - grafana.simonbrundin.com
  rules:
    - backendRefs:
        - name: kube-prometheus-stack-grafana
          port: 443
```

### Access

| Tjänst | URL | Namespace |
|--------|-----|----------|
| Grafana | https://grafana.simonbrundin.com | kube-prometheus-stack |
| Prometheus | Via Grafana eller internt | kube-prometheus-stack |
| Alertmanager | Via Grafana eller internt | kube-prometheus-stack |

### Metrics som samlas in

| Kategori | Metrics |
|----------|---------|
| **Kubernetes** | pod_cpu_usage, pod_memory_usage, deployment_status |
| **Nodes** | node_cpu_seconds, node_memory_MemAvailable, node_filesystem_* |
| **Etcd** | etcd_server_has_leader, etcd_mvcc_db_total_size |
| **API Server** | apiserver_request_total, apiserver_request_duration_seconds |
| **Scheduler** | scheduler_schedule_attempts_total |
| **Ingress** | nginx_ingress_controller_requests |

### Alert-regler

Vanliga alerts som genereras:

| Alert | Beskrivning |
|-------|-------------|
| `KubePodNotReady` | Pod inte ready efter 15 min |
| `KubeDeploymentReplicasMismatch` | Deployment har fel antal replicas |
| `KubeNodeNotReady` | Nod inte ready |
| `KubeCPUOvercommit` | CPU overcommit över viss gräns |
| `KubePersistentVolumeUsage` | PV nästan full |

### Kommandon för felsökning

```bash
# Lista Prometheus-instanser
kubectl get prometheus -n kube-prometheus-stack

# Visa Prometheus targets
kubectl get servicemonitor -A

# Lista alerts
kubectl get prometheusrule -n kube-prometheus-stack

# Visa Alertmanager config
kubectl get alertmanager -n kube-prometheus-stack -o yaml

# Logs från Prometheus
kubectl logs -n kube-prometheus-stack -l app=prometheus prometheus

# Logs från Grafana
kubectl logs -n kube-prometheus-stack -l app.kubernetes.io/name=grafana
```

### Åtkomst till Prometheus API

```bash
# Port-forward för att nå Prometheus
kubectl port-forward -n kube-prometheus-stack svc/prometheus-k8s 9090

# Query Prometheus API
curl -s http://localhost:9090/api/v1/query?query=up
```

## Kopplingar

- [[entities/kubernetes]]
- [[concepts/envoy-gateway]]
