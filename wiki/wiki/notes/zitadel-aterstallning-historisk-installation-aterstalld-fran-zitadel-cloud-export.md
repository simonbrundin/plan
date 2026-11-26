---
title: ZITADEL-återställning: historisk installation återställd från ZITADEL Cloud-export
type: note
tags: [zitadel, oidc, incident, migrering, projekt:simon]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [concepts/zitadel]
---

Den self-hosade ZITADEL-instansen på auth.simonbrundin.com hade en tom databas (ny installation 2026-11-26) och saknade den historiska installationen: organisationen simon-brundin-ab, projekten (ZITADEL, talos-oidc, kubernetes-oidc, Foodflow, Plan), användaren simonbrundin@gmail.com och alla OIDC/API-appar.

Varken Longhorn-snapshots, verifierade dumps eller någon lokal kopia av den gamla volymen fanns kvar (volymen skapades 2026-11-26, inga RecurringJobs/snapshots i klustret, worker-1/worker-2 offline och oinspekterbara). Den historiska installationen hittades i stället i ZITADEL Cloud (organisation 346062090475233657) och återställdes med ZITADELs officiella export/import: POST /admin/v1/export (med with_passwords: true) och POST /admin/v1/import. Importen bevarar organisation, projekt, application-ID:n, roller, grants och domän - men INTE det fullständiga event streamet/audit trailen, global policies, instance members, globala IDP:er, machine keys eller PAT:er.

Varför detta sätt: en verifierad export är bättre än att återkoppla okända Longhorn-filer, och ZITADEL:s eget migreringsflöde bevarar resurskopplingarna mellan organisation, projekt, användare och appar.

Lösningen gav: fungerande inloggning med Cloud-lösenordet, OIDC-login som fungerar för Flux, Dashboard, OpenClaw, Plan och Kubernetes (kubectl --context oidc-cluster1 auth whoami => zitadel-346062090475757945).

## Relaterade

- [[concepts/zitadel]]

