---
title: Flux source-controller flyttad från worker-3 efter livenessfel
type: note
tags: [flux, incident, cluster1, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Resources blev först grönt efter att cert-manager-webhook flyttades till worker-3. Därefter föll source-controller på worker-3 i CrashLoopBackOff och Flux kunde inte hämta artifacts (connection refused). Source-controller patchades live till worker-7 men kunde inte schemaläggas eftersom noden är SchedulingDisabled, och flyttades sedan till worker-4 där image-pull pågår. Kyverno admission är fortsatt instabil; resource-validating-webhookens failurePolicy står tillfälligt Ignore för att undvika att klustrets reconciliation blockeras. Återställ dessa live-patchar deklarativt när noderna är stabila.
