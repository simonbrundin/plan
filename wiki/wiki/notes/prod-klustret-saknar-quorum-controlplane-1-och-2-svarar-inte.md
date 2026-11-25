---
title: Prod-klustret saknar quorum: controlplane-1 och -2 svarar inte
type: note
tags: [kubernetes, talos, drift, projekt:plan]
created: 2026-11-25
updated: 2026-11-25
status: active
related: [entities/kubernetes, notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23]
---

Diagnos 2026-11-25 13:40 (prod-klustret, context admin@cluster1):

- 10.10.10.12 (controlplane-2) svarar inte på ping, 10.10.10.11 (controlplane-1) likaså. Bara 10.10.10.13 (controlplane-3) lever.
- Talos-API:t svarar på 10.10.10.13, men etcd svarar 'no leader' - quorum saknas eftersom 2 av 3 control planes är nere.
- Därför timeoutar kubectl mot 10.10.10.12:6443 och API-servern på controlplane-3 kan inte göra något vettigt.
- Klustret hade dessutom bara kube-system (inga CRD:er) redan 2026-11-23, så även med quorum tillbaka saknas FluxOperator. Se noten 'Klustret har byggts om - bara kube-system, Flux och appar saknas (2026-11-23)'.

Vad som ska hända när en människa kan gå fram:

1. Slå på controlplane-1 (10.10.10.11) och controlplane-2 (10.10.10.12) - fysiskt, eller via PoE-cykling i UniFi (port 11 och 12 enligt infrastructure/talos/nodes.yaml). 1Password-kommandorads 'op read' gick inte att använda i agent-sessionen (kontot var inte inloggat), så UniFi-nyckeln kunde inte hämtas.
2. Vänta in etcd (talosctl -n 10.10.10.13 etcd members ska visa tre medlemmar och en leader).
3. Bootstrap'a FluxOperator (infrastructure-flux/components/flux/flux-instance.yaml) - annars synkar inget.
4. Flux kör infrastructure-flux, och apps/plan/flux-kustomization.yaml (path environments/prod) drar in plan-prod.
5. Kör environments/prod/secrets/plan-api-key.sql i Plan-databasen. Hashen i SQL-filen är verifierad: den är sha256 av PLAN_API_KEY i environments/prod/secrets/agent-api-key.enc.yaml, och Plan hashar API-nycklar med hex-sha256.
6. Verifiera agentens deployment, databas och worker, och kör ett riktigt kodmål.

Kontrollerat utan klustret: prod-manifesten renderar (25 resurser, inklusive agentens Deployment, CNPG-cluster, RBAC och hemligheter), agent-imagen sha-d213719 är mergad, och prod-databasen är CNPG (plan-db-rw.plan.svc.cluster.local) - inte Neon - så API-nyckeln kan inte sättas in förrän klustret är igång.

## Relaterade

- [[notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23]]
- [[entities/kubernetes]]
