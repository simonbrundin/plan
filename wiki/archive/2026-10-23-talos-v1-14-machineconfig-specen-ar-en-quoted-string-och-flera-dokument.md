---
title: Talos v1.14: machineconfig-specen är en quoted string och flera dokument
type: note
tags: [bug, talos, v1.14, pxe, projekt:agents]
---

## Vad som hände

Talos v1.14 ändrade hur en MachineConfig ser ut, och det bröt flera verktyg som
läser configen:

1. `talosctl get machineconfig -o yaml` returnerar **två** MachineConfig-resurser
   (`id: persistent` och `id: v1alpha1`) med identisk `spec`. En `yaml.safe_load()`
   på outputen kastar `ComposerError: expected a single document in the stream`.
2. `spec` är en **quoted string** (`spec: "version: v1alpha1\nmachine:..."`),
   inte en literal block (`spec: |`). Verktyg som letar efter "spec: |" hittar
   ingenting.
3. `spec`-strängen innehåller **flera YAML-dokument**: ett legacy-dokument
   (`version: v1alpha1` med `machine:`/`cluster:`) plus ett dokument per config-typ
   (`UnattendedInstallConfig`, `KubeletConfig`, `KubePrismConfig`,
   `KubeAuthenticationConfig`, `KubeAuthorizerConfig` (två - `node` och `rbac`),
   ...).

## Varför det spelar roll

Varje verktyg som patchar, jämför eller genererar Talos-config måste hantera
alla tre punkterna. `simon talos update config` (i simon-cli) och
`pxe/generate-node-configs.py` (i infrastructure) gjorde inte det, så båda
failade med tomma/ogiltiga resultat.

## Vad som gäller framöver

- Extrahera configen med `yq` (mikefarah, inte jq-wrappern som också heter `yq`):
  `yq eval-all 'select(documentIndex == 0) | .spec' machineconfig.yaml`.
  `documentIndex == 0` väljer den permanenta resursen (båda har samma spec).
- En config är ett *multi-dokument*, inte ett enskilt YAML-dokument. Patcha
  med `talosctl patch machineconfig --patch @fil` (funkar med flera dokument),
  inte `apply-config` (kräver en komplett config med CA).
- `machine.install` finns inte i v1.14 - installationen ligger i
  `UnattendedInstallConfig` (`provisioning.diskSelector.match` är ett
  CEL-uttryck, t.ex. `disk.serial == "0x00002039"`). `machine.kubelet` ->
  `KubeletConfig`, `machine.features.kubePrism` -> `KubePrismConfig`,
  `cluster.apiServer` -> `KubeAPIServerConfig` (men `--oidc-*` ->
  `KubeAuthenticationConfig`), `cluster.extraManifests` ->
  `KubeExternalManifestConfig`.
- `machine.token` är **klustergemensamt** (samma token i alla noder, ligger i
  secrets-bundlen) - randomisera det inte per nod.
- `UserVolumeConfig` i v1.14 kräver `volumeType: disk` och stöder inte
  `minSize`/`maxSize`/`grow`.

Relaterat: [[entities/Talos Linux]], [[summaries/longhorn-disk-guide]]
