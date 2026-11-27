---
title: Sex faultade Longhorn-volymer sedan Longhorn-reinstallen 2026-11-26
type: note
tags: [longhorn, incident, data, pvc, arr-stack, foodflow, projekt:infrastructure]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/longhorn]
---

Efter att Longhorn installerades om 2026-11-26 skapades flera nya Longhorn-volymer for PVC:er som redan fanns. Sex av dem har aldrig blivit friska:

- arr-stack: pvc-3c36875a, pvc-5c2e462f, pvc-e0fc38c6 (1 GiB var)
- dispatcharr: pvc-5b998e8f (10 GiB)
- foodflow: pvc-a5084d41 (1 GiB)
- arr-stack: pvc-f5d8ebfb (50 GiB)

Alla ar state=detached och robustness=faulted, och deras Replica-CR:er pekar antingen pa worker-9 (som inte ar i klustret) eller har aldrig byggts. Eftersom volymerna ar detached bygger Longhorn inga nya replikor, och apparna som anvander PVC:erna fastnar i ContainerCreating (radarr, sonarr, qbittorrent) eller CrashLoopBackOff.

Detta var INTE en foljd av USB-SSD-utbytet: fyra av volymerna var redan faultade for dagens arbete, och den sjatte (foodflow) pekar pa worker-4:s standarddisk /var/lib/longhorn/, inte pa den USB-SSD som wipesades. Datan i dem ar med all sannolikhet redan forlorad fran den tidigare klusterincarnationen; volymerna maste skapas om for att apparna ska komma igang.

Satt: arendet ar en separat datanSanering, inte en diskgrej. Det bor inte blandas ihop med overallokeringsproblemet pa worker-1.

## Relaterade

- [[concepts/longhorn]]

