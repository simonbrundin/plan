---
title: Talos etcd: en gammal EPHEMERAL-kopia pa controlplanet ger en trasig raft-log
type: note
tags: [etcd, talos, incident, controlplane, projekt:infrastructure]
created: 2026-10-25
updated: 2026-10-25
status: active
---

Symtom: controlplane-3 crash-loopar etcd med 'task "etcd" failed: exit code 2' och loggen slutar i 'panic: tocommit(1556432) is out of range [lastIndex(1556430)]. Was the raft log corrupted, truncated, or lost?'. Samtidigt visar Kubernetes noden som Ready - Kubernetes Ready ar alltsa inte tillrackligt som verifiering for en controlplane.

Root cause: ETCD-volumeen ar en katalog under EPHEMERAL (/dev/sda6). Innehallet var en GAMMAL kopia: loggen sa 'restore consistentIndex 1556430' medan klustret lag pa ~1800000 raft-index. WAL-filen saknade de entries som skulle appliceras, sa etcd panikade direkt. Orsaken var samma handelse som hostname-incidenten: noden hade bytt namn till talos-myp-y75 och lamnat en etablerad etcd-medlem med det namnet (ID a2bc299691f726bc) i klustret, medan lokal data-dir pekade pa en medlem som inte längre fanns.

Fix (provad och verifierad):
1. Ta bort den stale medlemmen fran en frisk nod: 'talosctl etcd remove-member <ID> -n 10.10.10.11'. Kvar blir 2 medlemmar - quorum finns (2 av 2 kvarvarande).
2. Tomma nodens etcd-data genom att torka EPHEMERAL: 'talosctl reset --system-labels-to-wipe EPHEMERAL --graceful=false --reboot -n 10.10.10.13'. EPHEMERAL rymmer ETCD-katalogen; STATE (config) och META (boot-id) ligger pa egna partitioner och overlever.
3. Efter reboot ansluter noden sig automatisk som ny learner och promotas till full member.

Resultat: ny member-ID 199a9f61a6dc6e1b med hostname controlplane-3, 'talosctl health' OK pa alla punkter, alla tre controlplanes Ready i Kubernetes med identisk raft-index (~1803600).

Viktiga lardomar:
- 'talosctl etcd leave' fungerar INTE pa en nod med trasig etcd: LeaveCluster behover den lokala member-ID via etcd-tjänsten, som ar trasig, och timeoutar med 'failed to get local etcd member ID: context deadline exceeded'. Anvanda 'etcd remove-member' pa en frisk nod i stallet.
- En gammal EPHEMERAL-kopia ar en akta riskfaktor: controlplane-3 hade legat med en for gammal disk-bild medan resten av klustret fortsatte. EPHEMERAL ska wiped om en controlplane lamnar klustret en tid.
- Verifiera en controlplane med 'talosctl health' + 'talosctl etcd status', inte bara kubectl get nodes.
- En medlem i etcd-klustret behaller sitt gamla hostname tills den tas bort och ansluter om; hostnameConfig.auto i en generisk config ar en aktuell riskfaktor.
