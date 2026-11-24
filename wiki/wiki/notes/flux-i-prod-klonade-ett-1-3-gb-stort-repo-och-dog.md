---
title: Flux i prod klonade ett 1,3 GB stort repo och dog
type: note
tags: [flux, prod, incident, projekt:infrastructure]
created: 2026-11-24
updated: 2026-11-24
status: active
---

Symptom: flux-system-kustomizationen i prod stod kvar i 'Source artifact not found' i 20 timmar och source-controller kraschade i en omstartningsloop.

Rotorsak: infrastrukturrepot är 1,3 GB (1,2 GB i .git), framför allt pxe/Bootloaders och pxe/matchbox/assets. En kloning tar cirka 1 min 48 sekunder i klustret. Flux klona med go-git, som packar upp hela repot i minnet. Controlleren har 1 GB minnesgräns i profilen 'medium', så den blev OOM-killed (exit 137) mitt i kloningen. Med 60 sekunders timeout i GitRepository.spec.timeout hann kloningen aldrig bli klar, och varje omstart av controllern började om från början.

Fix i prod: jag lack in patchar i FluxInstance.spec.kustomize.patches, eftersom flux-operator äger både Deployment och GitRepository och återställer manuella patchar:
- Deployment source-controller: livenessProbe/readinessProbe timeoutSeconds 30 och failureThreshold 10 (profilens 1 sekund räcker inte när en kloning pågår)
- Deployment source-controller: minnesgräns 4Gi
- GitRepository flux-system: spec.timeout 10m

Rätt långsiktig fix är att dela infrastrukturrepot: ha bara infrastructure-flux/ i det repo som Flux synkar, och flytta pxe/, kvm/ och agent-os/ till ett eget repo. Då blir kloningen sekunder i stället för minuter.
