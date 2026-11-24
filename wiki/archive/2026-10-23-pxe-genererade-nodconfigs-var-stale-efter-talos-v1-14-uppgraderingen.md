---
title: PXE-genererade nodconfigs var stale efter Talos v1.14-uppgraderingen
type: note
tags: [bug, pxe, talos, v1.14, projekt:agents]
---

## Vad som hände

Klustret uppgraderades till Talos v1.14 den 22 oktober, och då **rotades
klustrets CA**. De nodspecifika Talos-configerna under
`pxe/matchbox/assets/talos-config/nodes/` (och `worker.yaml` /
`controlplane.yaml`) genererades sist den 14 oktober och innehöll alltså den
**gamla PKI:n**.

Effekt: en nod som bootar via PXE får en config med fel CA och kan inte joina
klustret korrekt - den installerar med gammal PKI.

## Varför det hände

Filerna är gitignorade (innehåller hemligheter) och regenereras manuellt med
`./pxe/generate-node-configs.py`. Uppgraderingen till v1.14 ändrade PKI:t men
generatorn kördes inte efteråt. Dessutom kunde generatorn inte köra alls
efter v1.14: den parsade `talosctl get machineconfig -o yaml` med
`yaml.safe_load()`, vilket kastar när outputen innehåller flera dokument.

## Vad som gäller framöver

- Efter en uppgradering som roterar PKI: kör `./pxe/generate-node-configs.py`
  och **synka resultatet till PXE-värden** (`/home/simon/PXE`, som
  docker-compose mountar) - filerna i repot räcker inte, PXE-servern serverar
  sin egen kopia.
- Kontrollera att den genererade configen har samma CA som den körande
  controlplanet innan en nod installeras om:
  `talosctl get machineconfig -n <cp> -o yaml | yq -r 'select(documentIndex==0) | .spec' | yq eval-all 'select(.kind == null) | .machine.ca.crt'`
- Validera alltid: `talosctl validate --config <fil> --mode metal`.

Relaterat: [[entities/Talos Linux]], [[notes/talos-v1-14-machineconfig-specen-ar-en-quoted-string-och-flera-dokument]]
