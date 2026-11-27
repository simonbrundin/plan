---
title: Overallokeringen pa worker-1/3/5 ar ett logiskt, inte ett fysiskt problem
type: note
tags: [longhorn, kapacitet, worker-1, drift, projekt:infrastructure]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/longhorn, notes/dedikerad-usb-ssd-per-worker-validerat-flode-for-worker-3-till-worker-8]
---

Bedomningen av overallokeringen pa worker-1, worker-3 och worker-5 blev enkel nar man jamfor logisk och fysisk anvandning.

Longhorn raknar pa logiskt storageScheduled (volymernas storlek: 200 GiB per nod) mot fysiskt storageMaximum (disken: 112-119 GiB). Ser det ut som overallokering. Men de faktiska replikorna innehaller bara cirka 6,4 GiB data per nod - matad direkt pa noden:

- worker-1 ssd-1: 6,5 G anvand av 120 G (113 G ledigt)
- worker-3 ssd-1: 6,4 G anvand av 112 G (106 G ledigt)
- worker-5 ssd-1: 6,4 G anvand av 112 G (106 G ledigt)
- worker-4 ssd-1: 4,6 G anvand av 239 G (234 G ledigt)
- worker-6 ssd-1: 8,6 G anvand av 447 G (439 G ledigt)

Slutsats: overallokeringen ar ett logiskt - inte ett fysiskt - problem just nu. Diskarna har 100+ GiB ledigt och de två 100 GiB-mediavolymerna (immich/immich-library och arr-stack/arr-stack-media) har vuxit med cirka 6 GiB. Klustrets totala overprovisioning ar 835,9 GiB schemalagt mot 1502 GiB max, alltsa 56 procent - inget varningstra hos Longhorn.

Risken är latent, inte aktiv: om arr-stack-media och immich-library vxer till sina fulla 100 GiB vardera kommer worker-1/3/5 att floda over, da tre repliker x 100 GiB logiskt ska ligga pa cirka 113 GiB fysiskt. Dagens skydd ar allowScheduling false pa worker-1:s ssd-1 (och pa worker-7/8:s stickar), sa inga nya replikor placeras pa overbelastade diskar.

Rekommenderad handlingsplan nar det blir relevant: byt en replika i taget genom att radera Longhorn Replica-CR:en for noden - Longhorn bygger da om en ny replika pa en nod med utrymme (worker-4 eller worker-6) och volymen ar fortfarande anvandbar under bygget (2 av 3 replikor kvar). Byggomgangen ar billig eftersom det bara ar cirka 6 GiB data. Gor det en volym i taget och vanta till robustness ar healthy igen.

Lardom: overprovisioning-varningar och storageScheduled ska alltid jamforas mot verklig df-anvandning pa noden innan man dragor slutsatser om risk. Logisk och fysisk allokering ar olika storheter.

## Relaterade

- [[notes/dedikerad-usb-ssd-per-worker-validerat-flode-for-worker-3-till-worker-8]]
- [[concepts/longhorn]]
