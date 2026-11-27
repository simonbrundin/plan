---
title: worker-1:s ssd-1 har 200 GiB schemalagt pa 119 GiB - och de tvA volymerna ar LEVANDE media
type: note
tags: [longhorn, worker-1, kapacitet, overprovisioning, projekt:simon]
---

Kartlagt 2026-11-27 pa worker-1 (10.10.10.21), Longhorn 1.12.1.

Bilden: ssd-1 (dedikerad USB-SSD, XFS pa /var/mnt/ssd-1) har storageMaximum 127 968 567 296 (119,2 GiB) och storageScheduled 214 748 364 800 (200 GiB). De 200 GiB ar tvA repliker om 100 GiB vardera - de enda replikorna av volymerna pvc-63da4c11 (immich/immich-library) och pvc-7bd63817 (arr-stack/arr-stack-media). BADA volymerna ar Bound PVC:er med konsumenter och ar degraded (1 av 3 replikor) med Scheduled=False 'disks are unavailable; insufficient storage'.

Viktig korrigering av en tidigare hypotes: de tvA 100 GiB-volymerna ar INTE doda. En filterning pa PVC-namn (pvc-63da4c11...) missade dem, for namnen ar immich-library och arr-stack-media. De ar levande mediadata och far inte tas bort.

Varfor overallokeringen inte star ut: Longhorns overProvisionLimit = (StorageMaximum - StorageReserved) * overProvisioningPercentage/100 = 127 968 567 296 * 200 % = 255 937 134 592 (238,4 GiB). 200 GiB var redan schemalagt, sa cirka 38,4 GiB logiskt utrymme fanns kvar - det hade kunnat lagt en ny PVC pa ssd-1 och atit den fysiska marginal (112,7 GiB ledigt) som de tvA levande replikorna behovde nar de skriver.

Beslut (taget i Flux-manifestet, reversible): ssd-1 har allowScheduling: false. Det stoppar nya replikor och rebuild-mal pa disken utan att rora de replikor som redan kor. Nodens egen Schedulable-condition forblir True - den beror pa node.Spec.AllowScheduling, inte pa enskilda diskar. Diskenns Schedulable-condition speglar halsa och kapacitet, inte spec.allowScheduling; det ar schemalagaren (replica_scheduler.go: 'if !diskSpec.AllowScheduling continue') som exkluderar disken.

Konsekvens och pris: volymerna forblir degraded 1/3 tills det finns kapacitet i klustret - och det finns inte (worker-3 12,1 GiB, worker-5 19,3 GiB, worker-9 42,1 GiB ledigt, och overprovisioningen pa dem ar redan uppskruvad). Den verkliga fixen ar kapacitet, inte en ny disk pa worker-1. ssd-1 kan inte heller anvandas som rebuild-mal for sina egna replikor, sa en diskforlust pa USB-SSD:n innebar tap av den enda kopian.
