---
title: Talos certifikat: kortlivade servercertifikat roteras automatiskt
type: note
tags: [talos, certifikat, dokumentation, projekt:infrastructure]
---

Talos v1.14-dokumentationen verifierades. Talos API-CA och Kubernetes API-CA är långlivade och ska inte roteras rutinmässigt; talosctl rotate-ca används endast vid komprometterad nyckel, återkallning eller CA-livstidens slut. Talos API-servercertifikat är kortlivade men roteras automatiskt av Talos. Ett giltigt certifikat med mindre än 48 timmar kvar är normalt; utgånget certifikat eller utebliven TLS-handshake är fel. Kubeletcertifikat behöver årlig nodomstart/uppgradering. Uppdaterade simon-cli health.sh commit aa090c1 så kortlivade giltiga Talos-certifikat blir information i stället för varning och tog bort den felaktiga rekommendationen om machine.ca.aggregatedWorkerRole. Uppdaterade infrastructure README commit 8be0701e med hållbar rutin.
