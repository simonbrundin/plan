---
title: Worker-7:s kubeletcertifikat signerades av en gammal Kubernetes-CA som finns kvar i git-historiken
type: note
tags: [pki, kubernetes, talos, worker, git-historia, projekt:infrastructure]
---

Worker-7 (10.10.10.27) svarar på port 10250 med ett kubelet-serving-certifikat (CN=system:node:worker-7, O=system:nodes, EKU serverAuth, SAN worker-7 + 10.10.10.27), giltigt 2026-08-24 till 2027-08-24.

Varken den aktuella kluster-CA:n (2026-10-22) eller ~/.kube/config:s CA (2026-10-14) signerar det. Leafets AuthorityKeyIdentifier är 75C44F22CD0A1D1FFD2D8A78A90DB3D20B25BA06.

Loopen: git-historikens talos/kubeconfig (blob 7d89000b, borttagen i commit 62369512 'removed') innehåller en Kubernetes-CA med exakt den SKID:n. openssl verify mot den rekonstruerade CA:n (/tmp/rec-k8s-ca.crt) ger OK, och ECDSA-verifiering av signaturen ger korrekt resultat - detta är ett verkligt signeringsbevis, inte en TypeError.

Den signerande nyckeln finns kvar i git-historiken: talos/secrets.yaml i commit 1c5ed2f3 (2025-08-28) innehåller certs.k8s.key vars SKID matchar leafets AKID. Den k8s-CAn är alltså signerare av worker-7:s kubeletleaf.

Varför: worker-7:s kubelet har en kubelet-serving-certifikat utfärdat av en CA från den ursprungliga cluster1-generationen (2025-01-20), inte av dagens CA (2026-10-22). Noden installerades alltså mot en äldre cluster-PKI. Konsekvens: kubelet på worker-7 kan inte autentisera mot dagens apiserver (HTTP 401 Unauthorized), och noden är inte medlem i nuvarande cluster1. Kubeletcertifikatet lever kvar så länge noden inte wipeas och Talos-PKI byts ut.

Viktig metodregel: jämför aldrig machine.ca.crt (O=talos) mot Kubernetes-CA (O=kubernetes) - de är separata hierarkier. Använd openssl verify eller ECDSA-verifiering, inte naiv pem-ismatch.

Följd: worker-3 till worker-8 behöver full Talos-PKI-wipe och återanslutning mot aktuell cluster1 via pxe/recover-worker.sh, vilket i sin tur kräver att PXE-värden (10.10.10.2, switch port 8) fungerar igen - den är sedan 2026-10-26 14:56 nere med noll trafik.
