---
title: Talos image-cache kräver EPHEMERAL -> cirkulärt beroende med STATE/META på en färsk metal-nod
type: note
tags: [pxe, talos, image-cache, state, meta, worker, projekt:infrastructure]
---

En oinstallerad metal-nod (t.ex. ROCK med tom eMMC) kan inte installera sig
sjalv. Kedjan:
  installeraren vill pulla metall-installer-imagen
    -> containerns image-cache ligger pa EPHEMERAL
    -> EPHEMERAL ligger i STATE-partitionen
    -> STATE kräver system_disk
    -> system_disk härleds ur META-partitionen
    -> META skapas bara av installeraren
  = cirkel. Symptom: pullen startar men avbryts med "lease does not exist:
  not found" (containerd-internt fel), och STATE-volymen failar med
  "error evaluating disk locator: no such attribute(s): system_disk".

FIX: stang av den lokala image-cachen i configen, sa att containern pullar
direkt utan EPHEMERAL:
  apiVersion: v1alpha1
  kind: ImageCacheConfig
  local:
    enabled: false
Darmast kan installeraren koras pa en farsk nod, och den skapar
EFI/META/STATE/EPHEMERAL (bootloader-steget failar for efivarfs-skallet pa
ROCK, men partitionerna blir ligga).

FOLJDFRAGA: utan STATE kan noden inte gora med i cluster1 overhuvudtaget
(kubelet och containern behover EPHEMERAL). En ROCK-nod maste alltsa ha
partitionerna, men fortfarande vara PXE-only.
