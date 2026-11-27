---
title: Dedikerad USB-SSD per worker: validerat flöde for worker-3 till worker-8
type: note
tags: [longhorn, talos, flux, worker, disk, incident, projekt:infrastructure]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/longhorn, entities/talos-linux, notes/flux-ager-coredns-configmap-och-longhorn-node-cr-ownership-reglerna-som-gor-det-saker]
---

Malet var att ge varje worker en dedikerad USB-SSD till Longhorn och att aterstalla de två 100 GiB-mediavolymerna (immich/immich-library och arr-stack/arr-stack-media), som var degraderade for att replikorna saknade fungerande diskar.

Varje steg misslyckades pa ett annat satt och varje fel larde en sak:

1. USB-bryggans by-id kolliderar. `usb-Shenzhen_YongzhenWeiye_DD564198338A1-0:0` fanns pa worker-3, worker-5 och worker-6 samtidigt. By-id fran en USB-brygga ar alltsa inte nodunikt - ATA-serialen ar det. Ratt regel: SATA/SSD via USB-brygga identifieras pa `ata-<MODELL>_<SERIAL>`.

2. Ordningen maste vara: ta bort UserVolumeConfig -> wipe -> applicera volymen igen. Talos lockar enheten medan volymen finns i maskinconfigen, annars: "blockdevice sda is in use by volume u-ssd-1".

3. `volumeType: disk` gor att hela enheten anvands som XFS. Gamla GPT-tabeller (t.ex. fran foregaeende Longhorn-installation) ger "filesystem type mismatch: gpt != xfs". Losningen var alltid wipe med `--drop-partition`, aldrig byte till `partition`.

4. Ingen levande Replica-CR pekade pa de gamla diskarnas UUID:n (9dc77a43, f487e0cc, 285eb7ba, 9a7e5553, e18a32e4) - de var orphanade kvarvarande Longhorn-XFS:er. Wipe var darfor saker.

5. Longhorn-noderna registreras deklarativt via Flux: `infrastructure-flux/components/longhorn-system/resources/node-<nod>.yaml`, med BADE standarddisken och ssd-1. Standarddisken maste finnas i spec, annars skriver Longhorns manager `syncDefaultDisks` over kartan och Flux-agar areten bryts.

6. `allowScheduling` valdes efter fysisk storlek: true for >=120 GB (en 100 GiB-replika ryms inom 200 % over-provisioning), false for de 62 GB-stickarna. Longhorn raknar pa logiskt `storageScheduled` mot fysiskt `storageMaximum` och stoppar inte en 100 GiB-replika pa en 60 GiB-disk.

7. Flux-steget `resources` behovde Longhorn i bootstrap health checks, annars kan inte Node-CR:er server-side dry-runas innan CRD:n finns.

Resultat: worker-3, -4, -5, -6 har nu riktiga SSD-kapaciteter (112/238/112/447 GiB), worker-7 och -8 har stickar med allowScheduling false, och de två 100 GiB-volymerna ar healthy 3/3 pa worker-1, worker-3 och worker-5. Total Longhorn-kapacitet 1502 GiB, varav 835,9 GiB schemalagt.

Kvar: worker-1:s ssd-1 har fortfarande 200 GiB logiskt schemalagt pa en 119 GiB-disk. Den satts till allowScheduling false for att stoppa nya replikor, men de tre 100 GiB-replikorna maste balanseras bort eller byggas om pa worker-4/worker-6.

## Relaterade

- [[entities/talos-linux]]
- [[notes/flux-ager-coredns-configmap-och-longhorn-node-cr-ownership-reglerna-som-gor-det-saker]]
- [[concepts/longhorn]]
