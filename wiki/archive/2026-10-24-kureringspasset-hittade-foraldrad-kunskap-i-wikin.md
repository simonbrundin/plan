---
title: Kureringspasset hittade foraldrad kunskap i wikin
type: note
tags: [wiki, kurering, karpathy, projekt:agents]
---

Under omsättningen till Karpathys tre-lagers-modell granskades varje sida mot den faktiska infrastrukturen i ~/repos/infrastructure. Fynden: (1) ArgoCD, Authentik, Vault, Headlamp, Teleport och gamla Argo Workflows fanns kvar i entities/kubernetes trots att klustret nu drivs av FluxOperator och Zitadel; (2) cert-manager-sidan angav en ClusterIssuer 'letsencrypt-prod' som inte finns - den heter cloudflare-clusterissuer; (3) cloudflared-komponenten ar utkommenterad i components/kustomization.yaml, sa sidan markerades deprecated; (4) Longhorn-versionen var 1.10.1 men sta pa 1.12.1; (5) pages hade duplicerade sektioner (t.ex. 'Kopplingar' två ganger i concepts/flux-operator) och avklippta fragment (' resurser', ' sidor') som fanns sedan tidigare och nu ar borttagna; (6) Authentik-sidan ar kvar men markerad deprecated med en pekning till Zitadel. FOR REGELN: en sida ska inte bara existera - den ska stamma mot koden. Lint hittar inte foraldrad innehall, bara staleness; det gatt att granska manuellt. Wiki-AGENTS.md lades pa kravet 'foraldrad kunskap uppdateras, inte lamnas liggande'.
