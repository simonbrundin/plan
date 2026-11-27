---
title: Longhorns manager-DaemonSet saknar tolerations - en tainted worker blir aldrig en Longhorn-nod
type: note
tags: [longhorn, worker-8, taint, daemonset, felsokning, projekt:infrastructure]
---

Worker-8 var Kubernetes Ready och Talos-disken var korrekt monterad pa /var/mnt/ssd-1, men Longhorn sag "Manager pod is missing: node worker-8 has no manager pod running on it" och status.diskStatus var null.

Rotorsak var inte disken utan en NoSchedule-taint: node.kubernetes.io/broken=2026-11-23. Longhorns manager DaemonSet (templates/daemonset-sa.yaml i longhorn-charten 1.12.1) har INGA tolerations, sa en tainted nod far aldrig en manager-pod. DaemonSetens desiredNumberScheduled var 7 av 8 noder - den tainted noden raknades inte ens in.

Tainten var stale: noden hade varit kontinuerligt Ready sedan 2026-11-25T23:57 utan resource pressure. Efter att tainten tagits bort startade manager-pod, instance-manager och CSI-plugin, och worker-8 blev en fullt fungerande Longhorn-nod med Ready=True och bade diskarna i status.diskStatus.

Lardom: innan man gor ingenting med diskar pa en nod som ar Ready=False i Longhorn ska man kolla:
- kubectl get node <nod> -o jsonpath='{.spec.taints}'
- Longhorns manager-pod pa noden (kubectl get pods -n longhorn-system -o wide | grep <nod>)

En nod utan manager-pod ar inte en disk-problem utan en scheduler-problem. Notera att samma taint-blockering galler for alla DaemonSets utan tolerations, inte bara Longhorn.

Samma taint fanns pa controlplane-2, men dar har noden redan control-plane-NoSchedule-tainten och Longhorn-kapacitet inte ska placeras.
